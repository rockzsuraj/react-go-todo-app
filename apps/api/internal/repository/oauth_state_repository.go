package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type OAuthStateRepository struct {
	db *pgxpool.Pool
}

func NewOAuthStateRepository(db *pgxpool.Pool) *OAuthStateRepository {
	return &OAuthStateRepository{db: db}
}

// Store persists a new OAuth state with a TTL.
func (r *OAuthStateRepository) Store(ctx context.Context, state string, ttl time.Duration) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO oauth_states (state, expires_at) VALUES ($1, $2)`,
		state, time.Now().Add(ttl),
	)
	return err
}

// Consume validates and atomically deletes the state — returns false if not found or expired.
func (r *OAuthStateRepository) Consume(ctx context.Context, state string) (bool, error) {
	tag, err := r.db.Exec(ctx,
		`DELETE FROM oauth_states WHERE state = $1 AND expires_at > now()`,
		state,
	)
	if err != nil {
		return false, err
	}
	return tag.RowsAffected() == 1, nil
}

// DeleteExpired cleans up stale rows — call periodically or on startup.
func (r *OAuthStateRepository) DeleteExpired(ctx context.Context) error {
	_, err := r.db.Exec(ctx, `DELETE FROM oauth_states WHERE expires_at <= now()`)
	return err
}
