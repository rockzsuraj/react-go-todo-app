package routes

import (
	"log/slog"
	"net/http"
	"os"
	"time"
	"react-todos/apps/api/internal/handlers"
	appMiddleware "react-todos/apps/api/internal/middleware"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func SetupRouter() http.Handler {
	r := chi.NewRouter()

	// 🔒 SECURITY MIDDLEWARE (Order matters!)
	// 1. Error handling - catch crashes first
	r.Use(appMiddleware.ErrorHandler)
	
	// 2. Rate limiting - block attackers early
	rateLimiter := appMiddleware.NewRateLimiter()
	r.Use(rateLimiter.RateLimit)
	
	// 3. Input validation - check for malicious content
	r.Use(appMiddleware.InputValidation)
	
	// 4. API key authentication - control access
	r.Use(appMiddleware.APIKeyAuth)

	// Production middleware
	logger := slog.Default()
	r.Use(appMiddleware.StructuredLogger(logger))
	r.Use(middleware.Recoverer)
	r.Use(appMiddleware.SecurityHeaders)
	r.Use(middleware.Timeout(30 * time.Second))
	r.Use(middleware.Compress(5))

	// CORS for frontend access
	r.Use(appMiddleware.CORS())

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy","service":"todo-api"}`))
	})

	// Readiness check
	r.Get("/ready", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ready"}`))
	})

	// API routes
	r.Route("/api", func(r chi.Router) {
		r.Get("/todos", handlers.GetTodos)
		r.Post("/todos", handlers.CreateTodoHandler)
		r.Put("/todos/{id}", handlers.UpdateTodoHandler)
		r.Delete("/todos/{id}", handlers.DeleteTodoHandler)
		r.Get("/health", handlers.HealthCheckHandler)
	})

	return r
}
