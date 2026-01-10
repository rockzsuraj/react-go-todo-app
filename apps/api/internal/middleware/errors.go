package middleware

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"runtime/debug"
)

// Standard error response format
type ErrorResponse struct {
	Success   bool   `json:"success"`
	Error     string `json:"error"`
	Timestamp string `json:"timestamp"`
}

// Error handling middleware - catches panics and hides sensitive info
func ErrorHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Catch any panics (crashes) that happen
		defer func() {
			if err := recover(); err != nil {
				// Log the real error for developers (in server logs)
				log.Printf("PANIC: %v", err)
				log.Printf("Stack trace: %s", debug.Stack())
				
				// Send safe error to user (no technical details)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				
				// Create safe error response
				response := ErrorResponse{
					Success:   false,
					Error:     "Internal server error",
					Timestamp: "2026-01-10T07:00:00Z",
				}
				
				// Only show details in development (not production)
				if os.Getenv("ENV") == "development" {
					response.Error = "Server error - check logs for details"
				}
				
				// Send JSON response
				json.NewEncoder(w).Encode(response)
				return
			}
		}()
		
		// Continue to next handler
		next.ServeHTTP(w, r)
	})
}

// Custom error response helper
func SendErrorResponse(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	
	response := ErrorResponse{
		Success:   false,
		Error:     message,
		Timestamp: "2026-01-10T07:00:00Z",
	}
	
	json.NewEncoder(w).Encode(response)
}