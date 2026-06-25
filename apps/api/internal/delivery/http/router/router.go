package router

import (
	"context"
	_ "embed"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"react-todos/apps/api/internal/infrastructure/config"
	"react-todos/apps/api/internal/delivery/http/handlers"
	appMiddleware "react-todos/apps/api/internal/delivery/http/middleware"
	"react-todos/apps/api/internal/domain/services"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

//go:embed static/.well-known/assetlinks.json
var assetLinksJSON []byte

type ReadinessCheck func(context.Context) error

// SetupRouter wires all routes. Handler structs are passed in so config is
// never re-loaded inside request handlers.
func SetupRouter(
	cfg config.AppConfig,
	authService services.AuthServicer,
	authHandler *handlers.AuthHandler,
	todoHandler *handlers.TodoHandler,
	sseHandler *handlers.SSEHandler,
	adminHandler *handlers.AdminHandler,
	authLimiter *appMiddleware.RateLimiter,
	readinessCheck ReadinessCheck,
) http.Handler {
	r := chi.NewRouter()

	// ===== GLOBAL MIDDLEWARE =====
	r.Use(appMiddleware.CORS(cfg.FrontendURL))
	r.Use(appMiddleware.ErrorHandler)
	r.Use(appMiddleware.StructuredLogger(slog.Default()))
	r.Use(middleware.Recoverer)
	r.Use(appMiddleware.SecurityHeaders)

	// ===== HEALTH (Public) =====
	r.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy"}`))
	})
	r.Get("/ready", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if readinessCheck != nil {
			ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
			defer cancel()
			if err := readinessCheck(ctx); err != nil {
				slog.Error("readiness check failed", "error", err)
				w.WriteHeader(http.StatusServiceUnavailable)
				_ = json.NewEncoder(w).Encode(map[string]string{"status": "unhealthy"})
				return
			}
		}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ready"})
	})

	// ===== ROOT ROUTE =====
	r.Get("/", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message":"Welcome to React Todos API","version":"1.0.0","endpoints":{"health":"/health","api":"/api"}}`))
	})

	// SSE stream subgroup (Authenticated, bypasses global Compress and Timeout)
	r.Group(func(r chi.Router) {
		r.Use(appMiddleware.AuthMiddleware(cfg.JWTSecret, authService))
		r.Get("/api/todos/events", sseHandler.Stream)
	})

	// Standard API subgroup (bypasses SSE, applies Compress and Timeout)
	r.Group(func(r chi.Router) {
		r.Use(middleware.Timeout(30 * time.Second))
		r.Use(middleware.Compress(5))

		r.Route("/api", func(r chi.Router) {

			// 1. Auth Sub-group
			r.Route("/auth", func(r chi.Router) {

				// --- PUBLIC AUTH ---
				// Web OAuth flow
				r.With(authLimiter.RateLimit(10, time.Minute)).
					Get("/google/login", authHandler.GoogleLogin)
				r.Get("/callback/google", authHandler.GoogleCallback)

				// Mobile: Android Credential Manager — POST { "id_token": "..." }
				r.With(authLimiter.RateLimit(10, time.Minute)).
					Post("/mobile/google", authHandler.MobileGoogleAuth)

				// Refresh is public but has a cooldown (no global rate limit)
				r.With(appMiddleware.RefreshCooldown(5*time.Second)).
					Post("/refresh", authHandler.RefreshToken)

				// --- PROTECTED AUTH ---
				r.Group(func(r chi.Router) {
					r.Use(appMiddleware.AuthMiddleware(cfg.JWTSecret, authService))

					r.Get("/me", authHandler.AuthMe)
					r.Post("/logout", authHandler.Logout)
				})
			})

			// 2. Data Sub-group (Requires Authentication)
			r.Group(func(r chi.Router) {
				r.Use(appMiddleware.AuthMiddleware(cfg.JWTSecret, authService))

				r.Get("/todos", todoHandler.GetTodos)
				r.Post("/todos", todoHandler.CreateTodoHandler)
				r.Put("/todos/{id}", todoHandler.UpdateTodoHandler)
				r.Delete("/todos/{id}", todoHandler.DeleteTodoHandler)
			})

			// 3. Admin Sub-group (Requires Authentication + Admin role)
			r.Group(func(r chi.Router) {
				r.Use(appMiddleware.AuthMiddleware(cfg.JWTSecret, authService))
				r.Use(appMiddleware.AdminOnly(cfg.JWTSecret))

				r.Post("/admin/revoke-user", adminHandler.RevokeUserTokens)
				r.Post("/admin/unblock-user", adminHandler.UnblockUser)
			})
		})
	})

	return r
}
