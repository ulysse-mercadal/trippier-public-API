// Package main is the entry point of the poi-api server.
// It wires configuration, providers, and the HTTP router together.
package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"github.com/trippier/poi-api/internal/config"
	"github.com/trippier/poi-api/internal/middleware"
	"github.com/trippier/poi-api/internal/providers"
	"github.com/trippier/poi-api/internal/providers/eventbrite"
	"github.com/trippier/poi-api/internal/providers/geonames"
	"github.com/trippier/poi-api/internal/providers/overpass"
	"github.com/trippier/poi-api/internal/providers/ticketmaster"
	"github.com/trippier/poi-api/internal/providers/wikipedia"
	"github.com/trippier/poi-api/internal/providers/wikivoyage"
	"github.com/trippier/poi-api/internal/search"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "config: %v\n", err)
		os.Exit(1)
	}

	log := buildLogger(cfg.LogLevel)
	defer log.Sync() //nolint:errcheck

	if cfg.GeoNamesUsername == "" {
		log.Info("POI_GEONAMES_USERNAME not set — geonames provider will be disabled")
	}


	rdb, err := buildRedis(cfg.RedisURL)
	if err != nil {
		log.Fatal("redis url", zap.Error(err))
	}

	pp := buildProviders(cfg)
	svc := search.NewService(pp, time.Duration(cfg.ProviderTimeout)*time.Second, log)
	handler := search.NewHandler(svc)

	globalAuth, eventsAuth := buildAuthMiddlewares(cfg)

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.SetTrustedProxies(nil) //nolint:errcheck
	r.Use(
		gin.Recovery(),
		middleware.SecureHeaders(),
		middleware.RequestID(),
		middleware.Logger(log),
		globalAuth,
	)

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	cacheTTL := time.Duration(cfg.CacheTTLSeconds) * time.Second
	pois := r.Group("/pois")
	pois.Use(middleware.Cache(rdb, cacheTTL))
	handler.RegisterRoutes(pois)

	events := pois.Group("/events")
	events.Use(eventsAuth)
	handler.RegisterEventRoutes(events)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Port),
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Info("poi-api starting", zap.String("addr", srv.Addr))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal("server error", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("shutting down server…")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Error("shutdown error", zap.Error(err))
	}
	log.Info("server stopped")
}

// buildAuthMiddlewares returns the global and events-specific rate-limit middlewares.
// When AUTH_DISABLED=true both are no-ops so the API runs without an auth-api dependency.
func buildAuthMiddlewares(cfg *config.Config) (global, events gin.HandlerFunc) {
	if cfg.AuthDisabled {
		return middleware.Passthrough(), middleware.Passthrough()
	}
	global = middleware.RateLimit(cfg.AuthAPIURL, cfg.InternalSecret, 1, "/health", "/pois/events", "/pois/events/slim")
	events = middleware.RateLimit(cfg.AuthAPIURL, cfg.InternalSecret, 10)
	return global, events
}

// buildProviders constructs the list of active POI and event providers based on config.
// GeoNames, Ticketmaster, and Eventbrite are only added when their credentials are configured.
func buildProviders(cfg *config.Config) []providers.Provider {
	pp := []providers.Provider{
		overpass.New(),
		wikivoyage.New(cfg.Lang),
		wikipedia.New(cfg.Lang),
		wikipedia.NewEventProvider(cfg.Lang),
	}
	if cfg.GeoNamesUsername != "" {
		pp = append(pp, geonames.New(cfg.GeoNamesUsername))
	}
	if cfg.TicketmasterAPIKey != "" {
		pp = append(pp, ticketmaster.New(cfg.TicketmasterAPIKey))
	}
	if cfg.EventbriteToken != "" {
		pp = append(pp, eventbrite.New(cfg.EventbriteToken))
	}
	return pp
}

// buildLogger returns a production zap logger, or a development logger when level is "debug".
func buildLogger(level string) *zap.Logger {
	var cfg zap.Config
	if level == "debug" {
		cfg = zap.NewDevelopmentConfig()
	} else {
		cfg = zap.NewProductionConfig()
	}
	log, err := cfg.Build()
	if err != nil {
		panic(fmt.Sprintf("build logger: %v", err))
	}
	return log
}

// buildRedis parses a Redis URL and returns a connected client.
func buildRedis(redisURL string) (*redis.Client, error) {
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, err
	}
	return redis.NewClient(opt), nil
}
