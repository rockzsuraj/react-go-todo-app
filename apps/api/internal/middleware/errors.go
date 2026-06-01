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

// ErrUnauthorized is a standard error used by AuthMiddleware
var ErrUnauthorized = errors.New("unauthorized")

// ErrForbidden is used for admin/permission checks
var ErrForbidden = errors.New("forbidden")

// ErrorHandler catches panics to prevent the server from crashing and hides sensitive info
func ErrorHandler(next http.Handler) http.Handler {
	logger := slog.Default()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				logger.Error(
					"panic recovered",
					"error", err,
					"stack", string(debug.Stack()),
				)

				sendJSONError(
					w,
					http.StatusInternalServerError,
					"ERR_INTERNAL",
					"Internal server error",
				)
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// SendError maps domain/auth errors to standardized HTTP responses
func SendError(w http.ResponseWriter, err error) {
	switch {
	// Matches the error returned by AuthMiddleware
	case errors.Is(err, ErrUnauthorized):
		sendJSONError(w, http.StatusUnauthorized, "ERR_UNAUTHORIZED", "Unauthorized")

	// Admin/permission denied
	case errors.Is(err, ErrForbidden):
		sendJSONError(w, http.StatusForbidden, "ERR_FORBIDDEN", "Forbidden")

	// Matches database/repository errors
	case errors.Is(err, repository.ErrNotFoundOrForbidden):
		sendJSONError(w, http.StatusNotFound, "ERR_NOT_FOUND", "Resource not found")

	default:
		sendJSONError(w, http.StatusInternalServerError, "ERR_INTERNAL", safeMessage())
	}
}

// sendJSONError is the low-level helper for formatting JSON responses
func sendJSONError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	// Ensure your dto.ErrorResponse matches your frontend expectations
	apiErr := dto.ErrorResponse(code, message, "")
	_ = json.NewEncoder(w).Encode(apiErr)
}

func safeMessage() string {
	if os.Getenv("ENV") == "development" {
		return "Server error (check logs)"
	}
	return "Internal server error"
}

func SendValidationError(w http.ResponseWriter, err error) {
	sendJSONError(w, http.StatusBadRequest, "ERR_VALIDATION", "Validation failed: "+err.Error())
}

// SendJSONErrorWithCode allows sending custom error codes
func SendJSONErrorWithCode(w http.ResponseWriter, status int, code, message string) {
	sendJSONError(w, status, code, message)
}
