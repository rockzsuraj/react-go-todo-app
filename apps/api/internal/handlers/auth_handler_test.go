package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"react-todos/apps/api/internal/middleware"
	"react-todos/apps/api/internal/models"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// --- Mock AuthService ---

type MockAuthService struct {
	GetUserByIDFunc                    func(ctx context.Context, id string) (*models.User, error)
	StoreRefreshTokenFunc              func(ctx context.Context, refreshID, userID, token string, expiresAt time.Time) error
	DeleteRefreshTokenFunc             func(ctx context.Context, token string) error
	ValidateAndRotateRefreshTokenFunc  func(ctx context.Context, token string) (string, error)
}

func (m *MockAuthService) HandleGoogleLogin(ctx context.Context, googleUserID, email, name, picture string) (*models.User, error) {
	return &models.User{ID: "test-user"}, nil
}

func (m *MockAuthService) GetUserByID(ctx context.Context, id string) (*models.User, error) {
	if m.GetUserByIDFunc != nil {
		return m.GetUserByIDFunc(ctx, id)
	}
	return &models.User{ID: id}, nil
}

func (m *MockAuthService) StoreRefreshToken(ctx context.Context, refreshID, userID, token string, expiresAt time.Time) error {
	if m.StoreRefreshTokenFunc != nil {
		return m.StoreRefreshTokenFunc(ctx, refreshID, userID, token, expiresAt)
	}
	return nil
}

func (m *MockAuthService) DeleteRefreshToken(ctx context.Context, token string) error {
	if m.DeleteRefreshTokenFunc != nil {
		return m.DeleteRefreshTokenFunc(ctx, token)
	}
	return nil
}

func (m *MockAuthService) ValidateAndRotateRefreshToken(ctx context.Context, token string) (string, error) {
	if m.ValidateAndRotateRefreshTokenFunc != nil {
		return m.ValidateAndRotateRefreshTokenFunc(ctx, token)
	}
	return "user-123", nil
}

// --- Test Refresh Token Flow ---

func TestRefreshToken_ValidToken(t *testing.T) {
	userID := uuid.New().String()
	refreshToken := "valid-refresh-token"
	
	// Mock service
	mockService := &MockAuthService{
		ValidateAndRotateRefreshTokenFunc: func(ctx context.Context, token string) (string, error) {
			if token != refreshToken {
				t.Errorf("Expected token %s, got %s", refreshToken, token)
			}
			return userID, nil
		},
		GetUserByIDFunc: func(ctx context.Context, id string) (*models.User, error) {
			if id != userID {
				t.Errorf("Expected userID %s, got %s", userID, id)
			}
			return &models.User{
				ID:   userID,
				Role: "user",
			}, nil
		},
	}
	InitAuthHandlers(mockService)

	// Create request with refresh token cookie
	req, _ := http.NewRequest("POST", "/api/auth/refresh", nil)
	req.AddCookie(&http.Cookie{
		Name:  "refresh_token",
		Value: refreshToken,
	})

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(RefreshToken)
	handler.ServeHTTP(rr, req)

	// Assertions
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Check response contains new access token
	var response map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatal(err)
	}

	if response["access_token"] == "" {
		t.Error("Expected access_token in response")
	}

	// Verify the new access token is valid JWT
	token, err := jwt.Parse(response["access_token"], func(token *jwt.Token) (interface{}, error) {
		return []byte("test-secret"), nil // This would need to match actual JWT secret
	})
	if err != nil {
		t.Errorf("Invalid JWT token: %v", err)
	}

	if !token.Valid {
		t.Error("JWT token is not valid")
	}
}

func TestRefreshToken_MobileHeader(t *testing.T) {
	userID := uuid.New().String()
	refreshToken := "mobile-refresh-token"
	
	// Mock service
	mockService := &MockAuthService{
		ValidateAndRotateRefreshTokenFunc: func(ctx context.Context, token string) (string, error) {
			return userID, nil
		},
		GetUserByIDFunc: func(ctx context.Context, id string) (*models.User, error) {
			return &models.User{
				ID:   userID,
				Role: "user",
			}, nil
		},
	}
	InitAuthHandlers(mockService)

	// Create request with Authorization header (mobile flow)
	req, _ := http.NewRequest("POST", "/api/auth/refresh", nil)
	req.Header.Set("Authorization", "Bearer "+refreshToken)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(RefreshToken)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var response map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatal(err)
	}

	if response["access_token"] == "" {
		t.Error("Expected access_token in response")
	}
}

