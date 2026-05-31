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

var (
	ErrNotFound     = errors.New("api key not found")
	ErrUserNotFound = errors.New("user not found")
)

// Service manages API keys and the per-user Redis token buckets. The quota
// (tokens_limit / tokens_reset_interval_secs) lives on the users row — keys
// belonging to the same user share a single bucket.
type Service struct {
	db       *pgxpool.Pool
	rdb      *redis.Client
	log      *zap.Logger
	lastUsed chan string
}

// lastUsedBufferSize caps the bounded channel that feeds the last_used_at
// background writer. Bursts beyond this size drop the oldest pending updates
// rather than spawning unbounded goroutines when Postgres slows down.
const lastUsedBufferSize = 1024

// New creates a Service and launches a single background worker that flushes
// last_used_at updates in batches, bounded by lastUsedBufferSize.
func New(db *pgxpool.Pool, rdb *redis.Client, log *zap.Logger) *Service {
	s := &Service{db: db, rdb: rdb, log: log, lastUsed: make(chan string, lastUsedBufferSize)}
	go s.lastUsedWorker()
	return s
}

// lastUsedWorker drains the lastUsed channel and writes last_used_at for each
// key ID. Runs until the channel is closed (process exit).
func (s *Service) lastUsedWorker() {
	for keyID := range s.lastUsed {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		_, _ = s.db.Exec(ctx, `UPDATE api_keys SET last_used_at = NOW() WHERE id = $1`, keyID)
		cancel()
	}
}

// markUsed enqueues a non-blocking last_used_at update. When the worker is
// backed up the call drops silently — keeping the request-path fast.
func (s *Service) markUsed(keyID string) {
	select {
	case s.lastUsed <- keyID:
	default:
	}
}

// CreateResult holds the one-time plaintext key and its metadata.
type CreateResult struct {
	PlaintextKey string
	Key          models.APIKey
}

// Create generates a new API key for userID, reading the quota from the users row.
func (s *Service) Create(ctx context.Context, userID, name string) (*CreateResult, error) {
	limit, interval, err := s.getUserQuota(ctx, userID)
	if err != nil {
		return nil, err
	}

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
		`INSERT INTO api_keys (user_id, name, key_hash_bcrypt, key_hash_sha256, key_prefix)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id`,
		userID, name, string(bcryptHash), sha256Hash, prefix,
	).Scan(&id)
	if err != nil {
		return nil, fmt.Errorf("insert: %w", err)
	}

	ttl := time.Duration(interval) * time.Second
	if err := rl.InitBucket(ctx, s.rdb, userID, limit, ttl); err != nil {
		s.log.Warn("could not prime redis bucket", zap.String("user_id", userID), zap.Error(err))
	}

	key := models.APIKey{
		ID:                      id,
		UserID:                  userID,
		Name:                    name,
		KeyPrefix:               prefix,
		TokensLimit:             limit,
		TokensResetIntervalSecs: interval,
		CreatedAt:               time.Now(),
	}
	return &CreateResult{PlaintextKey: plaintext, Key: key}, nil
}

// List returns all non-revoked keys for a user, enriched with the user-level
// quota (from users) and live Redis usage data (shared bucket).
func (s *Service) List(ctx context.Context, userID string) ([]models.APIKeyWithUsage, error) {
	limit, interval, err := s.getUserQuota(ctx, userID)
	if err != nil {
		return nil, err
	}

	rows, err := s.db.Query(ctx,
		`SELECT id, user_id, name, key_prefix, revoked, created_at, last_used_at
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
			&k.Revoked, &k.CreatedAt, &k.LastUsedAt,
		); err != nil {
			return nil, err
		}
		k.TokensLimit = limit
		k.TokensResetIntervalSecs = interval
		rawKeys = append(rawKeys, k)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	remaining, ttlSecs, err := rl.GetUsage(ctx, s.rdb, userID)
	if err != nil || remaining == -1 {
		remaining = limit
		ttlSecs = int64(interval)
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

// ValidateBySHA256 is the fast path used by internal middleware validation. The
// returned quota reflects the current users row, so admin updates take effect
// on the next request without restart.
func (s *Service) ValidateBySHA256(ctx context.Context, sha256Hash string) (*models.InternalKeyInfo, error) {
	var info models.InternalKeyInfo
	err := s.db.QueryRow(ctx,
		`SELECT k.user_id, k.id, u.tokens_limit, u.tokens_reset_interval_secs
		   FROM api_keys k JOIN users u ON u.id = k.user_id
		  WHERE k.key_hash_sha256 = $1 AND k.revoked = false`,
		sha256Hash,
	).Scan(&info.UserID, &info.KeyID, &info.TokensLimit, &info.TokensResetIntervalSecs)
	if errors.Is(err, pgx.ErrNoRows) {
		return &models.InternalKeyInfo{Valid: false}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	s.markUsed(info.KeyID)

	remaining, _, err := rl.GetUsage(ctx, s.rdb, info.UserID)
	if err != nil || remaining == -1 {
		ttl := time.Duration(info.TokensResetIntervalSecs) * time.Second
		rl.InitBucket(ctx, s.rdb, info.UserID, info.TokensLimit, ttl) //nolint:errcheck
	}

	info.Valid = true
	return &info, nil
}

// SetUserQuota updates the quota for a user and forces the Redis bucket to the
// new limit immediately. intervalSecs == 0 keeps the existing interval.
func (s *Service) SetUserQuota(ctx context.Context, userID string, limit, intervalSecs int) error {
	var newInterval int
	err := s.db.QueryRow(ctx,
		`UPDATE users
		    SET tokens_limit               = $1,
		        tokens_reset_interval_secs = COALESCE(NULLIF($2, 0), tokens_reset_interval_secs),
		        updated_at                 = NOW()
		  WHERE id = $3
		  RETURNING tokens_reset_interval_secs`,
		limit, intervalSecs, userID,
	).Scan(&newInterval)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrUserNotFound
	}
	if err != nil {
		return fmt.Errorf("update user quota: %w", err)
	}

	ttl := time.Duration(newInterval) * time.Second
	if err := rl.SetTokens(ctx, s.rdb, userID, limit, ttl); err != nil {
		return fmt.Errorf("redis set: %w", err)
	}
	return nil
}

// SetUserQuotaByEmail is a convenience for admin tooling: resolves the user by
// email and delegates to SetUserQuota.
func (s *Service) SetUserQuotaByEmail(ctx context.Context, email string, limit, intervalSecs int) (string, error) {
	var userID string
	err := s.db.QueryRow(ctx, `SELECT id FROM users WHERE email = $1`, email).Scan(&userID)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrUserNotFound
	}
	if err != nil {
		return "", fmt.Errorf("lookup user: %w", err)
	}
	if err := s.SetUserQuota(ctx, userID, limit, intervalSecs); err != nil {
		return userID, err
	}
	return userID, nil
}

// getUserQuota returns (tokens_limit, tokens_reset_interval_secs) for a user.
func (s *Service) getUserQuota(ctx context.Context, userID string) (int, int, error) {
	var limit, interval int
	err := s.db.QueryRow(ctx,
		`SELECT tokens_limit, tokens_reset_interval_secs FROM users WHERE id = $1`,
		userID,
	).Scan(&limit, &interval)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, 0, ErrUserNotFound
	}
	if err != nil {
		return 0, 0, fmt.Errorf("user quota: %w", err)
	}
	return limit, interval, nil
}

// randomBytes returns n random bytes encoded as a hex string.
func randomBytes(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("rand: %w", err)
	}
	return hex.EncodeToString(b), nil
}
