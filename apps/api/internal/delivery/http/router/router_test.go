package router

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"react-todos/apps/api/internal/infrastructure/config"
	"react-todos/apps/api/internal/delivery/http/handlers"
	appMiddleware "react-todos/apps/api/internal/delivery/http/middleware"
)

// newTestRouter builds a minimal router suitable for unit tests.
// Handler structs are constructed with nil services — tests that exercise
// protected routes should set up their own mock services.
func newTestRouter(readinessCheck ReadinessCheck) http.Handler {
	cfg := config.AppConfig{JWTSecret: "dev-jwt-secret", Env: "test"}
	return SetupRouter(
		cfg,
		nil, // authService — not needed for health/ready tests
		handlers.NewAuthHandler(nil, cfg, nil),
		handlers.NewTodoHandler(nil),
		handlers.NewSSEHandler(nil),
		handlers.NewAdminHandler(nil),
		appMiddleware.NewRateLimiter(nil),
		readinessCheck,
	)
}

func TestReadinessEndpoint(t *testing.T) {
	t.Run("ready", func(t *testing.T) {
		router := newTestRouter(func(context.Context) error { return nil })
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/ready", nil))

		if rr.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rr.Code)
		}
	})

	t.Run("dependency unavailable", func(t *testing.T) {
		router := newTestRouter(func(context.Context) error { return errors.New("db unavailable") })
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/ready", nil))

		if rr.Code != http.StatusServiceUnavailable {
			t.Fatalf("expected 503, got %d", rr.Code)
		}
	})
}
