package db

import (
	"context"
	"log/slog"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

// NewRedisClient creates a new connection pool for Redis and pings it to ensure health.
func NewRedisClient(rawURL string) (*redis.Client, error) {
	slog.Info("Connecting to Redis")

	opts := &redis.Options{
		Addr:         rawURL,
		DialTimeout:  3 * time.Second,
		ReadTimeout:  2 * time.Second,
		WriteTimeout: 2 * time.Second,
		PoolSize:     10,
	}
	if strings.Contains(rawURL, "://") {
		parsed, err := redis.ParseURL(rawURL)
		if err != nil {
			return nil, err
		}
		opts = parsed
		opts.DialTimeout = 3 * time.Second
		opts.ReadTimeout = 2 * time.Second
		opts.WriteTimeout = 2 * time.Second
		opts.PoolSize = 10
	}

	rdb := redis.NewClient(opts)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		slog.Error("Redis connection failed", "error", err)
		return nil, err
	}

	slog.Info("connected to Redis")
	return rdb, nil
}
