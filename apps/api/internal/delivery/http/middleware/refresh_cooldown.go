package middleware

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
	"time"

	"react-todos/apps/api/internal/delivery/http/dto"
)

type refreshEntry struct {
	lastFail     time.Time
	failureCount int
}

// refreshFailures tracks per-IP refresh failure counts.
// Entries are expired by the background cleanup goroutine started in init().
var refreshFailures sync.Map // key: string (IP), value: refreshEntry

func init() {
	// Evict stale entries every minute so the map does not grow without bound.
	// An entry is considered stale when it has not seen a failure for 5× the
	// maximum cooldown window (25 s at the default 5 s cooldown).
	go func() {
		const staleness = 5 * time.Minute
		for {
			time.Sleep(time.Minute)
			refreshFailures.Range(func(k, v any) bool {
				if time.Since(v.(refreshEntry).lastFail) > staleness {
					refreshFailures.Delete(k)
				}
				return true
			})
		}
	}()
}

// RefreshCooldown blocks IPs that present invalid refresh tokens repeatedly
// within a short window. It protects the /auth/refresh endpoint from token
// brute-forcing without rate-limiting legitimate retry behaviour.
//
// Key invariant: the failure counter is reset whenever the cooldown window
// has elapsed since the last failure — not only on a successful 200 response.
// Without this reset, a legitimate user whose session expired (two 401s, then
// a pause, then a fresh login attempt) would be permanently blocked because
// every attempt is rejected before reaching the handler, making a 200 impossible.
func RefreshCooldown(cooldown time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := GetClientIP(r)

			if v, ok := refreshFailures.Load(ip); ok {
				entry := v.(refreshEntry)

				if time.Since(entry.lastFail) >= cooldown {
					// The cooldown window has elapsed since the last failure.
					// Reset the counter so the IP gets a fresh slate. Without
					// this branch the counter only resets on HTTP 200, which
					// creates a deadlock: blocked IPs can never reach the handler
					// to receive a 200 that would unblock them.
					refreshFailures.Store(ip, refreshEntry{
						lastFail:     entry.lastFail,
						failureCount: 0,
					})
				} else if entry.failureCount >= 2 {
					// Still within the cooldown window AND has hit the threshold.
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusTooManyRequests)
					_ = json.NewEncoder(w).Encode(
						dto.ErrorResponse("ERR_TOO_MANY_ATTEMPTS", "Too many refresh attempts, please try again later", ""),
					)
					return
				}
			}

			// Wrap the response writer so we can inspect the status code after
			// the handler has written it.
			rw := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(rw, r)

			switch rw.status {
			case http.StatusUnauthorized:
				// Only count as a failure when a token was actually provided.
				// A request with no token at all is just unauthenticated, not an attack.
				if !requestHasRefreshToken(r) {
					return
				}
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

			case http.StatusOK:
				// Successful refresh — reset the failure counter for this IP.
				refreshFailures.Delete(ip)
			}
		})
	}
}

// requestHasRefreshToken reports whether the request carries a refresh token
// either as a cookie (web) or a Bearer header (mobile).
func requestHasRefreshToken(r *http.Request) bool {
	if _, err := r.Cookie("refresh_token"); err == nil {
		return true
	}
	return strings.HasPrefix(r.Header.Get("Authorization"), "Bearer ")
}

// statusRecorder captures the HTTP status code written by a handler so that
// middleware sitting above it can inspect the outcome.
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}
