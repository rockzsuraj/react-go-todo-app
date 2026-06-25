package middleware

import (
	"net/http"

	"github.com/go-chi/cors"
)

// CORS returns a middleware that restricts cross-origin requests to the
// explicitly allowed origins. Pass the runtime frontend URL (from AppConfig)
// so the list is never hard-coded and works across environments.
func CORS(env string, allowedOrigins ...string) func(http.Handler) http.Handler {
	origins := make([]string, 0, len(allowedOrigins)+1)
	if env != "production" {
		origins = append(origins, "http://localhost:3000")
	}
	for _, o := range allowedOrigins {
		if o != "" {
			origins = append(origins, o)
		}
	}

	return cors.Handler(cors.Options{
		AllowedOrigins:   origins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "X-API-Key"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300, // 5 minutes
	})
}
