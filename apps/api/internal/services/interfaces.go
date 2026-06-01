package services

import (
	"context"
	"time"
	"react-todos/apps/api/internal/models"

	"github.com/google/uuid"
)

type TodoServicer interface {
	GetAll(ctx context.Context, userID uuid.UUID, limit, offset int, sortBy, sortOrder string, filterCompleted *bool, filterAssigned string) ([]models.Todo, int, error)
	Create(ctx context.Context, userID uuid.UUID, assignedToName, description string) error
	Update(ctx context.Context, userID uuid.UUID, id int, assignedToName, description string, completed bool) error
	Delete(ctx context.Context, userID uuid.UUID, id int) error
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
	// Optional: user-level revoke/unblock for admin
	BlacklistAllForUser(ctx context.Context, userID string) error
	IsUserBlacklisted(ctx context.Context, userID string) (bool, error)
	UnblockUser(ctx context.Context, userID string) error
}

type AuthServicer interface {
	HandleGoogleLogin(ctx context.Context, googleUserID, email, name, picture string) (*models.User, error)
	GetUserByID(ctx context.Context, id string) (*models.User, error)
	StoreRefreshToken(ctx context.Context, refreshID, userID, token string, expiresAt time.Time) error
	DeleteRefreshToken(ctx context.Context, token string) error
	ValidateAndRotateRefreshToken(ctx context.Context, token string) (string, string, error)
	BlacklistToken(ctx context.Context, jti string, expiresAt time.Time) error
	IsTokenBlacklisted(ctx context.Context, jti string) (bool, error)
	BlacklistAllForUser(ctx context.Context, userID string) error
	IsUserBlacklisted(ctx context.Context, userID string) (bool, error)
	UnblockUser(ctx context.Context, userID string) error
}
