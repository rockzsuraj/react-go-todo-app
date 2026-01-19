package middleware

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"runtime/debug"

	"log/slog"

	"react-todos/apps/api/internal/dto"
	"react-todos/apps/api/internal/repository"
)

// Standard error response format
// Error handling middleware - catches panics and hides sensitive info
func ErrorHandler(next http.Handler) http.Handler {
	logger := slog.Default()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				// 🔴 Log real error (server-side only)
				logger.Error(
					"panic recovered",
					"error", err,
					"stack", string(debug.Stack()),
				)

				sendJSONError(
					w,
					http.StatusInternalServerError,
					"Internal server error",
				)
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// SendError maps domain errors → HTTP responses
func SendError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrUnauthorized):
		sendJSONError(w, http.StatusUnauthorized, "Unauthorized")

	case errors.Is(err, repository.ErrNotFoundOrForbidden):
		sendJSONError(w, http.StatusNotFound, "Resource not found")

	default:
		sendJSONError(w, http.StatusInternalServerError, safeMessage())
	}
}

// Low-level JSON error sender
func sendJSONError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	apiErr := dto.ErrorResponse("ERR_GENERIC", message, "")
	_ = json.NewEncoder(w).Encode(apiErr)
}

// Environment-aware error message
func safeMessage() string {
	if os.Getenv("ENV") == "development" {
		return "Server error (check logs)"
	}
	return "Internal server error"
}

// Validation error response helper
func SendValidationError(w http.ResponseWriter, err error) {
	sendJSONError(w, http.StatusBadRequest, "Validation validation failed: "+err.Error())
}
