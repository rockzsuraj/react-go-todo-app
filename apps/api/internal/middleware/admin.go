package middleware

import (
	"net/http"

	"github.com/golang-jwt/jwt/v5"
)

// AdminOnly middleware checks that the JWT has role = "admin"
func AdminOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := UserIDFromContext(r.Context())
		if userID == "" {
			SendError(w, ErrUnauthorized)
			return
		}

		// Extract JWT from header or cookie to read role claim
		tokenStr := ExtractToken(r)
		if tokenStr == "" {
			SendError(w, ErrUnauthorized)
			return
		}

		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, nil
			}
			// We don't need to verify signature here; it's already verified in AuthMiddleware
			return []byte("dummy"), nil
		})
		if err != nil || !token.Valid {
			SendError(w, ErrUnauthorized)
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			if role, ok := claims["role"].(string); ok && role == "admin" {
				next.ServeHTTP(w, r)
				return
			}
		}

		SendError(w, ErrForbidden)
	})
}
