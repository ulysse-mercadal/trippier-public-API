// Package auth implements user registration, email verification, and login.
package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/trippier/auth-api/internal/email"
	"github.com/trippier/auth-api/internal/models"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrEmailTaken     = errors.New("email already registered")
	ErrNotFound       = errors.New("user not found")
	ErrBadCredentials = errors.New("invalid email or password")
	ErrNotVerified    = errors.New("email not verified")
	ErrBadToken       = errors.New("invalid or expired verification token")
)

// Service handles user auth operations.
type Service struct {
	db        *pgxpool.Pool
	emailer   *email.Sender
	jwtSecret string
	appURL    string
}

// New creates a Service.
func New(db *pgxpool.Pool, emailer *email.Sender, jwtSecret, appURL string) *Service {
	return &Service{db: db, emailer: emailer, jwtSecret: jwtSecret, appURL: appURL}
}

// Register creates an unverified account and sends a verification email.
func (s *Service) Register(ctx context.Context, emailAddr, password string) error {
	if len(password) < 8 {
		return errors.New("password must be at least 8 characters")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("bcrypt: %w", err)
	}

	token, err := randomHex(32)
	if err != nil {
		return err
	}

	_, err = s.db.Exec(ctx,
		`INSERT INTO users (email, password_hash, verification_token) VALUES ($1, $2, $3)`,
		emailAddr, string(hash), token,
	)
	if err != nil {
		if isDuplicateKey(err) {
			return ErrEmailTaken
		}
		return fmt.Errorf("insert user: %w", err)
	}

	verifyURL := fmt.Sprintf("%s/api/auth/verify-email?token=%s", s.appURL, token)
	if err := s.emailer.SendVerification(emailAddr, verifyURL); err != nil {
		fmt.Printf("warn: could not send verification email to %s: %v\n", emailAddr, err)
	}

	return nil
}

// VerifyEmail marks the account as verified and clears the token.
func (s *Service) VerifyEmail(ctx context.Context, token string) error {
	tag, err := s.db.Exec(ctx,
		`UPDATE users SET verified = true, verification_token = NULL, updated_at = NOW()
		 WHERE verification_token = $1 AND verified = false`,
		token,
	)
	if err != nil {
		return fmt.Errorf("update: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrBadToken
	}
	return nil
}

// Login validates credentials and returns a signed JWT.
func (s *Service) Login(ctx context.Context, emailAddr, password string) (string, error) {
	var user models.User
	err := s.db.QueryRow(ctx,
		`SELECT id, email, password_hash, verified FROM users WHERE email = $1`,
		emailAddr,
	).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.Verified)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrBadCredentials
	}
	if err != nil {
		return "", fmt.Errorf("query: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", ErrBadCredentials
	}
	if !user.Verified {
		return "", ErrNotVerified
	}

	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": user.ID,
		"exp": time.Now().Add(24 * time.Hour).Unix(),
		"iat": time.Now().Unix(),
	})
	return tok.SignedString([]byte(s.jwtSecret))
}

// Me returns the user for a given ID.
func (s *Service) Me(ctx context.Context, userID string) (*models.User, error) {
	var u models.User
	err := s.db.QueryRow(ctx,
		`SELECT id, email, verified, created_at FROM users WHERE id = $1`,
		userID,
	).Scan(&u.ID, &u.Email, &u.Verified, &u.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}
	return &u, nil
}

// randomHex returns n random bytes as a hex string.
func randomHex(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("rand: %w", err)
	}
	return hex.EncodeToString(b), nil
}

// isDuplicateKey detects PostgreSQL unique-constraint violations.
func isDuplicateKey(err error) bool {
	return err != nil && (contains(err.Error(), "duplicate key") || contains(err.Error(), "23505"))
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0 && containsStr(s, sub))
}

func containsStr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
