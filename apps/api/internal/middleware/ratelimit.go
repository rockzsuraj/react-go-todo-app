package middleware

import (
	"net/http"
	"strconv"
	"sync"
	"time"
)

type RateLimiter struct {
	visitors map[string]*visitor
	mu       sync.Mutex
}

type visitor struct {
	count       int
	windowStart time.Time
}

func NewRateLimiter() *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*visitor),
	}
	go rl.cleanup()
	return rl
}

func (rl *RateLimiter) RateLimit(limit int, window time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := GetClientIP(r)

			now := time.Now()

			rl.mu.Lock()
			v, ok := rl.visitors[ip]
			if !ok {
				v = &visitor{windowStart: now}
				rl.visitors[ip] = v
			}

			if now.Sub(v.windowStart) >= window {
				v.count = 0
				v.windowStart = now
			}

			v.count++
			count := v.count
			retryAfter := time.Until(v.windowStart.Add(window))
			rl.mu.Unlock()

			if count > limit {
				retrySeconds := max(1, int(retryAfter.Round(time.Second)/time.Second))
				w.Header().Set("Retry-After", strconv.Itoa(retrySeconds))
				http.Error(w, "Too many requests", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func (rl *RateLimiter) cleanup() {
	for {
		time.Sleep(time.Minute)
		rl.mu.Lock()
		for ip, v := range rl.visitors {
			if time.Since(v.windowStart) > 3*time.Minute {
				delete(rl.visitors, ip)
			}
		}
		rl.mu.Unlock()
	}
}
