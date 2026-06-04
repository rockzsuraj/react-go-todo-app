package middleware

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"react-todos/apps/api/internal/services"
)

// UserIDKey is exported so it can be accessed by handlers in other packages.
type ctxKey string

const UserIDKey ctxKey = "user_id"

/* ================= MIDDLEWARE ================= */

func AuthMiddleware(jwtSecret string, authService services.AuthServicer) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			// 1. Skip strictly public health check routes
			if r.URL.Path == "/health" || r.URL.Path == "/ready" {
				next.ServeHTTP(w, r)
				return
			}

			// 2. Try JWT Authentication
			if userID, ok := authenticateJWT(r, jwtSecret, authService); ok {
				// Success: Add UserID to context and move to next handler
				ctx := context.WithValue(r.Context(), UserIDKey, userID)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			// 3. Try API Key Authentication fallback
			apiKey := r.Header.Get("X-API-Key")
			if apiKey != "" && isValidAPIKey(apiKey) {
				next.ServeHTTP(w, r)
				return
			}

			// 4. If all fails, send 401 Unauthorized
			// This uses the SendError function from your errors.go
			SendError(w, ErrUnauthorized)
		})
	}
}

/* ================= JWT LOGIC ================= */

// authenticateJWT returns (userID, success)
func authenticateJWT(r *http.Request, secret string, authService services.AuthServicer) (string, bool) {
	if secret == "" {
		return "", false
	}

	var tokenStr string

	// Try Cookie first (Web)
	if c, err := r.Cookie("token"); err == nil {
		tokenStr = c.Value
	}

	// Try Authorization header if cookie is missing (Mobile)
	if tokenStr == "" {
		authHeader := r.Header.Get("Authorization")
		if strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
			tokenStr = authHeader[len("Bearer "):]
		}
	}

	if tokenStr == "" {
		return "", false
	}

	// Parse and Validate the Token
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(secret), nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))

	if err != nil {
		slog.Warn("JWT parse error", "error", err)
		return "", false
	}
	if !token.Valid {
		slog.Warn("JWT token is invalid")
		return "", false
	}

	// Extract claims
	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		// Look for 'sub' (Subject) which usually holds the User ID
		if sub, ok := claims["sub"].(string); ok && sub != "" {
			// Check if token is blacklisted by jti
			if jti, ok := claims["jti"].(string); ok && jti != "" {
				blacklisted, err := authService.IsTokenBlacklisted(r.Context(), jti)
				if err != nil {
					slog.Error("failed to check JWT blacklist", "error", err)
					return "", false
				}
				if blacklisted {
					slog.Warn("JWT token is blacklisted", "jti", jti)
					return "", false
				}
			}
			// Check if user is blacklisted (admin revoke)
			userBlacklisted, err := authService.IsUserBlacklisted(r.Context(), sub)
			if err != nil {
				slog.Error("failed to check user blacklist", "error", err)
				return "", false
			}
			if userBlacklisted {
				slog.Warn("JWT user is blacklisted", "sub", sub)
				return "", false
			}
			return sub, true
		}
	}

	// Fixed: Returning empty string and false to match (string, bool) return signature
	return "", false
}

/* ================= CONTEXT HELPERS ================= */

// UserIDFromContext is used by Handlers to get the ID safely
func UserIDFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(UserIDKey).(string); ok {
		return v
	}
	return ""
}

// WithUserID is used for unit testing handlers
func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, UserIDKey, userID)
}

/* ================= UTILS ================= */

func isValidAPIKey(key string) bool {
	validKeys := os.Getenv("API_KEYS")
	if validKeys == "" {
		return false
	}
	for _, k := range strings.Split(validKeys, ",") {
		if strings.TrimSpace(k) == key {
			return true
		}
	}
	return false
}
