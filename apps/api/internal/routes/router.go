package routes

import (
	"log/slog"
	"net/http"
	"time"

	"react-todos/apps/api/internal/config"
	"react-todos/apps/api/internal/handlers"
	appMiddleware "react-todos/apps/api/internal/middleware"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func SetupRouter() http.Handler {
	r := chi.NewRouter()

	// ===== GLOBAL MIDDLEWARE =====
	r.Use(appMiddleware.CORS())
	r.Use(appMiddleware.ErrorHandler)
	r.Use(appMiddleware.StructuredLogger(slog.Default()))
	r.Use(middleware.Recoverer)
	r.Use(appMiddleware.SecurityHeaders)
	r.Use(middleware.Timeout(30 * time.Second))
	r.Use(middleware.Compress(5))

	// ===== HEALTH =====
	r.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy"}`))
	})

	r.Get("/ready", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ready"}`))
	})

	// ===== AUTH =====
	// Public and protected auth routes share the same base path. Mount once
	// and apply JWT middleware only to the protected subset.
	r.Route("/api/auth", func(r chi.Router) {
		// Public
		r.Get("/google/login", handlers.GoogleLogin)
		r.Get("/callback/google", handlers.GoogleCallback)
		r.Post("/refresh", handlers.RefreshToken)

		// Protected
		r.Group(func(r chi.Router) {
			r.Use(appMiddleware.JWTAuth(config.LoadAppConfig().JWTSecret))
			r.Get("/me", handlers.AuthMe)
			r.Post("/logout", handlers.Logout)
		})
	})

	// ===== API (PROTECTED) =====
	r.Route("/api", func(r chi.Router) {
		r.Use(appMiddleware.JWTAuth(config.LoadAppConfig().JWTSecret))
		r.Get("/todos", handlers.GetTodos)
		r.Post("/todos", handlers.CreateTodoHandler)
		r.Put("/todos/{id}", handlers.UpdateTodoHandler)
		r.Delete("/todos/{id}", handlers.DeleteTodoHandler)
	})

	return r
}
