package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	mw "github.com/trippier/auth-api/internal/middleware"
)

func init() { gin.SetMode(gin.TestMode) }

func signedToken(secret, subject string, expOffset time.Duration) string {
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": subject,
		"exp": time.Now().Add(expOffset).Unix(),
		"iat": time.Now().Unix(),
	})
	s, _ := tok.SignedString([]byte(secret))
	return s
}

// buildInternalAuth mirrors the production helper for use in tests.
func buildInternalAuth(secret string) string {
	ts := strconv.FormatInt(time.Now().Unix(), 10)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(ts))
	return ts + "." + hex.EncodeToString(mac.Sum(nil))
}

func newRouter(secret string) *gin.Engine {
	r := gin.New()
	r.GET("/protected", mw.JWTAuth(secret), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"user": c.GetString(mw.UserIDKey)})
	})
	return r
}

func TestJWTAuth_ValidToken(t *testing.T) {
	secret := "test-secret-32-chars-long-enough!"
	token := signedToken(secret, "user-123", time.Hour)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	newRouter(secret).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

func TestJWTAuth_MissingHeader(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)

	newRouter("secret").ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", w.Code)
	}
}

func TestJWTAuth_ExpiredToken(t *testing.T) {
	secret := "test-secret"
	token := signedToken(secret, "user-123", -time.Hour)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	newRouter(secret).ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", w.Code)
	}
}

func TestJWTAuth_WrongSecret(t *testing.T) {
	token := signedToken("correct-secret", "user-123", time.Hour)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	newRouter("wrong-secret").ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", w.Code)
	}
}

func TestJWTAuth_MalformedToken(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer not.a.jwt")

	newRouter("secret").ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", w.Code)
	}
}

// ── InternalAuth (HMAC-based) ─────────────────────────────────────────────────

func newInternalRouter(secret string) *gin.Engine {
	r := gin.New()
	r.POST("/internal", mw.InternalAuth(secret), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})
	return r
}

func TestInternalAuth_ValidHMAC(t *testing.T) {
	r := newInternalRouter("my-secret")

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/internal", nil)
	req.Header.Set("X-Internal-Auth", buildInternalAuth("my-secret"))
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

func TestInternalAuth_WrongSecret(t *testing.T) {
	r := newInternalRouter("correct-secret")

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/internal", nil)
	req.Header.Set("X-Internal-Auth", buildInternalAuth("wrong-secret"))
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", w.Code)
	}
}

func TestInternalAuth_MissingHeader(t *testing.T) {
	r := newInternalRouter("my-secret")

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/internal", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", w.Code)
	}
}

func TestInternalAuth_MalformedHeader(t *testing.T) {
	r := newInternalRouter("my-secret")

	for _, bad := range []string{"notimestamp", ".", "abc.def.extra", ""} {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/internal", nil)
		req.Header.Set("X-Internal-Auth", bad)
		r.ServeHTTP(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Errorf("header=%q: status = %d, want 401", bad, w.Code)
		}
	}
}

func TestInternalAuth_ExpiredTimestamp(t *testing.T) {
	r := newInternalRouter("my-secret")

	// Build a header with a timestamp 60s in the past (outside the 30s window).
	oldTS := strconv.FormatInt(time.Now().Add(-60*time.Second).Unix(), 10)
	mac := hmac.New(sha256.New, []byte("my-secret"))
	mac.Write([]byte(oldTS))
	staleHeader := oldTS + "." + hex.EncodeToString(mac.Sum(nil))

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/internal", nil)
	req.Header.Set("X-Internal-Auth", staleHeader)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401 for expired timestamp", w.Code)
	}
}
