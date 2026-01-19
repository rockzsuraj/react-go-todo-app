package services

import (
	"context"
	"errors"
	"time"

	"react-todos/apps/api/internal/models"
	"react-todos/apps/api/internal/repository"
)

type AuthService struct {
	userRepo         *repository.UserRepository
	refreshTokenRepo *repository.RefreshTokenRepository
}

func NewAuthService(userRepo *repository.UserRepository, refreshTokenRepo *repository.RefreshTokenRepository) *AuthService {
	return &AuthService{
		userRepo:         userRepo,
		refreshTokenRepo: refreshTokenRepo,
	}
}

// HandleGoogleLogin persists or updates a Google user and returns the user
func (s *AuthService) HandleGoogleLogin(
	ctx context.Context,
	googleUserID string,
	email string,
	name string,
	picture string,
) (*models.User, error) {

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return s.userRepo.UpsertGoogleUser(
		ctx,
		googleUserID,
		email,
		name,
		picture,
	)
}

// GetUserByID returns a user by ID
func (s *AuthService) GetUserByID(ctx context.Context, id string) (*models.User, error) {
	return s.userRepo.GetByID(ctx, id)
}

func (s *AuthService) StoreRefreshToken(ctx context.Context, refreshID, userID, token string, expiresAt time.Time) error {
	return s.refreshTokenRepo.Store(ctx, &models.RefreshToken{
		ID:        refreshID,
		UserID:    userID,
		Token:     token,
		ExpiresAt: expiresAt,
	})
}

func (s *AuthService) DeleteRefreshToken(ctx context.Context, token string) error {
	return s.refreshTokenRepo.Delete(ctx, token)
}

func (s *AuthService) ValidateAndRotateRefreshToken(ctx context.Context, token string) (string, error) {
	storedToken, err := s.refreshTokenRepo.Get(ctx, token)
	if err != nil {
		return "", err
	}
	if storedToken == nil {
		return "", errors.New("invalid token")
	}

	if time.Now().After(storedToken.ExpiresAt) {
		s.refreshTokenRepo.Delete(ctx, token)
		return "", errors.New("token expired")
	}

	// In a real rotation scenario, we might issue a new refresh token here.
	// For now, we just validate and return the UserID to issue a new access token.
	return storedToken.UserID, nil
}
