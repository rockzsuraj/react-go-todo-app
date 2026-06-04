package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRateLimiterUsesFixedWindow(t *testing.T) {
	limiter := NewRateLimiter()
	handler := limiter.RateLimit(1, 30*time.Millisecond)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	request := func() *httptest.ResponseRecorder {
		req := httptest.NewRequest("GET", "/", nil)
		req.RemoteAddr = "203.0.113.10:1234"
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		return rr
	}

	if got := request().Code; got != http.StatusNoContent {
		t.Fatalf("expected first request to pass, got %d", got)
	}

	blocked := request()
	if blocked.Code != http.StatusTooManyRequests {
		t.Fatalf("expected second request to be limited, got %d", blocked.Code)
	}
	if blocked.Header().Get("Retry-After") == "" {
		t.Fatal("expected Retry-After header")
	}

	time.Sleep(40 * time.Millisecond)
	if got := request().Code; got != http.StatusNoContent {
		t.Fatalf("expected request after original window to pass, got %d", got)
	}
}
