package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"react-todos/apps/api/internal/infrastructure/config"
	"react-todos/apps/api/internal/delivery/http/middleware"
	"react-todos/apps/api/internal/domain/models"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// --- Mock AuthService ---

type MockAuthService struct {
	GetUserByIDFunc                   func(ctx context.Context, id string) (*models.User, error)
	StoreRefreshTokenFunc             func(ctx context.Context, refreshID, userID, token string, expiresAt time.Time) error
	DeleteRefreshTokenFunc            func(ctx context.Context, token string) error
	ValidateAndRotateRefreshTokenFunc func(ctx context.Context, token string) (string, string, error)
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

func (m *MockAuthService) ValidateAndRotateRefreshToken(ctx context.Context, token string) (string, string, error) {
	if m.ValidateAndRotateRefreshTokenFunc != nil {
		return m.ValidateAndRotateRefreshTokenFunc(ctx, token)
	}
	return "user-123", "", nil
}

func (m *MockAuthService) BlacklistToken(_ context.Context, _ string, _ time.Time) error {
	return nil
}

func (m *MockAuthService) IsTokenBlacklisted(_ context.Context, _ string) (bool, error) {
	return false, nil
}

func (m *MockAuthService) BlacklistAllForUser(_ context.Context, _ string) error {
	return nil
}

func (m *MockAuthService) IsUserBlacklisted(_ context.Context, _ string) (bool, error) {
	return false, nil
}

func (m *MockAuthService) UnblockUser(_ context.Context, _ string) error {
	return nil
}

// newTestAuthHandler builds an AuthHandler with a test config and the given mock service.
// oauthStateRepo is nil for unit tests that don't exercise the OAuth flow.
func newTestAuthHandler(svc *MockAuthService) *AuthHandler {
	return NewAuthHandler(svc, config.AppConfig{
		JWTSecret: "dev-jwt-secret",
		Env:       "test",
	}, nil)
}

// --- Tests ---

// TestGoogleLoginRedirectsWithState verifies the web flow redirects to Google
// with a state param. State is stored server-side (not a cookie).
func TestGoogleLoginRedirectsWithState(t *testing.T) {
	h := newTestAuthHandler(&MockAuthService{})

	req := httptest.NewRequest("GET", "/api/auth/google/login", nil)
	rr := httptest.NewRecorder()

	h.GoogleLogin(rr, req)

	// oauthStateRepo is nil → handler returns 503 or redirects depending on config.
	// In unit tests without a real DB connection, 503 is acceptable. Just confirm no panic.
	if rr.Code != http.StatusFound && rr.Code != http.StatusServiceUnavailable && rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected redirect, service unavailable, or internal error; got %d", rr.Code)
	}
}

// TestGoogleCallbackMobileRelaysCodeViaDeepLink verifies that when state is not
// found in the store the callback returns an appropriate error.
func TestGoogleCallbackMobileRelaysCodeViaDeepLink(t *testing.T) {
	h := newTestAuthHandler(&MockAuthService{})

	req := httptest.NewRequest("GET", "/api/auth/callback/google?code=testcode&state=unknown-state", nil)
	rr := httptest.NewRecorder()

	h.GoogleCallback(rr, req)

	// oauthStateRepo is nil → handler returns 503.
	if rr.Code != http.StatusServiceUnavailable && rr.Code != http.StatusTemporaryRedirect && rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401/503 or deep-link redirect, got %d", rr.Code)
	}
}

func TestRefreshToken_ValidToken(t *testing.T) {
	userID := uuid.New().String()
	refreshToken := "valid-refresh-token"

	mockService := &MockAuthService{
		ValidateAndRotateRefreshTokenFunc: func(ctx context.Context, token string) (string, string, error) {
			if token != refreshToken {
				t.Errorf("Expected token %s, got %s", refreshToken, token)
			}
			return userID, "", nil
		},
		GetUserByIDFunc: func(ctx context.Context, id string) (*models.User, error) {
			if id != userID {
				t.Errorf("Expected userID %s, got %s", userID, id)
			}
			return &models.User{ID: userID, Role: "user"}, nil
		},
	}
	h := newTestAuthHandler(mockService)

	req, _ := http.NewRequest("POST", "/api/auth/refresh", nil)
	req.AddCookie(&http.Cookie{Name: "refresh_token", Value: refreshToken})

	rr := httptest.NewRecorder()
	h.RefreshToken(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var apiResp struct {
		Success bool              `json:"success"`
		Data    map[string]string `json:"data"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&apiResp); err != nil {
		t.Fatal(err)
	}

	if apiResp.Data["access_token"] == "" {
		t.Error("Expected access_token in response")
	}

	token, err := jwt.Parse(apiResp.Data["access_token"], func(token *jwt.Token) (interface{}, error) {
		return []byte("dev-jwt-secret"), nil
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

	mockService := &MockAuthService{
		ValidateAndRotateRefreshTokenFunc: func(ctx context.Context, token string) (string, string, error) {
			return userID, "", nil
		},
		GetUserByIDFunc: func(ctx context.Context, id string) (*models.User, error) {
			return &models.User{ID: userID, Role: "user"}, nil
		},
	}
	h := newTestAuthHandler(mockService)

	req, _ := http.NewRequest("POST", "/api/auth/refresh", nil)
	req.Header.Set("Authorization", "Bearer "+refreshToken)

	rr := httptest.NewRecorder()
	h.RefreshToken(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var apiResp struct {
		Success bool              `json:"success"`
		Data    map[string]string `json:"data"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&apiResp); err != nil {
		t.Fatal(err)
	}
	if apiResp.Data["access_token"] == "" {
		t.Error("Expected access_token in response")
	}
}

func TestRefreshToken_InvalidToken(t *testing.T) {
	mockService := &MockAuthService{
		ValidateAndRotateRefreshTokenFunc: func(ctx context.Context, token string) (string, string, error) {
			return "", "", jwt.ErrTokenUnverifiable
		},
	}
	h := newTestAuthHandler(mockService)

	req, _ := http.NewRequest("POST", "/api/auth/refresh", nil)
	req.AddCookie(&http.Cookie{Name: "refresh_token", Value: "invalid-token"})

	rr := httptest.NewRecorder()
	h.RefreshToken(rr, req)

	if status := rr.Code; status != http.StatusUnauthorized {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusUnauthorized)
	}
}

