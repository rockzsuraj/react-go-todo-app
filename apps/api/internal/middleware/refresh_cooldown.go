package middleware

import (
	"encoding/json"
	"net"
	"net/http"
	"sync"
	"time"

	"react-todos/apps/api/internal/dto"
)

type refreshEntry struct {
	lastFail     time.Time
	failureCount int
}

var refreshFailures sync.Map // key: ip, value: refreshEntry

func RefreshCooldown(cooldown time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip, _, _ := net.SplitHostPort(r.RemoteAddr)

			if v, ok := refreshFailures.Load(ip); ok {
				entry := v.(refreshEntry)
				// Only block if there are multiple failures in quick succession
				if time.Since(entry.lastFail) < cooldown {
					// Check if this is a repeated failure pattern
					if entry.failureCount >= 2 {
						w.Header().Set("Content-Type", "application/json")
						w.WriteHeader(http.StatusTooManyRequests)
						response := dto.ErrorResponse("ERR_TOO_MANY_ATTEMPTS", "Too many refresh attempts, please try again later", "")
						_ = json.NewEncoder(w).Encode(response)
						return
					}
				}
			}

			// Wrap response to detect failure
			rw := &statusRecorder{ResponseWriter: w, status: 200}
			next.ServeHTTP(rw, r)

			// Track failures with count
			if rw.status == http.StatusUnauthorized {
				if v, ok := refreshFailures.Load(ip); ok {
					entry := v.(refreshEntry)
					entry.lastFail = time.Now()
					entry.failureCount++
					refreshFailures.Store(ip, entry)
				} else {
					refreshFailures.Store(ip, refreshEntry{
						lastFail:     time.Now(),
						failureCount: 1,
					})
				}
			} else if rw.status == http.StatusOK {
				// Reset failure count on success
				if v, ok := refreshFailures.Load(ip); ok {
					entry := v.(refreshEntry)
					entry.failureCount = 0
					refreshFailures.Store(ip, entry)
				}
			}
		})
	}
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}
