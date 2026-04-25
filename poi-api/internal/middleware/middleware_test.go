package middleware_test

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/trippier/poi-api/internal/middleware"
	"go.uber.org/zap"
)

func validAuth(secret string) string {
	ts := strconv.FormatInt(time.Now().Unix(), 10)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(ts))
	return ts + "." + hex.EncodeToString(mac.Sum(nil))
}

// ── SecureHeaders ────────────────────────────────────────────────────────────

func TestSecureHeaders(t *testing.T) {
	r := gin.New()
	r.Use(middleware.SecureHeaders())
	r.GET("/", func(c *gin.Context) { c.Status(http.StatusOK) })

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/", nil))

	for _, h := range []string{"X-Content-Type-Options", "X-Frame-Options", "Referrer-Policy"} {
		if w.Header().Get(h) == "" {
			t.Errorf("missing header %s", h)
		}
	}
}

// ── RequestID ────────────────────────────────────────────────────────────────

func TestRequestID_Generated(t *testing.T) {
	r := gin.New()
	r.Use(middleware.RequestID())
	r.GET("/", func(c *gin.Context) {
		id := middleware.RequestIDFromCtx(c)
		if id == "" {
			t.Error("request ID not set in context")
		}
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/", nil))

	if w.Header().Get("X-Request-Id") == "" {
		t.Error("X-Request-Id header not set")
	}
}

func TestRequestID_Propagated(t *testing.T) {
	r := gin.New()
	r.Use(middleware.RequestID())
	r.GET("/", func(c *gin.Context) { c.Status(http.StatusOK) })

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Request-Id", "existing-id")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if got := w.Header().Get("X-Request-Id"); got != "existing-id" {
		t.Errorf("X-Request-Id = %q, want existing-id", got)
	}
}

// ── Logger ───────────────────────────────────────────────────────────────────

func TestLogger_DoesNotPanic(t *testing.T) {
	for _, status := range []int{200, 400, 500} {
		r := gin.New()
		r.Use(middleware.Logger(zap.NewNop()))
		r.GET("/", func(c *gin.Context) { c.Status(status) })

		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/?q=test", nil))
	}
}

// ── validInternalAuth ────────────────────────────────────────────────────────

func TestRateLimit_ValidInternalAuth(t *testing.T) {
	r := newRateLimitRouter("http://localhost:0", "mysecret")

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/pois/search", nil)
	req.Header.Set("X-Internal-Auth", validAuth("mysecret"))
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("valid internal auth: status = %d, want 200", w.Code)
	}
}

func TestRateLimit_ExpiredInternalAuth(t *testing.T) {
	r := newRateLimitRouter("http://localhost:0", "mysecret")

	// Timestamp 60s in the past — outside the ±30s window.
	ts := strconv.FormatInt(time.Now().Unix()-60, 10)
	mac := hmac.New(sha256.New, []byte("mysecret"))
	mac.Write([]byte(ts))
	expired := ts + "." + hex.EncodeToString(mac.Sum(nil))

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/pois/search", nil)
	req.Header.Set("X-Internal-Auth", expired)
	r.ServeHTTP(w, req)

	// Falls through to missing X-API-Key.
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expired internal auth: status = %d, want 401", w.Code)
	}
}

func TestRateLimit_MalformedInternalAuth(t *testing.T) {
	r := newRateLimitRouter("http://localhost:0", "mysecret")

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/pois/search", nil)
	req.Header.Set("X-Internal-Auth", "not-valid")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("malformed internal auth: status = %d, want 401", w.Code)
	}
}

// ── Cache ────────────────────────────────────────────────────────────────────

func newCacheRouter(rdb *redis.Client) *gin.Engine {
	r := gin.New()
	r.Use(middleware.Cache(rdb, time.Minute))
	r.GET("/pois/search", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"results": []string{}})
	})
	return r
}

func newRedis(t *testing.T) *redis.Client {
	t.Helper()
	mr := miniredis.RunT(t)
	return redis.NewClient(&redis.Options{Addr: mr.Addr()})
}

func TestCache_Miss(t *testing.T) {
	r := newCacheRouter(newRedis(t))

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/pois/search?lat=1&lng=2", nil))

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	if w.Header().Get("X-Cache") != "MISS" {
		t.Errorf("X-Cache = %q, want MISS", w.Header().Get("X-Cache"))
	}
}

func TestCache_Hit(t *testing.T) {
	r := newCacheRouter(newRedis(t))
	req := httptest.NewRequest(http.MethodGet, "/pois/search?lat=1&lng=2", nil)

	// First request populates the cache.
	r.ServeHTTP(httptest.NewRecorder(), req)

	// Second request should be a HIT.
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/pois/search?lat=1&lng=2", nil))

	if w.Header().Get("X-Cache") != "HIT" {
		t.Errorf("X-Cache = %q, want HIT", w.Header().Get("X-Cache"))
	}
}

func TestCache_Bypass_NoCache(t *testing.T) {
	r := newCacheRouter(newRedis(t))
	req := httptest.NewRequest(http.MethodGet, "/pois/search", nil)
	req.Header.Set("X-No-Cache", "1")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Header().Get("X-Cache") != "" {
		t.Errorf("X-No-Cache should bypass cache, got X-Cache=%q", w.Header().Get("X-Cache"))
	}
}

func TestCache_Bypass_NonGET(t *testing.T) {
	r := gin.New()
	r.Use(middleware.Cache(newRedis(t), time.Minute))
	r.POST("/pois/search", func(c *gin.Context) { c.Status(http.StatusOK) })

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/pois/search", nil))

	if w.Header().Get("X-Cache") != "" {
		t.Error("POST request should not be cached")
	}
}

func TestCache_KeyParameterOrderIndependent(t *testing.T) {
	rdb := newRedis(t)
	r := newCacheRouter(rdb)

	// Populate with one param order.
	r.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/pois/search?lng=2&lat=1", nil))

	// Different param order — should be a HIT (same canonical key).
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/pois/search?lat=1&lng=2", nil))

	if w.Header().Get("X-Cache") != "HIT" {
		t.Errorf("param-order-independent cache: X-Cache = %q, want HIT", w.Header().Get("X-Cache"))
	}
}
