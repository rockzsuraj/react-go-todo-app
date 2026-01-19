package middleware

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"react-todos/apps/api/internal/config"

	"github.com/golang-jwt/jwt/v5"
)

// Simple API key authentication with optional Bearer JWT support.
func APIKeyAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip auth for health checks and auth routes (support both /auth/ and /api/auth/)
		if r.URL.Path == "/health" || r.URL.Path == "/ready" || strings.HasPrefix(r.URL.Path, "/auth/") || strings.HasPrefix(r.URL.Path, "/api/auth/") {
			next.ServeHTTP(w, r)
			return
		}

		// Allow Bearer JWT tokens
		authHeader := r.Header.Get("Authorization")
		if strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
			tokenStr := strings.TrimSpace(authHeader[len("Bearer "):])
			appCfg := config.LoadAppConfig()
			if appCfg.JWTSecret != "" {
				token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
					if t.Method.Alg() != jwt.SigningMethodHS256.Alg() {
						return nil, fmt.Errorf("unexpected signing method")
					}
					return []byte(appCfg.JWTSecret), nil
				})
				if err == nil && token.Valid {
					next.ServeHTTP(w, r)
					return
				}
			}
		}

		// Get API key from request header
		apiKey := r.Header.Get("X-API-Key")

		// Check if API key is provided
		if apiKey == "" {
			sendAuthError(w, "API key required", http.StatusUnauthorized)
			return
		}

		// Validate the API key
		if !isValidAPIKey(apiKey) {
			sendAuthError(w, "Invalid API key", http.StatusUnauthorized)
			return
		}

		// API key is valid, continue to next handler
		next.ServeHTTP(w, r)
	})
}

// Check if API key is valid
func isValidAPIKey(key string) bool {
	// Get valid API keys from environment variable
	validKeys := os.Getenv("API_KEYS")
	if validKeys == "" {
		// Default key for development (change in production!)
		validKeys = "dev-key-12345,admin-key-67890"
	}

	// Split multiple keys by comma
	keys := strings.Split(validKeys, ",")

	// Check if provided key matches any valid key
	for _, validKey := range keys {
		if strings.TrimSpace(validKey) == key {
			return true // Found matching key
		}
	}

	return false // No matching key found
}

// Helper function for error responses (uses the one from errors.go)
func sendAuthError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	w.Write([]byte(`{"success":false,"error":"` + message + `"}`))
}
