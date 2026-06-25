package repository

import (
	"context"
	"time"

	"react-todos/apps/api/internal/domain/models"

	"github.com/google/uuid"
)

// TodoRepository is the data-access interface for the todos domain.
type TodoRepository interface {
	GetAllByUser(ctx context.Context, userID uuid.UUID, limit, offset int, sortBy, sortOrder string, filterCompleted *bool, filterAssigned string) ([]models.Todo, int, error)
	Create(ctx context.Context, todo models.Todo) error
	Update(ctx context.Context, id int, userID uuid.UUID, todo models.Todo) error
	Delete(ctx context.Context, id int, userID uuid.UUID) error
}

type UserRepository interface {
	UpsertGoogleUser(ctx context.Context, googleUserID, email, name, picture string) (*models.User, error)
	GetByID(ctx context.Context, id string) (*models.User, error)
}

type RefreshTokenRepository interface {
	Store(ctx context.Context, rt *models.RefreshToken) error
	Get(ctx context.Context, token string) (*models.RefreshToken, error)
	Delete(ctx context.Context, token string) error
	DeleteByUser(ctx context.Context, userID string) error
}

type TokenBlacklistRepository interface {
	IsBlacklisted(ctx context.Context, jti string) (bool, error)
	Add(ctx context.Context, jti string, expiresAt time.Time) error
	BlacklistAllForUser(ctx context.Context, userID string) error
	IsUserBlacklisted(ctx context.Context, userID string) (bool, error)
	UnblockUser(ctx context.Context, userID string) error
}
