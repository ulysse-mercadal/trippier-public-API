// Package config loads and exposes application configuration via environment variables.
package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// Config holds all runtime configuration for the poi-api server.
type Config struct {
	Port             string
	RedisURL         string
	CacheTTLSeconds  int
	RateLimitPerMin  int
	ProviderTimeout  int
	GeoNamesUsername string
	Lang             string
	LogLevel         string
}

// Load reads configuration from environment variables (prefixed POI_)
// and falls back to sensible defaults.
func Load() (*Config, error) {
	v := viper.New()
	v.SetEnvPrefix("POI")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	v.SetDefault("port", "8080")
	v.SetDefault("redis_url", "redis://localhost:6379")
	v.SetDefault("cache_ttl_seconds", 3600)
	v.SetDefault("rate_limit_per_min", 100)
	v.SetDefault("provider_timeout", 8)
	v.SetDefault("log_level", "info")
	v.SetDefault("lang", "en")

	cfg := &Config{
		Port:             v.GetString("port"),
		RedisURL:         v.GetString("redis_url"),
		CacheTTLSeconds:  v.GetInt("cache_ttl_seconds"),
		RateLimitPerMin:  v.GetInt("rate_limit_per_min"),
		ProviderTimeout:  v.GetInt("provider_timeout"),
		GeoNamesUsername: v.GetString("geonames_username"),
		Lang:             v.GetString("lang"),
		LogLevel:         v.GetString("log_level"),
	}

	if cfg.GeoNamesUsername == "" {
		fmt.Println("warning: POI_GEONAMES_USERNAME not set — geonames provider will be disabled")
	}

	return cfg, nil
}
