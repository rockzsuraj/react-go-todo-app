package middleware

import (
	"net/http"
	"strings"
)

// Input validation middleware - checks for malicious content
func InputValidation(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
		next.ServeHTTP(w, r)
	})
}

// Check for SQL injection patterns
func containsSQLInjection(input string) bool {
	// Common SQL injection patterns
	dangerous := []string{
		"'", "\"", ";", "--", "/*", "*/",
		"union", "select", "insert", "delete",
		"update", "drop", "exec", "script",
	}

	// Convert to lowercase for checking
	lower := strings.ToLower(input)

	// Check if input contains any dangerous patterns
	for _, pattern := range dangerous {
		if strings.Contains(lower, pattern) {
			return true // Found dangerous content!
		}
	}
	return false // Safe
}

// Check for XSS (Cross-Site Scripting) patterns
func containsXSS(input string) bool {
	// Common XSS attack patterns
	dangerous := []string{
		"<script", "</script>", "javascript:",
		"onload=", "onerror=", "eval(",
		"alert(", "document.cookie",
	}

	// Convert to lowercase for checking
	lower := strings.ToLower(input)

	// Check if input contains any dangerous patterns
	for _, pattern := range dangerous {
		if strings.Contains(lower, pattern) {
			return true // Found dangerous content!
		}
	}
	return false // Safe
}
