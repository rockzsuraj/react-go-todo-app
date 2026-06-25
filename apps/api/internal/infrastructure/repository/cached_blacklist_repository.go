package repository

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

// CachedBlacklistRepository wraps PostgresBlacklistRepository and adds Redis caching.
// It falls back gracefully to PostgreSQL if Redis is down or unavailable.
type CachedBlacklistRepository struct {
	postgres *PostgresBlacklistRepository
	rdb      *redis.Client
}

func NewCachedBlacklistRepository(db *pgxpool.Pool, rdb *redis.Client) *CachedBlacklistRepository {
	return &CachedBlacklistRepository{
		postgres: NewPostgresBlacklistRepository(db),
		rdb:      rdb,
	}
}

func (r *CachedBlacklistRepository) Add(ctx context.Context, jti string, expiresAt time.Time) error {
	// 1. Write to Postgres
	err := r.postgres.Add(ctx, jti, expiresAt)
	if err != nil {
		return err
	}

	// 2. Write to Redis (if available)
	if r.rdb != nil {
		ttl := time.Until(expiresAt)
		if ttl > 0 {
			err := r.rdb.Set(ctx, "blacklist:jti:"+jti, "1", ttl).Err()
			if err != nil {
				slog.Error("failed to cache blacklisted token in Redis", "jti", jti, "error", err)
			}
		}
	}
	return nil
}

func (r *CachedBlacklistRepository) IsBlacklisted(ctx context.Context, jti string) (bool, error) {
	// 1. Try Redis first
	if r.rdb != nil {
		val, err := r.rdb.Get(ctx, "blacklist:jti:"+jti).Result()
		if err == nil {
			return val == "1", nil
		}
		if !errors.Is(err, redis.Nil) {
			slog.Error("Redis error during blacklist check, falling back to Postgres", "jti", jti, "error", err)
		}
	}

	// 2. Fallback to Postgres
	return r.postgres.IsBlacklisted(ctx, jti)
}

func (r *CachedBlacklistRepository) BlacklistAllForUser(ctx context.Context, userID string) error {
	// 1. Write to Postgres
	err := r.postgres.BlacklistAllForUser(ctx, userID)
	if err != nil {
		return err
	}

	// 2. Write to Redis (cache user block state, e.g., for 30 days)
	if r.rdb != nil {
		err := r.rdb.Set(ctx, "blacklist:user:"+userID, "1", 30*24*time.Hour).Err()
		if err != nil {
			slog.Error("failed to cache blocked user in Redis", "userID", userID, "error", err)
		}
	}
	return nil
}

func (r *CachedBlacklistRepository) IsUserBlacklisted(ctx context.Context, userID string) (bool, error) {
	// 1. Try Redis first
	if r.rdb != nil {
		val, err := r.rdb.Get(ctx, "blacklist:user:"+userID).Result()
		if err == nil {
			return val == "1", nil
		}
		if !errors.Is(err, redis.Nil) {
			slog.Error("Redis error during user block check, falling back to Postgres", "userID", userID, "error", err)
		}
	}

	// 2. Fallback to Postgres
	return r.postgres.IsUserBlacklisted(ctx, userID)
}

func (r *CachedBlacklistRepository) UnblockUser(ctx context.Context, userID string) error {
	// 1. Write to Postgres
	err := r.postgres.UnblockUser(ctx, userID)
	if err != nil {
		return err
	}

	// 2. Delete from Redis
	if r.rdb != nil {
		err := r.rdb.Del(ctx, "blacklist:user:"+userID).Err()
		if err != nil && !errors.Is(err, redis.Nil) {
			slog.Error("failed to remove blocked user from Redis", "userID", userID, "error", err)
		}
	}
	return nil
}
