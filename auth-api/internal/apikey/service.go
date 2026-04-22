// Package apikey manages API-key lifecycle and token-bucket state.
package apikey

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/trippier/auth-api/internal/models"
	rl "github.com/trippier/auth-api/internal/ratelimit"
	"golang.org/x/crypto/bcrypt"
)

var ErrNotFound = errors.New("api key not found")

// Service manages API keys and their Redis token buckets.
type Service struct {
	db              *pgxpool.Pool
	rdb             *redis.Client
	defaultLimit    int
	defaultInterval int // seconds
}

// New creates a Service.
func New(db *pgxpool.Pool, rdb *redis.Client, defaultLimit, defaultInterval int) *Service {
	return &Service{
		db:              db,
		rdb:             rdb,
		defaultLimit:    defaultLimit,
		defaultInterval: defaultInterval,
	}
}

// CreateResult holds the one-time plaintext key and its metadata.
type CreateResult struct {
	PlaintextKey string
	Key          models.APIKey
}

// Create generates a new API key for userID.
func (s *Service) Create(ctx context.Context, userID, name string) (*CreateResult, error) {
	raw, err := randomBytes(20) // 40-char hex → "trp_" + 40 = 44 chars total
	if err != nil {
		return nil, err
	}
	plaintext := "trp_" + raw
	prefix := plaintext[:12] // "trp_XXXXXXXX"

	h := sha256.Sum256([]byte(plaintext))
	sha256Hash := hex.EncodeToString(h[:])

	bcryptHash, err := bcrypt.GenerateFromPassword([]byte(plaintext), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("bcrypt: %w", err)
	}

	var id string
	err = s.db.QueryRow(ctx,
		`INSERT INTO api_keys
		 (user_id, name, key_hash_bcrypt, key_hash_sha256, key_prefix, tokens_limit, tokens_reset_interval_secs)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 RETURNING id`,
		userID, name, string(bcryptHash), sha256Hash, prefix, s.defaultLimit, s.defaultInterval,
	).Scan(&id)
	if err != nil {
		return nil, fmt.Errorf("insert: %w", err)
	}

	ttl := time.Duration(s.defaultInterval) * time.Second
	if err := rl.SetTokens(ctx, s.rdb, sha256Hash, s.defaultLimit, ttl); err != nil {
		fmt.Printf("warn: could not prime redis bucket for key %s: %v\n", id, err)
	}

	key := models.APIKey{
		ID:                      id,
		UserID:                  userID,
		Name:                    name,
		KeyPrefix:               prefix,
		TokensLimit:             s.defaultLimit,
		TokensResetIntervalSecs: s.defaultInterval,
		CreatedAt:               time.Now(),
	}
	return &CreateResult{PlaintextKey: plaintext, Key: key}, nil
}

// List returns all non-revoked keys for a user, enriched with Redis usage data.
func (s *Service) List(ctx context.Context, userID string) ([]models.APIKeyWithUsage, error) {
	rows, err := s.db.Query(ctx,
		`SELECT id, user_id, name, key_prefix, key_hash_sha256, tokens_limit,
		        tokens_reset_interval_secs, revoked, created_at, last_used_at
		 FROM api_keys WHERE user_id = $1 AND revoked = false ORDER BY created_at DESC`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}
	defer rows.Close()

	var keys []models.APIKeyWithUsage
	for rows.Next() {
		var k models.APIKey
		var sha256Hash string
		if err := rows.Scan(
			&k.ID, &k.UserID, &k.Name, &k.KeyPrefix, &sha256Hash,
			&k.TokensLimit, &k.TokensResetIntervalSecs, &k.Revoked,
			&k.CreatedAt, &k.LastUsedAt,
		); err != nil {
			return nil, err
		}

		remaining, ttlSecs, err := rl.GetUsage(ctx, s.rdb, sha256Hash)
		if err != nil || remaining == -1 {
			remaining = k.TokensLimit // bucket not primed yet → assume full
			ttlSecs = int64(k.TokensResetIntervalSecs)
		}

		keys = append(keys, models.APIKeyWithUsage{
			APIKey:          k,
			TokensRemaining: remaining,
			ResetsInSecs:    ttlSecs,
		})
	}
	return keys, rows.Err()
}

// Revoke marks a key as revoked.
func (s *Service) Revoke(ctx context.Context, userID, keyID string) error {
	tag, err := s.db.Exec(ctx,
		`UPDATE api_keys SET revoked = true WHERE id = $1 AND user_id = $2`,
		keyID, userID,
	)
	if err != nil {
		return fmt.Errorf("update: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// ValidateBySHA256 is the fast path used by internal middleware validation.
func (s *Service) ValidateBySHA256(ctx context.Context, sha256Hash string) (*models.InternalKeyInfo, error) {
	var info models.InternalKeyInfo
	err := s.db.QueryRow(ctx,
		`SELECT user_id, id, tokens_limit, tokens_reset_interval_secs
		 FROM api_keys
		 WHERE key_hash_sha256 = $1 AND revoked = false`,
		sha256Hash,
	).Scan(&info.UserID, &info.KeyID, &info.TokensLimit, &info.TokensResetIntervalSecs)
	if errors.Is(err, pgx.ErrNoRows) {
		return &models.InternalKeyInfo{Valid: false}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	go func() {
		ctx2, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		s.db.Exec(ctx2, `UPDATE api_keys SET last_used_at = NOW() WHERE id = $1`, info.KeyID) //nolint:errcheck
	}()

	remaining, _, err := rl.GetUsage(ctx, s.rdb, sha256Hash)
	if err != nil || remaining == -1 {
		ttl := time.Duration(info.TokensResetIntervalSecs) * time.Second
		rl.SetTokens(ctx, s.rdb, sha256Hash, info.TokensLimit, ttl) //nolint:errcheck
	}

	info.Valid = true
	return &info, nil
}

func randomBytes(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("rand: %w", err)
	}
	return hex.EncodeToString(b), nil
}
