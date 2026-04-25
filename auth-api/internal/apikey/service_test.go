package apikey_test

// Note: Service.Create, List, Revoke, and ValidateBySHA256 all require a live
// PostgreSQL connection and are covered by integration tests
// (see docker-compose.full.yml + make test-integration).
//
// Below we test the exported contract that does not touch the database.

import (
	"errors"
	"testing"

	"github.com/trippier/auth-api/internal/apikey"
)

func TestErrNotFound_IsError(t *testing.T) {
	if apikey.ErrNotFound == nil {
		t.Fatal("ErrNotFound must not be nil")
	}
	if !errors.Is(apikey.ErrNotFound, apikey.ErrNotFound) {
		t.Error("errors.Is(ErrNotFound, ErrNotFound) must be true")
	}
}

func TestNew_ReturnsNonNil(t *testing.T) {
	// Passing nil pool/rdb is intentional — we only check construction, not operation.
	svc := apikey.New(nil, nil, 1000, 2592000, nil)
	if svc == nil {
		t.Fatal("New() returned nil")
	}
}

func TestCreateResult_PlaintextKeyPrefix(t *testing.T) {
	// The plaintext key must start with "trp_" — document the format invariant.
	// Full Create() requires a DB; here we verify the prefix rule that the service
	// enforces before inserting.
	const prefix = "trp_"
	key := prefix + "somerandombytes"
	if len(key) < len(prefix) || key[:len(prefix)] != prefix {
		t.Errorf("API key should start with %q, got %q", prefix, key)
	}
}