func TestRefreshToken_InvalidToken(t *testing.T) {
	// Mock service that returns error for invalid token
	mockService := &MockAuthService{
		ValidateAndRotateRefreshTokenFunc: func(ctx context.Context, token string) (string, error) {
			return "", jwt.ErrTokenUnverifiable
		},
	}
	InitAuthHandlers(mockService)

	// Create request with invalid refresh token
	req, _ := http.NewRequest("POST", "/api/auth/refresh", nil)
	req.AddCookie(&http.Cookie{
		Name:  "refresh_token",
		Value: "invalid-token",
	})

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(RefreshToken)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusUnauthorized {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusUnauthorized)
	}
}

func TestRefreshToken_NoToken(t *testing.T) {
	// Mock service (should not be called)
	mockService := &MockAuthService{
		ValidateAndRotateRefreshTokenFunc: func(ctx context.Context, token string) (string, error) {
			t.Fatal("ValidateAndRotateRefreshToken should not be called when no token provided")
			return "", nil
		},
	}
	InitAuthHandlers(mockService)

	// Create request with no token
	req, _ := http.NewRequest("POST", "/api/auth/refresh", nil)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(RefreshToken)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusUnauthorized {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusUnauthorized)
	}
}

func TestRefreshToken_ExpiredToken(t *testing.T) {
	// Mock service that simulates expired token
	mockService := &MockAuthService{
		ValidateAndRotateRefreshTokenFunc: func(ctx context.Context, token string) (string, error) {
			return "", jwt.ErrTokenExpired
		},
	}
	InitAuthHandlers(mockService)

	// Create request with expired refresh token
	req, _ := http.NewRequest("POST", "/api/auth/refresh", nil)
	req.AddCookie(&http.Cookie{
		Name:  "refresh_token",
		Value: "expired-token",
	})

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(RefreshToken)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusUnauthorized {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusUnauthorized)
	}
}

func TestLogout_WithRefreshToken(t *testing.T) {
	refreshToken := "test-refresh-token"
	
	// Mock service
	mockService := &MockAuthService{
		DeleteRefreshTokenFunc: func(ctx context.Context, token string) error {
			if token != refreshToken {
				t.Errorf("Expected token %s, got %s", refreshToken, token)
			}
			return nil
		},
	}
	InitAuthHandlers(mockService)

	// Create request with refresh token cookie
	req, _ := http.NewRequest("POST", "/api/auth/logout", nil)
	req.AddCookie(&http.Cookie{
		Name:  "refresh_token",
		Value: refreshToken,
	})

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(Logout)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Check that refresh token cookie is cleared
	cookies := rr.Result().Cookies()
	var refreshCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == "refresh_token" {
			refreshCookie = c
			break
		}
	}

	if refreshCookie == nil {
		t.Error("Expected refresh_token cookie to be cleared")
	} else if refreshCookie.MaxAge != -1 {
		t.Errorf("Expected refresh_token cookie MaxAge to be -1, got %d", refreshCookie.MaxAge)
	}

	// Check that access token cookie is cleared
	var accessCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == "token" {
			accessCookie = c
			break
		}
	}

	if accessCookie == nil {
		t.Error("Expected token cookie to be cleared")
	} else if accessCookie.MaxAge != -1 {
		t.Errorf("Expected token cookie MaxAge to be -1, got %d", accessCookie.MaxAge)
	}
}

func TestAuthMe_AuthenticatedUser(t *testing.T) {
	userID := uuid.New().String()
	expectedUser := &models.User{
		ID:       userID,
		Email:    "test@example.com",
		Name:     "Test User",
		Role:     "user",
		IsActive: true,
	}

	// Mock service
	mockService := &MockAuthService{
		GetUserByIDFunc: func(ctx context.Context, id string) (*models.User, error) {
			if id != userID {
				t.Errorf("Expected userID %s, got %s", userID, id)
			}
			return expectedUser, nil
		},
	}
	InitAuthHandlers(mockService)

	// Create request with user context
	req, _ := http.NewRequest("GET", "/api/auth/me", nil)
	ctx := middleware.WithUserID(req.Context(), userID)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(AuthMe)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatal(err)
	}

	user, ok := response["user"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected user object in response")
	}

	if user["ID"] != expectedUser.ID {
		t.Errorf("Expected user ID %s, got %v", expectedUser.ID, user["ID"])
	}

	if user["Email"] != expectedUser.Email {
		t.Errorf("Expected user email %s, got %v", expectedUser.Email, user["Email"])
	}
}

func TestAuthMe_UnauthenticatedUser(t *testing.T) {
	// Mock service (should not be called)
	mockService := &MockAuthService{
		GetUserByIDFunc: func(ctx context.Context, id string) (*models.User, error) {
			t.Fatal("GetUserByID should not be called for unauthenticated user")
			return nil, nil
		},
	}
	InitAuthHandlers(mockService)

	// Create request without user context
	req, _ := http.NewRequest("GET", "/api/auth/me", nil)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(AuthMe)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusUnauthorized {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusUnauthorized)
	}
}
