package middleware_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/trippier/poi-api/internal/middleware"
)

func init() { gin.SetMode(gin.TestMode) }

// fakeAuthAPI returns an httptest.Server that responds to /internal/check-rate-limit.
// It verifies that the caller sends a valid X-Internal-Auth header (ts.hmac format).
func fakeAuthAPI(t *testing.T, allowed bool, remaining int) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("X-Internal-Auth")
		if !strings.Contains(auth, ".") {
			t.Errorf("X-Internal-Auth header missing or malformed: %q", auth)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		body := map[string]interface{}{
			"allowed":        allowed,
			"remaining":      remaining,
			"limit":          100,
			"resets_in_secs": 3600,
		}
		if !allowed {
			body["error"] = "rate limit exceeded"
		}
		json.NewEncoder(w).Encode(body)
	}))
}

func fakeAuthAPIInvalidKey(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"allowed": false,
			"error":   "invalid api key",
		})
	}))
}

func newRateLimitRouter(authURL, secret string) *gin.Engine {
	r := gin.New()
	r.Use(middleware.RateLimit(authURL, secret, 1, "/health"))
	r.GET("/health", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) })
	r.GET("/pois/search", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"results": []string{}}) })
	return r
}

func TestRateLimit_ExemptRoute(t *testing.T) {
	// /health is exempt — no X-API-Key needed, auth-api not called.
	r := newRateLimitRouter("http://localhost:0", "secret")

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("health status = %d, want 200", w.Code)
	}
}

func TestRateLimit_MissingAPIKey(t *testing.T) {
	r := newRateLimitRouter("http://localhost:0", "secret")

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/pois/search", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", w.Code)
	}
}

func TestRateLimit_Allowed(t *testing.T) {
	srv := fakeAuthAPI(t, true, 99)
	defer srv.Close()

	r := newRateLimitRouter(srv.URL, "secret")

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/pois/search", nil)
	req.Header.Set("X-API-Key", "trp_valid_key")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
	if w.Header().Get("X-RateLimit-Remaining") != "99" {
		t.Errorf("X-RateLimit-Remaining = %q, want 99", w.Header().Get("X-RateLimit-Remaining"))
	}
}

func TestRateLimit_Exceeded(t *testing.T) {
	srv := fakeAuthAPI(t, false, 0)
	defer srv.Close()

	r := newRateLimitRouter(srv.URL, "secret")

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/pois/search", nil)
	req.Header.Set("X-API-Key", "trp_exhausted_key")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Errorf("status = %d, want 429", w.Code)
	}
}

func TestRateLimit_InvalidKey(t *testing.T) {
	srv := fakeAuthAPIInvalidKey(t)
	defer srv.Close()

	r := newRateLimitRouter(srv.URL, "secret")

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/pois/search", nil)
	req.Header.Set("X-API-Key", "trp_bad_key")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", w.Code)
	}
}

func TestRateLimit_AuthAPIDown(t *testing.T) {
	// Point to a server that refuses connections.
	r := newRateLimitRouter("http://localhost:1", "secret")

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/pois/search", nil)
	req.Header.Set("X-API-Key", "trp_some_key")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("status = %d, want 503", w.Code)
	}
}
