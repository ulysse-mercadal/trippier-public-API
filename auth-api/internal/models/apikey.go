package models

import "time"

// APIKey represents a key row (never exposes the raw secret).
type APIKey struct {
	ID                       string     `json:"id"`
	UserID                   string     `json:"user_id"`
	Name                     string     `json:"name"`
	KeyPrefix                string     `json:"key_prefix"`
	TokensLimit              int        `json:"tokens_limit"`
	TokensResetIntervalSecs  int        `json:"tokens_reset_interval_secs"`
	Revoked                  bool       `json:"revoked"`
	CreatedAt                time.Time  `json:"created_at"`
	LastUsedAt               *time.Time `json:"last_used_at"`
}

// APIKeyWithUsage extends APIKey with live Redis-backed usage data.
type APIKeyWithUsage struct {
	APIKey
	TokensRemaining int   `json:"tokens_remaining"`
	ResetsInSecs    int64 `json:"resets_in_secs"`
}

// InternalKeyInfo is returned to other services via the internal endpoint.
type InternalKeyInfo struct {
	Valid                    bool   `json:"valid"`
	UserID                   string `json:"user_id"`
	KeyID                    string `json:"key_id"`
	TokensLimit              int    `json:"tokens_limit"`
	TokensResetIntervalSecs  int    `json:"tokens_reset_interval_secs"`
}
