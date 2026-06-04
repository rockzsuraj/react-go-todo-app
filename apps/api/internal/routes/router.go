package routes

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"react-todos/apps/api/internal/config"
	"react-todos/apps/api/internal/handlers"
	appMiddleware "react-todos/apps/api/internal/middleware"
	"react-todos/apps/api/internal/services"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type ReadinessCheck func(context.Context) error

func SetupRouter(authService services.AuthServicer, readinessCheck ReadinessCheck) http.Handler {
	r := chi.NewRouter()
	cfg := config.LoadAppConfig()

	// ===== GLOBAL MIDDLEWARE =====
	r.Use(appMiddleware.CORS())
	r.Use(appMiddleware.ErrorHandler)
	r.Use(appMiddleware.StructuredLogger(slog.Default()))
	r.Use(middleware.Recoverer)
	r.Use(appMiddleware.SecurityHeaders)
	r.Use(middleware.Timeout(30 * time.Second))
	r.Use(middleware.Compress(5))

	authLimiter := appMiddleware.NewRateLimiter()

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
		w.Write([]byte("{\"message\":\"Welcome to React Todos API\",\"version\":\"1.0.0\",\"endpoints\":{\"health\":\"/health\",\"api\":\"/api\"}}"))
	})

	r.Route("/api", func(r chi.Router) {

		// 1. Auth Sub-group
		r.Route("/auth", func(r chi.Router) {

			// --- PUBLIC AUTH ---
			r.With(authLimiter.RateLimit(10, time.Minute)).
				Get("/google/login", handlers.GoogleLogin)
			r.Get("/callback/google", handlers.GoogleCallback)

			// Refresh is public but has a cooldown (no global rate limit)
			r.With(appMiddleware.RefreshCooldown(30*time.Second)).
				Post("/refresh", handlers.RefreshToken)

			// --- PROTECTED AUTH ---
			r.Group(func(r chi.Router) {
				r.Use(appMiddleware.AuthMiddleware(cfg.JWTSecret, authService))

				r.Get("/me", handlers.AuthMe)
				r.Post("/logout", handlers.Logout)
			})
		})

		// 2. Data Sub-group (Requires Authentication)
		r.Group(func(r chi.Router) {
			r.Use(appMiddleware.AuthMiddleware(cfg.JWTSecret, authService))

			r.Get("/todos", handlers.GetTodos)
			r.Post("/todos", handlers.CreateTodoHandler)
			r.Put("/todos/{id}", handlers.UpdateTodoHandler)
			r.Delete("/todos/{id}", handlers.DeleteTodoHandler)
		})

		// 3. Admin Sub-group (Requires Authentication + Admin role)
		r.Group(func(r chi.Router) {
			r.Use(appMiddleware.AuthMiddleware(cfg.JWTSecret, authService))
			r.Use(appMiddleware.AdminOnly(cfg.JWTSecret))

			r.Post("/admin/revoke-user", handlers.RevokeUserTokens)
			r.Post("/admin/unblock-user", handlers.UnblockUser)
		})
	})

	return r
}
