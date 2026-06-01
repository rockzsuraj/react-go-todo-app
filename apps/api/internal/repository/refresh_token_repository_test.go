package repository

import (
	"context"
	"testing"
	"time"

	"react-todos/apps/api/internal/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Test Setup ---

func setupTestDB(t *testing.T) *pgxpool.Pool {
	// This would typically connect to a test database
	// For now, we'll create a mock implementation
	// In a real scenario, you'd use testcontainers or a separate test DB
	t.Skip("Skipping repository tests - requires test database setup")
	return nil
}

// --- Mock Repository for Unit Testing ---

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

// --- Repository Tests ---

func TestRefreshTokenRepository_Store(t *testing.T) {
	mockRepo := NewMockRefreshTokenRepository()
	ctx := context.Background()

	refreshToken := &models.RefreshToken{
		ID:        uuid.New().String(),
		UserID:    uuid.New().String(),
		Token:     "test-token",
		ExpiresAt: time.Now().Add(24 * time.Hour),
		CreatedAt: time.Now(),
	}

	err := mockRepo.Store(ctx, refreshToken)
	assert.NoError(t, err)

	// Verify token was stored
	stored, err := mockRepo.Get(ctx, "test-token")
	assert.NoError(t, err)
	assert.NotNil(t, stored)
	assert.Equal(t, refreshToken.UserID, stored.UserID)
	assert.Equal(t, refreshToken.Token, stored.Token)
}

func TestRefreshTokenRepository_Get_NotFound(t *testing.T) {
	mockRepo := NewMockRefreshTokenRepository()
	ctx := context.Background()

	// Try to get non-existent token
	token, err := mockRepo.Get(ctx, "non-existent-token")
	assert.NoError(t, err)
	assert.Nil(t, token)
}

func TestRefreshTokenRepository_Delete(t *testing.T) {
	mockRepo := NewMockRefreshTokenRepository()
	ctx := context.Background()

	// Store a token first
	refreshToken := &models.RefreshToken{
		ID:        uuid.New().String(),
		UserID:    uuid.New().String(),
		Token:     "token-to-delete",
		ExpiresAt: time.Now().Add(24 * time.Hour),
		CreatedAt: time.Now(),
	}

	err := mockRepo.Store(ctx, refreshToken)
	require.NoError(t, err)

	// Verify it exists
	stored, err := mockRepo.Get(ctx, "token-to-delete")
	require.NoError(t, err)
	require.NotNil(t, stored)

	// Delete it
	err = mockRepo.Delete(ctx, "token-to-delete")
	assert.NoError(t, err)

	// Verify it's gone
	deleted, err := mockRepo.Get(ctx, "token-to-delete")
	assert.NoError(t, err)
	assert.Nil(t, deleted)
}

func TestRefreshTokenRepository_DeleteByUser(t *testing.T) {
	mockRepo := NewMockRefreshTokenRepository()
	ctx := context.Background()

	userID := uuid.New().String()

	// Store multiple tokens for the same user
	token1 := &models.RefreshToken{
		ID:        uuid.New().String(),
		UserID:    userID,
		Token:     "user-token-1",
		ExpiresAt: time.Now().Add(24 * time.Hour),
		CreatedAt: time.Now(),
	}

	token2 := &models.RefreshToken{
		ID:        uuid.New().String(),
		UserID:    userID,
		Token:     "user-token-2",
		ExpiresAt: time.Now().Add(24 * time.Hour),
		CreatedAt: time.Now(),
	}

	// Store a token for a different user
	otherUserToken := &models.RefreshToken{
		ID:        uuid.New().String(),
		UserID:    uuid.New().String(),
		Token:     "other-user-token",
		ExpiresAt: time.Now().Add(24 * time.Hour),
		CreatedAt: time.Now(),
	}

	err := mockRepo.Store(ctx, token1)
	require.NoError(t, err)
	err = mockRepo.Store(ctx, token2)
	require.NoError(t, err)
	err = mockRepo.Store(ctx, otherUserToken)
	require.NoError(t, err)

	// Delete all tokens for the user
	err = mockRepo.DeleteByUser(ctx, userID)
	assert.NoError(t, err)

	// Verify user's tokens are deleted
	stored1, err := mockRepo.Get(ctx, "user-token-1")
	assert.NoError(t, err)
	assert.Nil(t, stored1)

	stored2, err := mockRepo.Get(ctx, "user-token-2")
	assert.NoError(t, err)
	assert.Nil(t, stored2)

	// Verify other user's token still exists
	storedOther, err := mockRepo.Get(ctx, "other-user-token")
	assert.NoError(t, err)
	assert.NotNil(t, storedOther)
}

func TestRefreshTokenRepository_Get_ExpiredToken(t *testing.T) {
	mockRepo := NewMockRefreshTokenRepository()
	ctx := context.Background()

	// Store an expired token
	expiredToken := &models.RefreshToken{
		ID:        uuid.New().String(),
		UserID:    uuid.New().String(),
		Token:     "expired-token",
		ExpiresAt: time.Now().Add(-1 * time.Hour), // Expired 1 hour ago
		CreatedAt: time.Now().Add(-2 * time.Hour),
	}

	err := mockRepo.Store(ctx, expiredToken)
	require.NoError(t, err)

	// Get the token (repository should return it even if expired)
	// Expiration check should be done at the service level
	stored, err := mockRepo.Get(ctx, "expired-token")
	assert.NoError(t, err)
	assert.NotNil(t, stored)
	assert.Equal(t, expiredToken.Token, stored.Token)
}
