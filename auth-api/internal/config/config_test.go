package config_test

import (
	"strings"
	"testing"

	"github.com/trippier/auth-api/internal/config"
)

func TestLoad_DefaultPort(t *testing.T) {
	t.Setenv("AUTH_JWT_SECRET", "32-char-secret-value-for-testing!")
	t.Setenv("AUTH_INTERNAL_SECRET", "16-char-secret!!")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Port == "" {
		t.Error("Port should have a default value")
	}
}

func TestLoad_ExplicitValues(t *testing.T) {
	t.Setenv("AUTH_PORT", "9999")
	t.Setenv("AUTH_JWT_SECRET", "32-char-secret-value-for-testing!")
	t.Setenv("AUTH_INTERNAL_SECRET", "16-char-secret!!")
	t.Setenv("AUTH_SMTP_FROM", "custom@example.com")
	t.Setenv("AUTH_LOG_LEVEL", "debug")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Port != "9999" {
		t.Errorf("Port = %q, want 9999", cfg.Port)
	}
	if cfg.SMTPFrom != "custom@example.com" {
		t.Errorf("SMTPFrom = %q, want custom@example.com", cfg.SMTPFrom)
	}
	if cfg.LogLevel != "debug" {
		t.Errorf("LogLevel = %q, want debug", cfg.LogLevel)
	}
}

func TestLoad_JWTSecretTooShort(t *testing.T) {
	t.Setenv("AUTH_JWT_SECRET", "tooshort")
	t.Setenv("AUTH_INTERNAL_SECRET", "16-char-secret!!")

	_, err := config.Load()
	if err == nil {
		t.Fatal("expected error for JWT secret shorter than 32 chars")
	}
	if !strings.Contains(err.Error(), "AUTH_JWT_SECRET") {
		t.Errorf("error should mention AUTH_JWT_SECRET, got: %v", err)
	}
}

func TestLoad_InternalSecretTooShort(t *testing.T) {
	t.Setenv("AUTH_JWT_SECRET", "32-char-secret-value-for-testing!")
	t.Setenv("AUTH_INTERNAL_SECRET", "tooshort")

	_, err := config.Load()
	if err == nil {
		t.Fatal("expected error for internal secret shorter than 16 chars")
	}
	if !strings.Contains(err.Error(), "AUTH_INTERNAL_SECRET") {
		t.Errorf("error should mention AUTH_INTERNAL_SECRET, got: %v", err)
	}
}

func TestLoad_JWTSecretExactly32Chars(t *testing.T) {
	t.Setenv("AUTH_JWT_SECRET", "12345678901234567890123456789012") // exactly 32
	t.Setenv("AUTH_INTERNAL_SECRET", "1234567890123456")            // exactly 16

	_, err := config.Load()
	if err != nil {
		t.Fatalf("32-char JWT secret should be accepted, got: %v", err)
	}
}

func TestLoad_DefaultTokenLimits(t *testing.T) {
	t.Setenv("AUTH_JWT_SECRET", "32-char-secret-value-for-testing!")
	t.Setenv("AUTH_INTERNAL_SECRET", "16-char-secret!!")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.DefaultTokensLimit <= 0 {
		t.Errorf("DefaultTokensLimit should be positive, got %d", cfg.DefaultTokensLimit)
	}
	if cfg.DefaultResetIntervalS <= 0 {
		t.Errorf("DefaultResetIntervalS should be positive, got %d", cfg.DefaultResetIntervalS)
	}
}
