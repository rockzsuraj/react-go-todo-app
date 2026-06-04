package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"react-todos/apps/api/internal/models"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Mock Repository ---

type MockUserRepository struct {
	users map[string]*models.User
}

func NewMockUserRepository() *MockUserRepository {
	return &MockUserRepository{
		users: make(map[string]*models.User),
	}
}

func (m *MockUserRepository) UpsertGoogleUser(ctx context.Context, googleUserID, email, name, picture string) (*models.User, error) {
	user := &models.User{
		ID:             uuid.New().String(),
		Provider:       "google",
		ProviderUserID: googleUserID,
		Email:          email,
		Name:           name,
		Picture:        picture,
		Role:           "user",
		IsActive:       true,
	}
	m.users[user.ID] = user
	return user, nil
}

func (m *MockUserRepository) GetByID(ctx context.Context, id string) (*models.User, error) {
	if user, exists := m.users[id]; exists {
		return user, nil
	}
	return nil, errors.New("user not found")
}

type MockRefreshTokenRepository struct {
	tokens map[string]*models.RefreshToken
}

func NewMockRefreshTokenRepository() *MockRefreshTokenRepository {
	return &MockRefreshTokenRepository{
		tokens: make(map[string]*models.RefreshToken),
	}
}

func (m *MockRefreshTokenRepository) Store(ctx context.Context, rt *models.RefreshToken) error {
	m.tokens[rt.Token] = rt
	return nil
}

func (m *MockRefreshTokenRepository) Get(ctx context.Context, token string) (*models.RefreshToken, error) {
	if rt, exists := m.tokens[token]; exists {
		return rt, nil
	}
	return nil, nil // Not found
}

func (m *MockRefreshTokenRepository) Delete(ctx context.Context, token string) error {
	delete(m.tokens, token)
	return nil
}

func (m *MockRefreshTokenRepository) DeleteByUser(ctx context.Context, userID string) error {
	for token, rt := range m.tokens {
		if rt.UserID == userID {
			delete(m.tokens, token)
		}
	}
	return nil
}

type MockTokenBlacklistRepository struct{}

func (m *MockTokenBlacklistRepository) IsBlacklisted(ctx context.Context, jti string) (bool, error) {
	return false, nil
}

func (m *MockTokenBlacklistRepository) Add(ctx context.Context, jti string, expiresAt time.Time) error {
	return nil
}

func (m *MockTokenBlacklistRepository) BlacklistAllForUser(ctx context.Context, userID string) error {
	return nil
}

func (m *MockTokenBlacklistRepository) IsUserBlacklisted(ctx context.Context, userID string) (bool, error) {
	return false, nil
}

func (m *MockTokenBlacklistRepository) UnblockUser(ctx context.Context, userID string) error {
	return nil
}

// --- Service Tests ---

func TestAuthService_StoreRefreshToken(t *testing.T) {
	userRepo := NewMockUserRepository()
	refreshRepo := NewMockRefreshTokenRepository()
	blacklistRepo := &MockTokenBlacklistRepository{}
	authService := NewAuthService(userRepo, refreshRepo, blacklistRepo)

	ctx := context.Background()
	refreshID := uuid.New().String()
	userID := uuid.New().String()
	token := "test-refresh-token"
	expiresAt := time.Now().Add(30 * 24 * time.Hour)

	err := authService.StoreRefreshToken(ctx, refreshID, userID, token, expiresAt)
	assert.NoError(t, err)

	// Verify token was stored
	stored, err := refreshRepo.Get(ctx, token)
	assert.NoError(t, err)
	assert.NotNil(t, stored)
	assert.Equal(t, userID, stored.UserID)
	assert.Equal(t, token, stored.Token)
	assert.Equal(t, expiresAt.Unix(), stored.ExpiresAt.Unix())
}

func TestAuthService_DeleteRefreshToken(t *testing.T) {
	userRepo := NewMockUserRepository()
	refreshRepo := NewMockRefreshTokenRepository()
	blacklistRepo := &MockTokenBlacklistRepository{}
	authService := NewAuthService(userRepo, refreshRepo, blacklistRepo)

	ctx := context.Background()
	token := "token-to-delete"

	// Store token first
	refreshToken := &models.RefreshToken{
		ID:        uuid.New().String(),
		UserID:    uuid.New().String(),
		Token:     token,
		ExpiresAt: time.Now().Add(24 * time.Hour),
		CreatedAt: time.Now(),
	}
	err := refreshRepo.Store(ctx, refreshToken)
	require.NoError(t, err)

	// Delete token
	err = authService.DeleteRefreshToken(ctx, token)
	assert.NoError(t, err)

	// Verify token is deleted
	stored, err := refreshRepo.Get(ctx, token)
	assert.NoError(t, err)
	assert.Nil(t, stored)
}

