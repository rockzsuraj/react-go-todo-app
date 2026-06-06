package routes

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestReadinessEndpoint(t *testing.T) {
	t.Run("ready", func(t *testing.T) {
		router := SetupRouter(nil, func(context.Context) error { return nil })
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/ready", nil))

		if rr.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rr.Code)
		}
	})

	t.Run("dependency unavailable", func(t *testing.T) {
		router := SetupRouter(nil, func(context.Context) error { return errors.New("db unavailable") })
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/ready", nil))

		if rr.Code != http.StatusServiceUnavailable {
			t.Fatalf("expected 503, got %d", rr.Code)
		}
	})
}
