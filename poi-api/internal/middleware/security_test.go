package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/trippier/poi-api/internal/middleware"
)

func newSecurityRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.SecureHeaders())
	r.Use(middleware.CORS())
	r.GET("/test", func(c *gin.Context) { c.Status(http.StatusOK) })
	return r
}

func TestCORS_AllowOrigin(t *testing.T) {
	r := newSecurityRouter()
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/test", nil))

	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "*" {
		t.Errorf("Access-Control-Allow-Origin = %q, want %q", got, "*")
	}
}

func TestCORS_AllowMethods(t *testing.T) {
	r := newSecurityRouter()
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/test", nil))

	methods := w.Header().Get("Access-Control-Allow-Methods")
	if methods == "" {
		t.Error("Access-Control-Allow-Methods should not be empty")
	}
}

func TestCORS_AllowHeaders_ContainsByokHeaders(t *testing.T) {
	r := newSecurityRouter()
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/test", nil))

	headers := w.Header().Get("Access-Control-Allow-Headers")
	for _, expected := range []string{
		"X-API-Key",
		"X-Ticketmaster-Key",
		"X-Eventbrite-Token",
		"X-Meetup-Token",
		"X-Foursquare-Key",
		"X-Baidu-Key",
	} {
		if !containsHeader(headers, expected) {
			t.Errorf("Access-Control-Allow-Headers missing %q", expected)
		}
	}
}

func TestCORS_Preflight(t *testing.T) {
	r := newSecurityRouter()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodOptions, "/test", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("OPTIONS status = %d, want 204", w.Code)
	}
}

func containsHeader(headers, target string) bool {
	// Simple substring check — the header list is comma-separated with spaces.
	return len(headers) > 0 && (headers == target ||
		len(headers) >= len(target) &&
			(headers[:len(target)] == target ||
				len(headers) > len(target) && containsSubstring(headers, target)))
}

func containsSubstring(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
