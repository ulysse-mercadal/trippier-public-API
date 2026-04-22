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

// RateLimit validates the caller's X-API-Key against the auth-api token bucket.
// Each request costs `cost` tokens. Routes in exempt are passed through without
// a key (e.g. "/health", "/pois/providers").
//
// On auth-api unavailability the request is rejected with 503 to prevent
// unlimited access during outages.
func RateLimit(authAPIURL, internalSecret string, cost int, exempt ...string) gin.HandlerFunc {
	exemptSet := make(map[string]struct{}, len(exempt))
	for _, p := range exempt {
		exemptSet[p] = struct{}{}
	}
	client := &http.Client{Timeout: 5 * time.Second}

	return func(c *gin.Context) {
		if _, ok := exemptSet[c.FullPath()]; ok {
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

// buildInternalAuth returns an X-Internal-Auth header value: "<ts>.<hmac-sha256(secret, ts)>".
// Using a timestamp-bound HMAC prevents replaying a captured header.
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
