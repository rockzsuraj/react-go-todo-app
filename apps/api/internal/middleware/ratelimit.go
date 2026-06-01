package middleware

import (
	"net"
	"net/http"
	"sync"
	"time"
)

type RateLimiter struct {
	visitors map[string]*visitor
	mu       sync.Mutex
}

type visitor struct {
	count    int
	lastSeen time.Time
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
			ip, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				http.Error(w, "invalid IP", http.StatusBadRequest)
				return
			}

			now := time.Now()

			rl.mu.Lock()
			v, ok := rl.visitors[ip]
			if !ok {
				v = &visitor{}
				rl.visitors[ip] = v
			}

			if now.Sub(v.lastSeen) > window {
				v.count = 0
			}

			v.count++
			v.lastSeen = now
			count := v.count
			rl.mu.Unlock()

			if count > limit {
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
			if time.Since(v.lastSeen) > 3*time.Minute {
				delete(rl.visitors, ip)
			}
		}
		rl.mu.Unlock()
	}
}
