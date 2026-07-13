// Package auth implements user registration, email verification, and login.
package auth

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/trippier/auth-api/internal/email"
	"github.com/trippier/auth-api/internal/models"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrEmailTaken      = errors.New("email already registered")
	ErrWeakPassword    = errors.New("password must be at least 8 characters")
	ErrBadCredentials  = errors.New("invalid email or password")
	ErrNotVerified     = errors.New("email not verified")
	ErrBadToken        = errors.New("invalid or expired verification code")
	ErrAlreadyVerified = errors.New("email already verified")
)

// dummyHash is used in Login for constant-time bcrypt comparison when the email doesn't exist, to prevent enumeration via timing.
var dummyHash = mustBcrypt([]byte("dummy-sentinel-value"))

// mustBcrypt hashes in with bcrypt and returns the resulting hash, panicking
// if hashing fails.
func mustBcrypt(in []byte) []byte {
	h, err := bcrypt.GenerateFromPassword(in, bcrypt.DefaultCost)
	if err != nil {
		panic(fmt.Sprintf("bcrypt init: %v", err))
	}
	return h
}

// Service handles user auth operations.
type Service struct {
	db        *pgxpool.Pool
	emailer   *email.Sender
	jwtSecret string
	appURL    string
}

// New creates a Service using db as the connection pool, emailer to send
// verification emails, jwtSecret to sign JWTs, and appURL as the base URL of
// the app included in emails. It returns the configured Service.
func New(db *pgxpool.Pool, emailer *email.Sender, jwtSecret, appURL string) *Service {
	return &Service{db: db, emailer: emailer, jwtSecret: jwtSecret, appURL: appURL}
}

// Register creates an unverified account for emailAddr with the given
// password and sends a 6-digit OTP code by email; clientIP and userAgent
// identify the requester in the OTP email. It returns an error if
// registration fails.
func (s *Service) Register(ctx context.Context, emailAddr, password, clientIP, userAgent string) error {
	if len(password) < 8 {
		return ErrWeakPassword
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("bcrypt: %w", err)
	}

	code, err := randomCode()
	if err != nil {
		return err
	}

	_, err = s.db.Exec(ctx,
		`INSERT INTO users (email, password_hash, verification_token, verification_token_expires_at)
		 VALUES ($1, $2, $3, NOW() + INTERVAL '15 minutes')`,
		strings.ToLower(emailAddr), string(hash), code,
	)
	if err != nil {
		if isDuplicateKey(err) {
			return ErrEmailTaken
		}
		return fmt.Errorf("insert user: %w", err)
	}

	if err := s.emailer.SendOTPCode(emailAddr, code, clientIP, userAgent, s.appURL); err != nil {
		return fmt.Errorf("send otp email: %w", err)
	}

	return nil
}

// VerifyCode verifies that code is the current 6-digit OTP for emailAddr,
// marks the account as verified, and returns a signed JWT so the user is
// immediately logged in, along with an error, if any.
func (s *Service) VerifyCode(ctx context.Context, emailAddr, code string) (string, error) {
	var userID string
	err := s.db.QueryRow(ctx,
		`UPDATE users
		    SET verified = true,
		        verification_token = NULL,
		        verification_token_expires_at = NULL,
		        updated_at = NOW()
		  WHERE email = $1
		    AND verification_token = $2
		    AND verified = false
		    AND verification_token_expires_at > NOW()
		  RETURNING id`,
		strings.ToLower(emailAddr), code,
	).Scan(&userID)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrBadToken
	}
	if err != nil {
		return "", fmt.Errorf("verify code: %w", err)
	}

	return s.signJWT(userID)
}

// Login validates emailAddr and password against stored credentials and
// returns a signed JWT along with an error, if any.
func (s *Service) Login(ctx context.Context, emailAddr, password string) (string, error) {
	var user models.User
	err := s.db.QueryRow(ctx,
		`SELECT id, email, password_hash, verified FROM users WHERE email = $1`,
		strings.ToLower(emailAddr),
	).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.Verified)
	if errors.Is(err, pgx.ErrNoRows) {
		_ = bcrypt.CompareHashAndPassword(dummyHash, []byte(password))
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

	return s.signJWT(user.ID)
}

// ResendCode generates a new OTP for the unverified account with emailAddr
// and sends it by email; clientIP and userAgent identify the requester in
// the OTP email. It returns an error if resending fails.
func (s *Service) ResendCode(ctx context.Context, emailAddr, clientIP, userAgent string) error {
	code, err := randomCode()
	if err != nil {
		return err
	}

	result, err := s.db.Exec(ctx,
		`UPDATE users
		    SET verification_token = $2,
		        verification_token_expires_at = NOW() + INTERVAL '15 minutes',
		        updated_at = NOW()
		  WHERE email = $1
		    AND verified = false`,
		strings.ToLower(emailAddr), code,
	)
	if err != nil {
		return fmt.Errorf("resend code: %w", err)
	}
	if result.RowsAffected() == 0 {
		return ErrAlreadyVerified
	}

	if err := s.emailer.SendOTPCode(emailAddr, code, clientIP, userAgent, s.appURL); err != nil {
		return fmt.Errorf("send otp email: %w", err)
	}

	return nil
}

// Me fetches the user record with the given userID and returns it along
// with an error, if any.
func (s *Service) Me(ctx context.Context, userID string) (*models.User, error) {
	var u models.User
	err := s.db.QueryRow(ctx,
		`SELECT id, email, verified, created_at FROM users WHERE id = $1`,
		userID,
	).Scan(&u.ID, &u.Email, &u.Verified, &u.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}
	return &u, nil
}

// signJWT creates a signed JWT with userID as the subject claim and returns
// it along with an error, if any.
func (s *Service) signJWT(userID string) (string, error) {
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(24 * time.Hour).Unix(),
		"iat": time.Now().Unix(),
	})
	return tok.SignedString([]byte(s.jwtSecret))
}

// randomCode returns a uniformly random 6-digit decimal string, using
// rejection sampling to avoid modulo bias, along with an error, if any.
func randomCode() (string, error) {
	const maxBias = (1 << 32) - ((1 << 32) % 1_000_000)
	for {
		var n uint32
		if err := binary.Read(rand.Reader, binary.BigEndian, &n); err != nil {
			return "", fmt.Errorf("rand: %w", err)
		}
		if n < maxBias {
			return fmt.Sprintf("%06d", n%1_000_000), nil
		}
	}
}

// isDuplicateKey reports whether err represents a PostgreSQL
// unique-constraint (duplicate-key) violation.
func isDuplicateKey(err error) bool {
	return err != nil && (strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "23505"))
}
