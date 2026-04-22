package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

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
	token := signedToken(secret, "user-123", -time.Hour) // expired

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

func TestInternalAuth_ValidSecret(t *testing.T) {
	r := gin.New()
	r.POST("/internal", mw.InternalAuth("my-secret"), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/internal", nil)
	req.Header.Set("X-Internal-Secret", "my-secret")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

func TestInternalAuth_InvalidSecret(t *testing.T) {
	r := gin.New()
	r.POST("/internal", mw.InternalAuth("my-secret"), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/internal", nil)
	req.Header.Set("X-Internal-Secret", "wrong")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", w.Code)
	}
}
