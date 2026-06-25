package auth

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"react-todos/apps/api/internal/domain/models"
	"react-todos/apps/api/internal/domain/repository"
	infraAuth "react-todos/apps/api/internal/infrastructure/auth"

	"github.com/google/uuid"
)

// Sentinel errors for token validation — use errors.Is() in callers,
// never compare .Error() strings.
var (
	ErrRefreshTokenNotFound = errors.New("refresh token not found")
	ErrRefreshTokenExpired  = errors.New("refresh token expired")
)

type AuthService struct {
	userRepo         repository.UserRepository
	refreshTokenRepo repository.RefreshTokenRepository
	blacklistRepo    repository.TokenBlacklistRepository
}

func NewAuthService(userRepo repository.UserRepository, refreshTokenRepo repository.RefreshTokenRepository, blacklistRepo repository.TokenBlacklistRepository) *AuthService {
	return &AuthService{
		userRepo:         userRepo,
		refreshTokenRepo: refreshTokenRepo,
		blacklistRepo:    blacklistRepo,
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

func (s *AuthService) ValidateAndRotateRefreshToken(ctx context.Context, token string) (string, string, error) {
	slog.Debug("Validating refresh token", "token_length", len(token))

	// ── 1. Fetch and validate the existing token ──────────────────────────
	storedToken, err := s.refreshTokenRepo.Get(ctx, token)
	if err != nil {
		slog.Error("Failed to get refresh token from repository", "error", err)
		return "", "", err
	}
	if storedToken == nil {
		slog.Warn("Refresh token not found in database")
		return "", "", ErrRefreshTokenNotFound
	}

	slog.Debug("Found stored refresh token", "user_id", storedToken.UserID, "expires_at", storedToken.ExpiresAt)

	if time.Now().After(storedToken.ExpiresAt) {
		slog.Warn("Refresh token expired", "expires_at", storedToken.ExpiresAt)
		// Best-effort cleanup — ignore error, token is already useless
		_ = s.refreshTokenRepo.Delete(ctx, token)
		return "", "", ErrRefreshTokenExpired
	}

	// ── 2. Generate and STORE the new token BEFORE deleting the old one ──
	// This is the safe order. If store fails, the old token is still valid
	// and the user can retry. Only after the new token is safely persisted
	// do we consume (delete) the old one.
	newRefreshToken, err := infraAuth.GenerateRefreshToken()
	if err != nil {
		slog.Error("Failed to generate new refresh token", "error", err)
		return "", "", err
	}

	err = s.refreshTokenRepo.Store(ctx, &models.RefreshToken{
		ID:        uuid.NewString(),
		UserID:    storedToken.UserID,
		Token:     newRefreshToken,
		ExpiresAt: time.Now().Add(30 * 24 * time.Hour),
	})
	if err != nil {
		slog.Error("Failed to store new refresh token — old token preserved", "error", err)
		return "", "", err
	}

	// ── 3. Consume the old token now that the new one is safely stored ────
	// If this delete fails, the user briefly has two valid tokens — that is
	// far better than having zero (permanent lockout). Log it but proceed.
	if err = s.refreshTokenRepo.Delete(ctx, token); err != nil {
		slog.Warn("Failed to delete old refresh token after rotation; token may be reused once", "error", err)
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
