package middleware

import (
	"net/http/httptest"
	"testing"
)

func TestGetClientIPIgnoresForwardedHeadersFromPublicPeer(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "203.0.113.10:1234"
	req.Header.Set("X-Forwarded-For", "198.51.100.20")

	if got := GetClientIP(req); got != "203.0.113.10" {
		t.Fatalf("expected direct peer IP, got %q", got)
	}
}

func TestGetClientIPTrustsForwardedHeadersFromPrivateProxy(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "10.0.0.2:1234"
	req.Header.Set("X-Forwarded-For", "198.51.100.20, 10.0.0.1")

	if got := GetClientIP(req); got != "198.51.100.20" {
		t.Fatalf("expected forwarded client IP, got %q", got)
	}
}
