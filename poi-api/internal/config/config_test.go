package config_test

import (
	"testing"

	"github.com/trippier/poi-api/internal/config"
)

func TestLoad_Defaults(t *testing.T) {
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if cfg.Port == "" {
		t.Error("Port should have a default value")
	}
	if cfg.RedisURL == "" {
		t.Error("RedisURL should have a default value")
	}
	if cfg.CacheTTLSeconds <= 0 {
		t.Errorf("CacheTTLSeconds = %d, want > 0", cfg.CacheTTLSeconds)
	}
	if cfg.ProviderTimeout <= 0 {
		t.Errorf("ProviderTimeout = %d, want > 0", cfg.ProviderTimeout)
	}
	if cfg.Lang == "" {
		t.Error("Lang should default to 'en'")
	}
	if cfg.LogLevel == "" {
		t.Error("LogLevel should have a default value")
	}
}

func TestLoad_DefaultPort(t *testing.T) {
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if cfg.Port != "8080" {
		t.Errorf("default Port = %q, want %q", cfg.Port, "8080")
	}
}

func TestLoad_DefaultLang(t *testing.T) {
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if cfg.Lang != "en" {
		t.Errorf("default Lang = %q, want %q", cfg.Lang, "en")
	}
}

func TestLoad_DefaultAuthDisabled(t *testing.T) {
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if cfg.AuthDisabled {
		t.Error("AuthDisabled should default to false")
	}
}

func TestLoad_EnvOverride(t *testing.T) {
	t.Setenv("POI_PORT", "9090")
	t.Setenv("POI_LANG", "fr")
	t.Setenv("POI_CACHE_TTL_SECONDS", "7200")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if cfg.Port != "9090" {
		t.Errorf("Port = %q, want %q", cfg.Port, "9090")
	}
	if cfg.Lang != "fr" {
		t.Errorf("Lang = %q, want %q", cfg.Lang, "fr")
	}
	if cfg.CacheTTLSeconds != 7200 {
		t.Errorf("CacheTTLSeconds = %d, want 7200", cfg.CacheTTLSeconds)
	}
}

func TestLoad_GeoNamesUsername(t *testing.T) {
	t.Setenv("POI_GEONAMES_USERNAME", "testuser")
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if cfg.GeoNamesUsername != "testuser" {
		t.Errorf("GeoNamesUsername = %q, want %q", cfg.GeoNamesUsername, "testuser")
	}
}

func TestLoad_AuthDisabledEnv(t *testing.T) {
	t.Setenv("POI_AUTH_DISABLED", "true")
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if !cfg.AuthDisabled {
		t.Error("AuthDisabled should be true when POI_AUTH_DISABLED=true")
	}
}
