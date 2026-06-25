package middleware

import (
	"net"
	"net/http"
	"strings"
)

// GetClientIP trusts forwarding headers only when the direct peer is a local or
// private-network reverse proxy. This prevents public clients from spoofing
// their identity to bypass IP-based controls.
func GetClientIP(r *http.Request) string {
	remoteIP := parseIP(r.RemoteAddr)
	if remoteIP == nil {
		return r.RemoteAddr
	}

	if remoteIP.IsPrivate() || remoteIP.IsLoopback() {
		if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
			for _, candidate := range strings.Split(xff, ",") {
				if ip := net.ParseIP(strings.TrimSpace(candidate)); ip != nil {
					return ip.String()
				}
			}
		}

		if ip := net.ParseIP(strings.TrimSpace(r.Header.Get("X-Real-IP"))); ip != nil {
			return ip.String()
		}
	}

	return remoteIP.String()
}

func parseIP(address string) net.IP {
	host, _, err := net.SplitHostPort(address)
	if err == nil {
		return net.ParseIP(host)
	}
	return net.ParseIP(address)
}
