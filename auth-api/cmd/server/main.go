// Package main starts the auth-api HTTP server.
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

	"github.com/trippier/auth-api/internal/apikey"
	"github.com/trippier/auth-api/internal/auth"
	"github.com/trippier/auth-api/internal/config"
	"github.com/trippier/auth-api/internal/db"
	"github.com/trippier/auth-api/internal/email"
	mw "github.com/trippier/auth-api/internal/middleware"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "config: %v\n", err)
		os.Exit(1)
	}

	log := buildLogger(cfg.LogLevel)
	defer log.Sync() //nolint:errcheck

	ctx := context.Background()

	pool, err := db.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatal("db", zap.Error(err))
	}
	defer pool.Close()

	opt, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		log.Fatal("redis url", zap.Error(err))
	}
	rdb := redis.NewClient(opt)
	defer rdb.Close() //nolint:errcheck

	emailer := email.New(cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPFrom)

	authSvc := auth.New(pool, emailer, cfg.JWTSecret, cfg.AppURL)
	authHandler := auth.NewHandler(authSvc, cfg.AppURL)

	keySvc := apikey.New(pool, rdb, cfg.DefaultTokensLimit, cfg.DefaultResetIntervalS, log)
	keyHandler := apikey.NewHandler(keySvc)

	jwtAuth := mw.JWTAuth(cfg.JWTSecret)
	internalAuth := mw.InternalAuth(cfg.InternalSecret)

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.SetTrustedProxies(nil) //nolint:errcheck
	r.Use(gin.Logger(), gin.Recovery(), mw.SecureHeaders())

	r.GET("/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })

	loginLimiter := mw.IPRateLimit(rdb, 10, time.Minute)
	registerLimiter := mw.IPRateLimit(rdb, 5, 15*time.Minute)
	authHandler.RegisterRoutes(r.Group(""), jwtAuth, loginLimiter, registerLimiter)
	keyHandler.RegisterRoutes(
		r.Group("/api-keys"),
		r.Group("/internal", internalAuth),
		jwtAuth,
	)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Port),
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Info("auth-api starting", zap.String("addr", srv.Addr))
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
