// Package middleware provides Gin middleware for auth-api.
package middleware

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

const UserIDKey = "userID"

// internalAuthWindow is the maximum clock skew accepted for X-Internal-Auth timestamps.
const internalAuthWindow = 30 * time.Second

// JWTAuth validates Bearer tokens and stores the user ID in the context.
func JWTAuth(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing or invalid Authorization header"})
			return
		}
		tokenStr := strings.TrimPrefix(header, "Bearer ")

		tok, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(secret), nil
		})
		if err != nil || !tok.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		claims, ok := tok.Claims.(jwt.MapClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid claims"})
			return
		}

		userID, _ := claims["sub"].(string)
		if userID == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing sub claim"})
			return
		}

		c.Set(UserIDKey, userID)
		c.Next()
	}
}

// InternalAuth validates X-Internal-Auth: <unix_ts>.<hmac-sha256(secret, ts)>.
// The timestamp must be within ±30 s of server time to prevent replay attacks.
func InternalAuth(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := validateInternalAuth(c.GetHeader("X-Internal-Auth"), secret); err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		c.Next()
	}
}

// validateInternalAuth parses and verifies an X-Internal-Auth header value.
// Exported for use in tests and shared tooling.
func validateInternalAuth(header, secret string) error {
	parts := strings.SplitN(header, ".", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return errors.New("malformed header")
	}

	ts, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return errors.New("invalid timestamp")
	}

	diff := time.Now().Unix() - ts
	window := int64(internalAuthWindow.Seconds())
	if diff > window || diff < -window {
		return errors.New("timestamp out of window")
	}

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(parts[0]))
	expected := hex.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(expected), []byte(parts[1])) {
		return errors.New("invalid signature")
	}
	return nil
}

// BuildInternalAuth builds a valid X-Internal-Auth header value for the given secret.
// Use this in service-to-service calls instead of sending the raw secret.
func BuildInternalAuth(secret string) string {
	ts := strconv.FormatInt(time.Now().Unix(), 10)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(ts))
	return ts + "." + hex.EncodeToString(mac.Sum(nil))
}
