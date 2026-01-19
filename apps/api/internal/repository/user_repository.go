package repository

import (
	"context"
	"time"

	"react-todos/apps/api/internal/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) UpsertGoogleUser(
	ctx context.Context,
	providerUserID, email, name, picture string,
) (*models.User, error) {

	var u models.User
	now := time.Now()

	err := r.db.QueryRow(ctx, `
		INSERT INTO users (
			provider,
			provider_user_id,
			email,
			name,
			picture,
			last_login_at
		)
		VALUES ('google', $1, $2, $3, $4, $5)
		ON CONFLICT (provider, provider_user_id)
		DO UPDATE SET
			email = EXCLUDED.email,
			name = EXCLUDED.name,
			picture = EXCLUDED.picture,
			last_login_at = EXCLUDED.last_login_at,
			updated_at = now()
		RETURNING
			id, provider, provider_user_id, email, name, picture, role, is_active;
	`,
		providerUserID,
		email,
		name,
		picture,
		now,
	).Scan(
		&u.ID,
		&u.Provider,
		&u.ProviderUserID,
		&u.Email,
		&u.Name,
		&u.Picture,
		&u.Role,
		&u.IsActive,
	)

	if err != nil {
		return nil, err
	}

	return &u, nil
}

// GetByID returns a user by ID
func (r *UserRepository) GetByID(ctx context.Context, id string) (*models.User, error) {
	var u models.User

	err := r.db.QueryRow(ctx, `
		SELECT id, provider, provider_user_id, email, name, picture, role, is_active, created_at, updated_at, last_login_at
		FROM users
		WHERE id = $1
	`, id).Scan(
		&u.ID,
		&u.Provider,
		&u.ProviderUserID,
		&u.Email,
		&u.Name,
		&u.Picture,
		&u.Role,
		&u.IsActive,
		&u.CreatedAt,
		&u.UpdatedAt,
		&u.LastLoginAt,
	)

	if err != nil {
		return nil, err
	}

	return &u, nil
}
