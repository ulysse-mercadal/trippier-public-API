// Package models defines shared data structures.
package models

import "time"

// User represents an account row.
type User struct {
	ID                UUID      `json:"id"`
	Email             string    `json:"email"`
	PasswordHash      string    `json:"-"`
	Verified          bool      `json:"verified"`
	VerificationToken string    `json:"-"`
	CreatedAt         time.Time `json:"created_at"`
}

// UUID is a string alias used for clarity.
type UUID = string
