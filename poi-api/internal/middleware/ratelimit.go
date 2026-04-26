package middleware

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type rateLimitResponse struct {
	Allowed      bool   `json:"allowed"`
	Remaining    int    `json:"remaining"`
	Limit        int    `json:"limit"`
	ResetsInSecs int64  `json:"resets_in_secs"`
	Error        string `json:"error"`
}

// rateLimitClient is shared across all RateLimit middleware instances to reuse TCP connections.
var rateLimitClient = &http.Client{Timeout: 5 * time.Second}

// Passthrough returns a no-op middleware that forwards every request unconditionally.
// Used when AUTH_DISABLED=true to run the API without an auth-api dependency.
func Passthrough() gin.HandlerFunc {
	return func(c *gin.Context) { c.Next() }
}

// RateLimit validates X-API-Key against the auth-api token bucket.
// Paths listed in exempt and requests with a valid X-Internal-Auth header bypass the check entirely.
// On auth-api unavailability the request is rejected with 503 to prevent unlimited access.
func RateLimit(authAPIURL, internalSecret string, cost int, exempt ...string) gin.HandlerFunc {
	exemptSet := make(map[string]struct{}, len(exempt))
	for _, p := range exempt {
		exemptSet[p] = struct{}{}
	}
	client := rateLimitClient

	return func(c *gin.Context) {
		if _, ok := exemptSet[c.FullPath()]; ok {
			c.Next()
			return
		}

		if h := c.GetHeader("X-Internal-Auth"); h != "" && validInternalAuth(h, internalSecret) {
			c.Next()
			return
		}

		apiKey := c.GetHeader("X-API-Key")
		if apiKey == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "X-API-Key header required"})
			return
		}

		rlResp, err := checkRateLimit(c.Request.Context(), client, authAPIURL, internalSecret, apiKey, cost)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"error": "rate-limit service unavailable"})
			return
		}

		if !rlResp.Allowed {
			if rlResp.Error == "invalid api key" {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid api key"})
				return
			}
			c.Header("X-RateLimit-Remaining", "0")
			c.Header("X-RateLimit-Reset", strconv.FormatInt(
				time.Now().Add(time.Duration(rlResp.ResetsInSecs)*time.Second).Unix(), 10,
			))
			c.Header("Retry-After", strconv.FormatInt(rlResp.ResetsInSecs, 10))
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error":          "rate limit exceeded",
				"resets_in_secs": rlResp.ResetsInSecs,
			})
			return
		}

		c.Header("X-RateLimit-Limit", strconv.Itoa(rlResp.Limit))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(rlResp.Remaining))
		c.Header("X-RateLimit-Reset", strconv.FormatInt(
			time.Now().Add(time.Duration(rlResp.ResetsInSecs)*time.Second).Unix(), 10,
		))
		c.Next()
	}
}

// validInternalAuth returns true if the HMAC signature matches and the timestamp is within ±30 s.
func validInternalAuth(header, secret string) bool {
	parts := strings.SplitN(header, ".", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return false
	}
	ts, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return false
	}
	diff := time.Now().Unix() - ts
	if diff > 30 || diff < -30 {
		return false
	}
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(parts[0]))
	expected := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(parts[1]))
}

// buildInternalAuth returns "<unix-ts>.<hmac-sha256(secret, ts)>" for service-to-service auth.
func buildInternalAuth(secret string) string {
	ts := strconv.FormatInt(time.Now().Unix(), 10)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(ts))
	return ts + "." + hex.EncodeToString(mac.Sum(nil))
}

func checkRateLimit(ctx context.Context, client *http.Client, authAPIURL, secret, apiKey string, cost int) (*rateLimitResponse, error) {
	payload, _ := json.Marshal(map[string]interface{}{
		"api_key": apiKey,
		"cost":    cost,
	})

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		fmt.Sprintf("%s/internal/check-rate-limit", authAPIURL),
		bytes.NewReader(payload),
	)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Internal-Auth", buildInternalAuth(secret))

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var rl rateLimitResponse
	if err := json.Unmarshal(body, &rl); err != nil {
		return nil, err
	}
	return &rl, nil
}
