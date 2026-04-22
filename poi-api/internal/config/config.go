// Package config loads and exposes application configuration via environment variables.
package config

import (
	"strings"

	"github.com/spf13/viper"
)

// Config holds all runtime configuration for the poi-api server.
type Config struct {
	Port             string
	RedisURL         string
	CacheTTLSeconds  int
	ProviderTimeout  int
	GeoNamesUsername string
	Lang             string
	LogLevel         string
	AuthAPIURL       string
	InternalSecret   string
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
	v.SetDefault("provider_timeout", 8)
	v.SetDefault("log_level", "info")
	v.SetDefault("lang", "en")
	v.SetDefault("auth_api_url", "http://auth-api:8081")
	v.SetDefault("internal_secret", "change-me-internal-secret")

	cfg := &Config{
		Port:             v.GetString("port"),
		RedisURL:         v.GetString("redis_url"),
		CacheTTLSeconds:  v.GetInt("cache_ttl_seconds"),
		ProviderTimeout:  v.GetInt("provider_timeout"),
		GeoNamesUsername: v.GetString("geonames_username"),
		Lang:             v.GetString("lang"),
		LogLevel:         v.GetString("log_level"),
		AuthAPIURL:       v.GetString("auth_api_url"),
		InternalSecret:   v.GetString("internal_secret"),
	}

	return cfg, nil
}
