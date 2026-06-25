package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// maxRateLimitEntries is the upper bound on the number of IPs tracked in memory
// at any moment. When the map reaches this size a new entry evicts the oldest
// one before being inserted, preventing unbounded growth during IP-rotating
// attacks between cleanup cycles.
const maxRateLimitEntries = 10_000

type RateLimiter struct {
	rdb      *redis.Client
	visitors map[string]*visitor
	mu       sync.Mutex
}

type visitor struct {
	count       int
	windowStart time.Time
}

func NewRateLimiter(rdb *redis.Client) *RateLimiter {
	rl := &RateLimiter{
		rdb:      rdb,
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

			// Try Redis rate limiting first
			if rl.rdb != nil {
				ctx, cancel := context.WithTimeout(r.Context(), 500*time.Millisecond)
				defer cancel()

				windowStart := now.Truncate(window)
				key := "ratelimit:" + ip + ":" + strconv.FormatInt(windowStart.Unix(), 10)

				pipe := rl.rdb.TxPipeline()
				incrCmd := pipe.Incr(ctx, key)
				pipe.Expire(ctx, key, window+time.Second)
				_, err := pipe.Exec(ctx)

				if err == nil {
					count := int(incrCmd.Val())
					retryAfter := time.Until(windowStart.Add(window))

					if count > limit {
						retrySeconds := max(1, int(retryAfter.Round(time.Second)/time.Second))
						w.Header().Set("Retry-After", strconv.Itoa(retrySeconds))
						http.Error(w, "Too many requests", http.StatusTooManyRequests)
						return
					}

					next.ServeHTTP(w, r)
					return
				}

				slog.Error("Redis rate limiter failed, falling back to in-memory", "ip", ip, "error", err)
			}

			// In-memory fallback
			rl.mu.Lock()
			v, ok := rl.visitors[ip]
			if !ok {
				// Enforce the entry cap before inserting a new IP.
				if len(rl.visitors) >= maxRateLimitEntries {
					rl.evictOldest()
				}
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

// evictOldest removes the entry with the oldest window start time.
// Must be called with rl.mu held.
func (rl *RateLimiter) evictOldest() {
	var oldestIP string
	var oldestTime time.Time

	for ip, v := range rl.visitors {
		if oldestIP == "" || v.windowStart.Before(oldestTime) {
			oldestIP = ip
			oldestTime = v.windowStart
		}
	}

	if oldestIP != "" {
		delete(rl.visitors, oldestIP)
	}
}

// cleanup periodically removes entries that have been inactive for 3+ minutes,
// keeping steady-state memory low under normal traffic.
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
