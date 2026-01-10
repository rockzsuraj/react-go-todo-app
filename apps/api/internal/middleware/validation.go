package middleware

import (
	"net/http"
	"strings"
)

// Input validation middleware - checks for malicious content
func InputValidation(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 1. Limit request body size to 1MB (prevent huge uploads)
		r.Body = http.MaxBytesReader(w, r.Body, 1048576)
		
		// 2. Check URL for SQL injection attempts
		if containsSQLInjection(r.URL.Path) {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}
		
		// 3. Check User-Agent header for XSS attempts
		userAgent := r.Header.Get("User-Agent")
		if containsXSS(userAgent) {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}
		
		// If all checks pass, continue to next handler
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