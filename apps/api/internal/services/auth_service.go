package services

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"log/slog"
	"time"

	"react-todos/apps/api/internal/models"

	"github.com/google/uuid"
)

type AuthService struct {
	userRepo         UserRepository
	refreshTokenRepo RefreshTokenRepository
	blacklistRepo    TokenBlacklistRepository
}

func NewAuthService(userRepo UserRepository, refreshTokenRepo RefreshTokenRepository, blacklistRepo TokenBlacklistRepository) *AuthService {
	return &AuthService{
		userRepo:         userRepo,
		refreshTokenRepo: refreshTokenRepo,
		blacklistRepo:    blacklistRepo,
	}
}

func generateRefreshToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
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

func (s *AuthService) ValidateAndRotateRefreshToken(ctx context.Context, token string) (string, string, error) {
	slog.Debug("Validating refresh token", "token_length", len(token))
	
	// Get the stored refresh token
	storedToken, err := s.refreshTokenRepo.Get(ctx, token)
	if err != nil {
		slog.Error("Failed to get refresh token from repository", "error", err)
		return "", "", err
	}
	if storedToken == nil {
		slog.Warn("Refresh token not found in database")
		return "", "", errors.New("invalid token")
	}

	slog.Debug("Found stored refresh token", "user_id", storedToken.UserID, "expires_at", storedToken.ExpiresAt)

	// Check if token is expired
	if time.Now().After(storedToken.ExpiresAt) {
		slog.Warn("Refresh token expired", "expires_at", storedToken.ExpiresAt)
		s.refreshTokenRepo.Delete(ctx, token)
		return "", "", errors.New("token expired")
	}

	// Delete the old token (consume it)
	err = s.refreshTokenRepo.Delete(ctx, token)
	if err != nil {
		slog.Error("Failed to delete old refresh token", "error", err)
		return "", "", err
	}

	// Generate a new refresh token
	newRefreshToken, err := generateRefreshToken()
	if err != nil {
		slog.Error("Failed to generate new refresh token", "error", err)
		return "", "", err
	}

	newRefreshID := uuid.NewString()

	// Store the new refresh token
	err = s.refreshTokenRepo.Store(ctx, &models.RefreshToken{
		ID:        newRefreshID,
		UserID:    storedToken.UserID,
		Token:     newRefreshToken,
		ExpiresAt: time.Now().Add(30 * 24 * time.Hour),
	})
	if err != nil {
		slog.Error("Failed to store new refresh token", "error", err)
		return "", "", err
	}

	slog.Info("Successfully rotated refresh token", "user_id", storedToken.UserID)
	return storedToken.UserID, newRefreshToken, nil
}

func (s *AuthService) BlacklistToken(ctx context.Context, jti string, expiresAt time.Time) error {
	return s.blacklistRepo.Add(ctx, jti, expiresAt)
}

func (s *AuthService) IsTokenBlacklisted(ctx context.Context, jti string) (bool, error) {
	return s.blacklistRepo.IsBlacklisted(ctx, jti)
}

func (s *AuthService) BlacklistAllForUser(ctx context.Context, userID string) error {
	return s.blacklistRepo.BlacklistAllForUser(ctx, userID)
}

func (s *AuthService) IsUserBlacklisted(ctx context.Context, userID string) (bool, error) {
	return s.blacklistRepo.IsUserBlacklisted(ctx, userID)
}

func (s *AuthService) UnblockUser(ctx context.Context, userID string) error {
	return s.blacklistRepo.UnblockUser(ctx, userID)
}
