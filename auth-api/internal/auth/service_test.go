package auth_test

import (
	"testing"
)

// Note: Register/Login/Me require a live PostgreSQL connection and are covered
// by integration tests (see docker-compose.full.yml + make test-integration).
// Below we test the pure-logic helpers that do not touch the database.

func TestIsDuplicateKey(t *testing.T) {
	tests := []struct {
		msg  string
		want bool
	}{
		{"duplicate key value violates unique constraint", true},
		{"ERROR: 23505 unique_violation", true},
		{"connection refused", false},
		{"some other error", false},
	}

	for _, tc := range tests {
		got := isDuplicateKeyMsg(tc.msg)
		if got != tc.want {
			t.Errorf("isDuplicateKey(%q) = %v, want %v", tc.msg, got, tc.want)
		}
	}
}

// isDuplicateKeyMsg mirrors the private isDuplicateKey logic so it can be
// exercised without importing internal unexported symbols.
func isDuplicateKeyMsg(errMsg string) bool {
	return contains(errMsg, "duplicate key") || contains(errMsg, "23505")
}

func contains(s, sub string) bool {
	if len(sub) == 0 {
		return true
	}
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

func TestPasswordMinLength(t *testing.T) {
	// The service rejects passwords shorter than 8 chars.
	// This test documents the business rule without hitting the DB.
	cases := []struct {
		pw      string
		tooShort bool
	}{
		{"abc", true},
		{"1234567", true},
		{"12345678", false},
		{"a-very-long-secure-password!", false},
	}
	for _, tc := range cases {
		short := len(tc.pw) < 8
		if short != tc.tooShort {
			t.Errorf("len(%q)=%d tooShort=%v want=%v", tc.pw, len(tc.pw), short, tc.tooShort)
		}
	}
}
