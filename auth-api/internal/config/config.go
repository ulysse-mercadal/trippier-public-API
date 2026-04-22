// Package config loads runtime configuration from environment variables.
package config

import (
	"strings"

	"github.com/spf13/viper"
)

// Config holds all runtime configuration for auth-api.
type Config struct {
	Port                  string
	DatabaseURL           string
	RedisURL              string
	JWTSecret             string
	InternalSecret        string
	SMTPHost              string
	SMTPPort              int
	SMTPFrom              string
	AppURL                string
	DefaultTokensLimit    int
	DefaultResetIntervalS int
}

// Load reads configuration from environment variables (prefixed AUTH_).
func Load() *Config {
	v := viper.New()
	v.SetEnvPrefix("AUTH")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	v.SetDefault("PORT", "8081")
	v.SetDefault("DATABASE_URL", "postgres://trippier:trippier@postgres:5432/trippier?sslmode=disable")
	v.SetDefault("REDIS_URL", "redis://redis:6379")
	v.SetDefault("JWT_SECRET", "change-me-in-production-32-chars!!")
	v.SetDefault("INTERNAL_SECRET", "change-me-internal-secret")
	v.SetDefault("SMTP_HOST", "mailhog")
	v.SetDefault("SMTP_PORT", 1025)
	v.SetDefault("SMTP_FROM", "noreply@trippier.dev")
	v.SetDefault("APP_URL", "http://localhost:3000")
	v.SetDefault("DEFAULT_TOKENS_LIMIT", 1000)
	v.SetDefault("DEFAULT_RESET_INTERVAL_S", 3600)

	return &Config{
		Port:                  v.GetString("PORT"),
		DatabaseURL:           v.GetString("DATABASE_URL"),
		RedisURL:              v.GetString("REDIS_URL"),
		JWTSecret:             v.GetString("JWT_SECRET"),
		InternalSecret:        v.GetString("INTERNAL_SECRET"),
		SMTPHost:              v.GetString("SMTP_HOST"),
		SMTPPort:              v.GetInt("SMTP_PORT"),
		SMTPFrom:              v.GetString("SMTP_FROM"),
		AppURL:                v.GetString("APP_URL"),
		DefaultTokensLimit:    v.GetInt("DEFAULT_TOKENS_LIMIT"),
		DefaultResetIntervalS: v.GetInt("DEFAULT_RESET_INTERVAL_S"),
	}
}