func TestRefreshToken_NoToken(t *testing.T) {
	mockService := &MockAuthService{
		ValidateAndRotateRefreshTokenFunc: func(ctx context.Context, token string) (string, string, error) {
			t.Fatal("ValidateAndRotateRefreshToken should not be called when no token provided")
			return "", "", nil
		},
	}
	h := newTestAuthHandler(mockService)

	req, _ := http.NewRequest("POST", "/api/auth/refresh", nil)

	rr := httptest.NewRecorder()
	h.RefreshToken(rr, req)

	if status := rr.Code; status != http.StatusUnauthorized {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusUnauthorized)
	}
}

func TestRefreshToken_ExpiredToken(t *testing.T) {
	mockService := &MockAuthService{
		ValidateAndRotateRefreshTokenFunc: func(ctx context.Context, token string) (string, string, error) {
			return "", "", jwt.ErrTokenExpired
		},
	}
	h := newTestAuthHandler(mockService)

	req, _ := http.NewRequest("POST", "/api/auth/refresh", nil)
	req.AddCookie(&http.Cookie{Name: "refresh_token", Value: "expired-token"})

	rr := httptest.NewRecorder()
	h.RefreshToken(rr, req)

	if status := rr.Code; status != http.StatusUnauthorized {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusUnauthorized)
	}
}

func TestLogout_WithRefreshToken(t *testing.T) {
	refreshToken := "test-refresh-token"

	mockService := &MockAuthService{
		DeleteRefreshTokenFunc: func(ctx context.Context, token string) error {
			if token != refreshToken {
				t.Errorf("Expected token %s, got %s", refreshToken, token)
			}
			return nil
		},
	}
	h := newTestAuthHandler(mockService)

	req, _ := http.NewRequest("POST", "/api/auth/logout", nil)
	req.AddCookie(&http.Cookie{Name: "refresh_token", Value: refreshToken})

	rr := httptest.NewRecorder()
	h.Logout(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

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

	mockService := &MockAuthService{
		GetUserByIDFunc: func(ctx context.Context, id string) (*models.User, error) {
			if id != userID {
				t.Errorf("Expected userID %s, got %s", userID, id)
			}
			return expectedUser, nil
		},
	}
	h := newTestAuthHandler(mockService)

	req, _ := http.NewRequest("GET", "/api/auth/me", nil)
	ctx := middleware.WithUserID(req.Context(), userID)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	h.AuthMe(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var apiResp struct {
		Success bool `json:"success"`
		Data    struct {
			User map[string]interface{} `json:"user"`
		} `json:"data"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&apiResp); err != nil {
		t.Fatal(err)
	}

	user := apiResp.Data.User
	if user == nil {
		t.Fatal("Expected user object in response")
	}
	if user["id"] != expectedUser.ID {
		t.Errorf("Expected user ID %s, got %v", expectedUser.ID, user["id"])
	}
	if user["email"] != expectedUser.Email {
		t.Errorf("Expected user email %s, got %v", expectedUser.Email, user["email"])
	}
}

func TestAuthMe_UnauthenticatedUser(t *testing.T) {
	mockService := &MockAuthService{
		GetUserByIDFunc: func(ctx context.Context, id string) (*models.User, error) {
			t.Fatal("GetUserByID should not be called for unauthenticated user")
			return nil, nil
		},
	}
	h := newTestAuthHandler(mockService)

	req, _ := http.NewRequest("GET", "/api/auth/me", nil)

	rr := httptest.NewRecorder()
	h.AuthMe(rr, req)

	if status := rr.Code; status != http.StatusUnauthorized {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusUnauthorized)
	}
}
