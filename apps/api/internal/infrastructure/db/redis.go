package db

import (
	"context"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
)

// NewRedisClient creates a new connection pool for Redis and pings it to ensure health.
func NewRedisClient(addr string) (*redis.Client, error) {
	slog.Info("Connecting to Redis", "addr", addr)
	rdb := redis.NewClient(&redis.Options{
		Addr:         addr,
		DialTimeout:  3 * time.Second,
		ReadTimeout:  2 * time.Second,
		WriteTimeout: 2 * time.Second,
		PoolSize:     10,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		slog.Error("Redis connection failed", "error", err)
		return nil, err
	}

	slog.Info("✅ Connected to Redis")
	return rdb, nil
}
