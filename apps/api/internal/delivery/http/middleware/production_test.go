package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSecurityHeadersSetHSTSOnlyInProduction(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	prod := httptest.NewRecorder()
	SecurityHeaders("production")(next).ServeHTTP(prod, httptest.NewRequest(http.MethodGet, "/", nil))
	if got := prod.Header().Get("Strict-Transport-Security"); got == "" {
		t.Fatal("expected HSTS in production")
	}

	dev := httptest.NewRecorder()
	SecurityHeaders("development")(next).ServeHTTP(dev, httptest.NewRequest(http.MethodGet, "/", nil))
	if got := dev.Header().Get("Strict-Transport-Security"); got != "" {
		t.Fatalf("expected no HSTS in development, got %q", got)
	}
}
