// Package models defines shared data structures.
package models

import "time"

// User represents an account row. The DB carries a verification_token column
// but it is never read into Go — the OTP flow queries it inline by SQL, so
// keeping it as a struct field would only add a populated-but-unread surface.
type User struct {
	ID           UUID      `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Verified     bool      `json:"verified"`
	CreatedAt    time.Time `json:"created_at"`
}

// UUID is a string alias used for clarity.
type UUID = string
