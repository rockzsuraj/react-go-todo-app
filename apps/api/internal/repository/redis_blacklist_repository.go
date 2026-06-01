package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisBlacklistRepository struct {
	client *redis.Client
}

func NewRedisBlacklistRepository(client *redis.Client) *RedisBlacklistRepository {
	return &RedisBlacklistRepository{client: client}
}

// Add adds a JWT jti to the blacklist with TTL until its original expiration
func (r *RedisBlacklistRepository) Add(ctx context.Context, jti string, expiresAt time.Time) error {
	ttl := time.Until(expiresAt)
	if ttl <= 0 {
		return fmt.Errorf("token already expired")
	}
	key := "blacklist:" + jti
	return r.client.Set(ctx, key, "1", ttl).Err()
}

// IsBlacklisted checks if a JWT jti is in the blacklist
func (r *RedisBlacklistRepository) IsBlacklisted(ctx context.Context, jti string) (bool, error) {
	key := "blacklist:" + jti
	_, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// BlacklistAllForUser blacklists all active tokens for a user_id by storing a marker.
// Middleware must check both jti and user_id markers.
func (r *RedisBlacklistRepository) BlacklistAllForUser(ctx context.Context, userID string) error {
	key := "user_blacklist:" + userID
	// Set a long TTL (e.g., 30 days) or manage via explicit unblock
	return r.client.Set(ctx, key, "1", 30*24*time.Hour).Err()
}

// IsUserBlacklisted checks if a user_id is blacklisted (all tokens revoked)
func (r *RedisBlacklistRepository) IsUserBlacklisted(ctx context.Context, userID string) (bool, error) {
	key := "user_blacklist:" + userID
	_, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// UnblockUser removes a user-level blacklist
func (r *RedisBlacklistRepository) UnblockUser(ctx context.Context, userID string) error {
	key := "user_blacklist:" + userID
	return r.client.Del(ctx, key).Err()
}

// Stats returns simple stats for monitoring (optional)
func (r *RedisBlacklistRepository) Stats(ctx context.Context) (map[string]string, error) {
	jtiCount, err := r.client.Keys(ctx, "blacklist:*").Result()
	if err != nil {
		return nil, err
	}
	userCount, err := r.client.Keys(ctx, "user_blacklist:*").Result()
	if err != nil {
		return nil, err
	}
	return map[string]string{
		"blacklisted_jtis": fmt.Sprintf("%d", len(jtiCount)),
		"blacklisted_users": fmt.Sprintf("%d", len(userCount)),
	}, nil
}
