package middleware

import (
	"net/http"
	"strings"
)

func ExtractToken(r *http.Request) string {
	// 1️⃣ Prefer Authorization header
	auth := r.Header.Get("Authorization")
	if strings.HasPrefix(auth, "Bearer ") {
		return strings.TrimPrefix(auth, "Bearer ")
	}

	// 2️⃣ Fallback to cookie
	if c, err := r.Cookie("token"); err == nil {
		return c.Value
	}

	return ""
}
