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
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

var ErrNotFound = errors.New("api key not found")

// Service manages API keys and their Redis token buckets.
type Service struct {
	db              *pgxpool.Pool
	rdb             *redis.Client
	log             *zap.Logger
	defaultLimit    int
	defaultInterval int // seconds
}

// New creates a Service.
func New(db *pgxpool.Pool, rdb *redis.Client, defaultLimit, defaultInterval int, log *zap.Logger) *Service {
	return &Service{
		db:              db,
		rdb:             rdb,
		log:             log,
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

	// Prime the user-level bucket only if it does not already exist — a second
	// key for the same user must not reset the remaining token count.
	ttl := time.Duration(s.defaultInterval) * time.Second
	if err := rl.InitBucket(ctx, s.rdb, userID, s.defaultLimit, ttl); err != nil {
		s.log.Warn("could not prime redis bucket", zap.String("user_id", userID), zap.Error(err))
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

// List returns all non-revoked keys for a user, enriched with the shared
// user-level Redis usage data (all keys draw from the same pool).
func (s *Service) List(ctx context.Context, userID string) ([]models.APIKeyWithUsage, error) {
	rows, err := s.db.Query(ctx,
		`SELECT id, user_id, name, key_prefix,
		        tokens_limit, tokens_reset_interval_secs, revoked, created_at, last_used_at
		 FROM api_keys WHERE user_id = $1 AND revoked = false ORDER BY created_at DESC`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}
	defer rows.Close()

	var rawKeys []models.APIKey
	for rows.Next() {
		var k models.APIKey
		if err := rows.Scan(
			&k.ID, &k.UserID, &k.Name, &k.KeyPrefix,
			&k.TokensLimit, &k.TokensResetIntervalSecs, &k.Revoked,
			&k.CreatedAt, &k.LastUsedAt,
		); err != nil {
			return nil, err
		}
		rawKeys = append(rawKeys, k)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Fetch the shared user-level bucket once for all keys.
	remaining, ttlSecs, err := rl.GetUsage(ctx, s.rdb, userID)
	if err != nil || remaining == -1 {
		remaining = s.defaultLimit
		ttlSecs = int64(s.defaultInterval)
	}

	keys := make([]models.APIKeyWithUsage, 0, len(rawKeys))
	for _, k := range rawKeys {
		keys = append(keys, models.APIKeyWithUsage{
			APIKey:          k,
			TokensRemaining: remaining,
			ResetsInSecs:    ttlSecs,
		})
	}
	return keys, nil
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
		`SELECT user_id, id FROM api_keys WHERE key_hash_sha256 = $1 AND revoked = false`,
		sha256Hash,
	).Scan(&info.UserID, &info.KeyID)
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

	// Use global config for limit/interval — the bucket is per-user, not per-key.
	info.TokensLimit = s.defaultLimit
	info.TokensResetIntervalSecs = s.defaultInterval

	// Prime the user bucket lazily if it disappeared from Redis.
	remaining, _, err := rl.GetUsage(ctx, s.rdb, info.UserID)
	if err != nil || remaining == -1 {
		ttl := time.Duration(s.defaultInterval) * time.Second
		rl.InitBucket(ctx, s.rdb, info.UserID, s.defaultLimit, ttl) //nolint:errcheck
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
