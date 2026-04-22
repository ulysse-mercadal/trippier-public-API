package apikey_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	rl "github.com/trippier/auth-api/internal/ratelimit"
)

func init() { gin.SetMode(gin.TestMode) }

// newTestRedis starts an in-process Redis and returns a client for it.
func newTestRedis(t *testing.T) (*miniredis.Miniredis, *redis.Client) {
	t.Helper()
	mr := miniredis.RunT(t)
	return mr, redis.NewClient(&redis.Options{Addr: mr.Addr()})
}

func postJSON(t *testing.T, r http.Handler, path string, body interface{}) *httptest.ResponseRecorder {
	t.Helper()
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

// checkRateLimitHandler is an isolated implementation of the check-rate-limit
// endpoint that uses miniredis directly — no PostgreSQL required.
// It mirrors exactly the logic in apikey.Handler.checkRateLimit but replaces
// the DB-backed ValidateBySHA256 call with a pre-seeded bucket.
func checkRateLimitHandler(rdb *redis.Client, tokensLimit, resetIntervalSecs int) gin.HandlerFunc {
	return func(c *gin.Context) {
		var body struct {
			APIKey string `json:"api_key" binding:"required"`
			Cost   int    `json:"cost"    binding:"required,min=1"`
		}
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		sum := sha256.Sum256([]byte(body.APIKey))
		hash := hex.EncodeToString(sum[:])

		remaining, ttlSecs, notFound, insufficient, err := rl.Deduct(c.Request.Context(), rdb, hash, body.Cost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "rate-limit error"})
			return
		}
		if notFound {
			ttl := time.Duration(resetIntervalSecs) * time.Second
			_ = rl.SetTokens(c.Request.Context(), rdb, hash, tokensLimit, ttl)
			remaining, ttlSecs, _, insufficient, err = rl.Deduct(c.Request.Context(), rdb, hash, body.Cost)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "rate-limit error"})
				return
			}
		}
		if insufficient {
			c.JSON(http.StatusOK, gin.H{
				"allowed":        false,
				"remaining":      0,
				"limit":          tokensLimit,
				"resets_in_secs": ttlSecs,
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"allowed":        true,
			"remaining":      remaining,
			"limit":          tokensLimit,
			"resets_in_secs": ttlSecs,
		})
	}
}

func seedBucket(t *testing.T, rdb *redis.Client, plaintext string, limit int) {
	t.Helper()
	sum := sha256.Sum256([]byte(plaintext))
	hash := hex.EncodeToString(sum[:])
	if err := rl.SetTokens(context.Background(), rdb, hash, limit, time.Hour); err != nil {
		t.Fatalf("seed bucket: %v", err)
	}
}

func TestCheckRateLimit_Allowed(t *testing.T) {
	mr, rdb := newTestRedis(t)
	defer mr.Close()

	plaintext := "trp_test_key_abc123"
	seedBucket(t, rdb, plaintext, 100)

	r := gin.New()
	r.POST("/internal/check-rate-limit", checkRateLimitHandler(rdb, 100, 3600))

	w := postJSON(t, r, "/internal/check-rate-limit", map[string]interface{}{
		"api_key": plaintext,
		"cost":    10,
	})

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d body = %s", w.Code, w.Body.String())
	}
	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)

	if resp["allowed"] != true {
		t.Errorf("expected allowed=true, got %v", resp["allowed"])
	}
	if remaining, ok := resp["remaining"].(float64); !ok || remaining != 90 {
		t.Errorf("expected remaining=90, got %v", resp["remaining"])
	}
}

func TestCheckRateLimit_Insufficient(t *testing.T) {
	mr, rdb := newTestRedis(t)
	defer mr.Close()

	plaintext := "trp_test_key_def456"
	seedBucket(t, rdb, plaintext, 5)

	r := gin.New()
	r.POST("/internal/check-rate-limit", checkRateLimitHandler(rdb, 5, 3600))

	w := postJSON(t, r, "/internal/check-rate-limit", map[string]interface{}{
		"api_key": plaintext,
		"cost":    10,
	})

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d body = %s", w.Code, w.Body.String())
	}
	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)

	if resp["allowed"] != false {
		t.Errorf("expected allowed=false when cost > remaining, got %v", resp["allowed"])
	}
}

func TestCheckRateLimit_BucketNotFound_AutoInit(t *testing.T) {
	mr, rdb := newTestRedis(t)
	defer mr.Close()

	// No pre-seeding: bucket does not exist in Redis.
	r := gin.New()
	r.POST("/internal/check-rate-limit", checkRateLimitHandler(rdb, 1000, 3600))

	w := postJSON(t, r, "/internal/check-rate-limit", map[string]interface{}{
		"api_key": "trp_fresh_key_xyz",
		"cost":    1,
	})

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d body = %s", w.Code, w.Body.String())
	}
	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)

	if resp["allowed"] != true {
		t.Errorf("expected auto-init to allow first request, got %v", resp["allowed"])
	}
}

func TestCheckRateLimit_InvalidInput(t *testing.T) {
	mr, rdb := newTestRedis(t)
	defer mr.Close()

	r := gin.New()
	r.POST("/internal/check-rate-limit", checkRateLimitHandler(rdb, 100, 3600))

	// Missing api_key.
	w := postJSON(t, r, "/internal/check-rate-limit", map[string]interface{}{
		"cost": 5,
	})
	if w.Code != http.StatusBadRequest {
		t.Errorf("missing api_key: status = %d, want 400", w.Code)
	}

	// Cost = 0 (min=1 binding).
	w = postJSON(t, r, "/internal/check-rate-limit", map[string]interface{}{
		"api_key": "trp_some_key",
		"cost":    0,
	})
	if w.Code != http.StatusBadRequest {
		t.Errorf("cost=0: status = %d, want 400", w.Code)
	}
}
