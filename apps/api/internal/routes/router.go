package routes

import (
	"net/http"
	"react-todos/internal/handlers"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func SetupRouter() http.Handler {
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Route("/api", func(r chi.Router) {
		r.Get("/todos", handlers.GetTodosHandler)
		r.Post("/todos", handlers.CreateTodoHandler)
		r.Put("/todos/{id}", handlers.UpdateTodoHandler)
		r.Delete("/todos/{id}", handlers.DeleteTodoHandler)
	})

	return r
}
