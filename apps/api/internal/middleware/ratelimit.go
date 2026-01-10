package middleware

import (
	"net/http"
	"sync"
	"time"
)

// Simple rate limiter - allows 60 requests per minute per IP
type RateLimiter struct {
	visitors map[string]*visitor
	mu       sync.RWMutex
}

type visitor struct {
	count    int
	lastSeen time.Time
}

func NewRateLimiter() *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*visitor),
	}
	
	// Clean up old visitors every minute
	go rl.cleanup()
	return rl
}

func (rl *RateLimiter) RateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr
		
		rl.mu.Lock()
		v, exists := rl.visitors[ip]
		if !exists {
			v = &visitor{count: 0, lastSeen: time.Now()}
			rl.visitors[ip] = v
		}
		
		// Reset count if more than 1 minute passed
		if time.Since(v.lastSeen) > time.Minute {
			v.count = 0
		}
		
		v.count++
		v.lastSeen = time.Now()
		rl.mu.Unlock()
		
		// Block if more than 60 requests per minute
		if v.count > 60 {
			http.Error(w, "Too many requests", http.StatusTooManyRequests)
			return
		}
		
		next.ServeHTTP(w, r)
	})
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