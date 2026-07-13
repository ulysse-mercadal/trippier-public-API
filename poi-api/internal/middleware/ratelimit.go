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

// rateLimitResponse is the decoded body of an auth-api rate-limit check call.
type rateLimitResponse struct {
	Allowed      bool   `json:"allowed"`
	Remaining    int    `json:"remaining"`
	Limit        int    `json:"limit"`
	ResetsInSecs int64  `json:"resets_in_secs"`
	Error        string `json:"error"`
}

// rateLimitClient is shared across all RateLimit middleware instances to reuse TCP connections.
var rateLimitClient = &http.Client{Timeout: 5 * time.Second}

// Passthrough builds a no-op middleware that forwards every request
// unconditionally (used when AUTH_DISABLED=true). It returns a gin.HandlerFunc
// that always calls Next.
func Passthrough() gin.HandlerFunc {
	return func(c *gin.Context) { c.Next() }
}

// RateLimit validates X-API-Key against the auth-api token bucket, failing
// closed with 503 if auth-api is unreachable. authAPIURL is the base URL of
// the auth-api service, internalSecret is the shared secret used for
// service-to-service HMAC auth, cost is the token cost charged for each
// request, and exempt lists paths that bypass the check entirely. It returns
// a gin.HandlerFunc enforcing the rate limit.
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

// validInternalAuth checks that the HMAC signature matches and the timestamp
// is within ±30 s. header is the value of the X-Internal-Auth header and
// secret is the shared secret used to compute the HMAC. It returns true if
// the header is a valid, fresh signature.
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

// buildInternalAuth builds the "<unix-ts>.<hmac-sha256(secret, ts)>" token
// for service-to-service auth, using secret as the shared HMAC key. It
// returns the signed internal auth token.
func buildInternalAuth(secret string) string {
	ts := strconv.FormatInt(time.Now().Unix(), 10)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(ts))
	return ts + "." + hex.EncodeToString(mac.Sum(nil))
}

// checkRateLimit calls the auth-api internal endpoint to check and consume
// rate-limit tokens. ctx is the request context propagated to the outbound
// call, client is the HTTP client used to reach auth-api, authAPIURL is the
// base URL of the auth-api service, secret is the shared secret for the
// internal auth header, apiKey is the API key being checked, and cost is the
// token cost to charge. It returns the parsed rate-limit response, or an
// error if the call fails.
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
