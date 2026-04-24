// Package main starts the auth-api HTTP server.
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
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
		log.Fatalf("config: %v", err)
	}

	ctx := context.Background()

	pool, err := db.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("db: %v", err)
	}
	defer pool.Close()

	opt, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		log.Fatalf("redis url: %v", err)
	}
	rdb := redis.NewClient(opt)
	defer rdb.Close()

	emailer := email.New(cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPFrom)

	authSvc := auth.New(pool, emailer, cfg.JWTSecret, cfg.AppURL)
	authHandler := auth.NewHandler(authSvc, cfg.AppURL)

	keySvc := apikey.New(pool, rdb, cfg.DefaultTokensLimit, cfg.DefaultResetIntervalS)
	keyHandler := apikey.NewHandler(keySvc)

	jwtAuth := mw.JWTAuth(cfg.JWTSecret)
	internalAuth := mw.InternalAuth(cfg.InternalSecret)

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.SetTrustedProxies(nil) //nolint:errcheck
	r.Use(gin.Logger(), gin.Recovery(), mw.SecureHeaders())

	r.GET("/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })

	authHandler.RegisterRoutes(r.Group(""), jwtAuth)
	keyHandler.RegisterRoutes(
		r.Group("/api-keys"),
		r.Group("/internal", internalAuth),
		jwtAuth,
	)

	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("auth-api listening on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("server: %v", err)
	}
}
