package middleware

import (
	"net"
	"net/http"
	"strings"
)

// GetClientIP extracts the real client IP address, respecting reverse proxies.
func GetClientIP(r *http.Request) string {
	// 1. Check X-Forwarded-For header (standard for most proxies/load balancers)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// X-Forwarded-For can contain multiple comma-separated IPs (client, proxy1, proxy2).
		// The first one is the client's original IP.
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			ip := strings.TrimSpace(ips[0])
			if ip != "" {
				return ip
			}
		}
	}

	// 2. Check X-Real-IP header (common alternative)
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// 3. Fallback to RemoteAddr
	ip, _, _ := net.SplitHostPort(r.RemoteAddr)
	return ip
}
