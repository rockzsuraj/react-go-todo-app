package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresBlacklistRepository struct {
	db *pgxpool.Pool
}

func NewPostgresBlacklistRepository(db *pgxpool.Pool) *PostgresBlacklistRepository {
	return &PostgresBlacklistRepository{db: db}
}

func (r *PostgresBlacklistRepository) Add(ctx context.Context, jti string, expiresAt time.Time) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO token_blacklist (jti, expires_at) VALUES ($1, $2) ON CONFLICT (jti) DO NOTHING`,
		jti, expiresAt,
	)
	return err
}

func (r *PostgresBlacklistRepository) IsBlacklisted(ctx context.Context, jti string) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM token_blacklist WHERE jti = $1 AND expires_at > now())`,
		jti,
	).Scan(&exists)
	return exists, err
}

func (r *PostgresBlacklistRepository) BlacklistAllForUser(ctx context.Context, userID string) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO user_blacklist (user_id, expires_at) VALUES ($1, $2)
		 ON CONFLICT (user_id) DO UPDATE SET expires_at = EXCLUDED.expires_at`,
		userID, time.Now().Add(30*24*time.Hour),
	)
	return err
}

func (r *PostgresBlacklistRepository) IsUserBlacklisted(ctx context.Context, userID string) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM user_blacklist WHERE user_id = $1 AND expires_at > now())`,
		userID,
	).Scan(&exists)
	return exists, err
}

func (r *PostgresBlacklistRepository) UnblockUser(ctx context.Context, userID string) error {
	_, err := r.db.Exec(ctx, `DELETE FROM user_blacklist WHERE user_id = $1`, userID)
	return err
}
