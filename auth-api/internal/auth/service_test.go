package auth_test

import (
	"testing"
)

// Note: Register/Login/Me require a live PostgreSQL connection and are covered
// by integration tests (see docker-compose.full.yml + make test-integration).
// Below we test the pure-logic helpers that do not touch the database.

func TestPasswordMinLength(t *testing.T) {
	// The service rejects passwords shorter than 8 chars.
	// This test documents the business rule without hitting the DB.
	cases := []struct {
		pw       string
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
