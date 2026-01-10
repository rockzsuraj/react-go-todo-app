package middleware

import (
	"net/http"
	"os"
	"strings"
)

// Simple API key authentication
func APIKeyAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip auth for health checks (so monitoring tools work)
		if r.URL.Path == "/health" || r.URL.Path == "/ready" {
			next.ServeHTTP(w, r)
			return
		}
		
		// Get API key from request header
		apiKey := r.Header.Get("X-API-Key")
		
		// Check if API key is provided
		if apiKey == "" {
			SendErrorResponse(w, "API key required", http.StatusUnauthorized)
			return
		}
		
		// Validate the API key
		if !isValidAPIKey(apiKey) {
			SendErrorResponse(w, "Invalid API key", http.StatusUnauthorized)
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

// Helper function for error responses (reuse from errors.go)
func SendErrorResponse(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	
	response := struct {
		Success bool   `json:"success"`
		Error   string `json:"error"`
	}{
		Success: false,
		Error:   message,
	}
	
	// Simple JSON response without importing encoding/json
	w.Write([]byte(`{"success":false,"error":"` + message + `"}`))
}