func TestAuthService_ValidateAndRotateRefreshToken_ValidToken(t *testing.T) {
	userRepo := NewMockUserRepository()
	refreshRepo := NewMockRefreshTokenRepository()
	blacklistRepo := &MockTokenBlacklistRepository{}
	authService := NewAuthService(userRepo, refreshRepo, blacklistRepo)

	ctx := context.Background()
	userID := uuid.New().String()
	token := "valid-token"

	// Store valid token
	refreshToken := &models.RefreshToken{
		ID:        uuid.New().String(),
		UserID:    userID,
		Token:     token,
		ExpiresAt: time.Now().Add(1 * time.Hour), // Not expired
		CreatedAt: time.Now(),
	}
	err := refreshRepo.Store(ctx, refreshToken)
	require.NoError(t, err)

	// Validate token
	resultUserID, newRotatedToken, err := authService.ValidateAndRotateRefreshToken(ctx, token)
	assert.NoError(t, err)
	assert.Equal(t, userID, resultUserID)
	assert.NotEmpty(t, newRotatedToken)
}

func TestAuthService_ValidateAndRotateRefreshToken_InvalidToken(t *testing.T) {
	userRepo := NewMockUserRepository()
	refreshRepo := NewMockRefreshTokenRepository()
	blacklistRepo := &MockTokenBlacklistRepository{}
	authService := NewAuthService(userRepo, refreshRepo, blacklistRepo)

	ctx := context.Background()

	// Try to validate non-existent token
	resultUserID, newRotatedToken, err := authService.ValidateAndRotateRefreshToken(ctx, "invalid-token")
	assert.Error(t, err)
	assert.Equal(t, "", resultUserID)
	assert.Equal(t, "", newRotatedToken)
	assert.Contains(t, err.Error(), "invalid token")
}

func TestAuthService_ValidateAndRotateRefreshToken_ExpiredToken(t *testing.T) {
	userRepo := NewMockUserRepository()
	refreshRepo := NewMockRefreshTokenRepository()
	blacklistRepo := &MockTokenBlacklistRepository{}
	authService := NewAuthService(userRepo, refreshRepo, blacklistRepo)

	ctx := context.Background()
	userID := uuid.New().String()
	token := "expired-token"

	// Store expired token
	refreshToken := &models.RefreshToken{
		ID:        uuid.New().String(),
		UserID:    userID,
		Token:     token,
		ExpiresAt: time.Now().Add(-1 * time.Hour), // Expired 1 hour ago
		CreatedAt: time.Now().Add(-2 * time.Hour),
	}
	err := refreshRepo.Store(ctx, refreshToken)
	require.NoError(t, err)

	// Validate token (should fail and delete expired token)
	resultUserID, newRotatedToken, err := authService.ValidateAndRotateRefreshToken(ctx, token)
	assert.Error(t, err)
	assert.Equal(t, "", resultUserID)
	assert.Equal(t, "", newRotatedToken)
	assert.Contains(t, err.Error(), "token expired")

	// Verify expired token was deleted
	stored, err := refreshRepo.Get(ctx, token)
	assert.NoError(t, err)
	assert.Nil(t, stored)
}

func TestAuthService_HandleGoogleLogin(t *testing.T) {
	userRepo := NewMockUserRepository()
	refreshRepo := NewMockRefreshTokenRepository()
	blacklistRepo := &MockTokenBlacklistRepository{}
	authService := NewAuthService(userRepo, refreshRepo, blacklistRepo)

	ctx := context.Background()
	googleUserID := "google-123"
	email := "test@example.com"
	name := "Test User"
	picture := "https://example.com/avatar.jpg"

	user, err := authService.HandleGoogleLogin(ctx, googleUserID, email, name, picture)
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, googleUserID, user.ProviderUserID)
	assert.Equal(t, email, user.Email)
	assert.Equal(t, name, user.Name)
	assert.Equal(t, picture, user.Picture)
	assert.Equal(t, "google", user.Provider)
	assert.Equal(t, "user", user.Role)
	assert.True(t, user.IsActive)
}

func TestAuthService_GetUserByID(t *testing.T) {
	userRepo := NewMockUserRepository()
	refreshRepo := NewMockRefreshTokenRepository()
	blacklistRepo := &MockTokenBlacklistRepository{}
	authService := NewAuthService(userRepo, refreshRepo, blacklistRepo)

	ctx := context.Background()

	// Create a user first
	user := &models.User{
		ID:             uuid.New().String(),
		Provider:       "google",
		ProviderUserID: "google-123",
		Email:          "test@example.com",
		Name:           "Test User",
		Role:           "user",
		IsActive:       true,
	}
	userRepo.users[user.ID] = user

	// Get user by ID
	retrievedUser, err := authService.GetUserByID(ctx, user.ID)
	assert.NoError(t, err)
	assert.NotNil(t, retrievedUser)
	assert.Equal(t, user.ID, retrievedUser.ID)
	assert.Equal(t, user.Email, retrievedUser.Email)

	// Try to get non-existent user
	_, err = authService.GetUserByID(ctx, uuid.New().String())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user not found")
}
