package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"react-todos/apps/api/internal/models"

	"github.com/golang-jwt/jwt/v5"
)

type authServiceStub struct{}

func (authServiceStub) HandleGoogleLogin(context.Context, string, string, string, string) (*models.User, error) {
	return nil, nil
}
func (authServiceStub) GetUserByID(context.Context, string) (*models.User, error) { return nil, nil }
func (authServiceStub) StoreRefreshToken(context.Context, string, string, string, time.Time) error {
	return nil
}
func (authServiceStub) DeleteRefreshToken(context.Context, string) error { return nil }
func (authServiceStub) ValidateAndRotateRefreshToken(context.Context, string) (string, string, error) {
	return "", "", nil
}
func (authServiceStub) BlacklistToken(context.Context, string, time.Time) error { return nil }
func (authServiceStub) IsTokenBlacklisted(context.Context, string) (bool, error) {
	return false, nil
}
func (authServiceStub) BlacklistAllForUser(context.Context, string) error { return nil }
func (authServiceStub) IsUserBlacklisted(context.Context, string) (bool, error) {
	return false, nil
}
func (authServiceStub) UnblockUser(context.Context, string) error { return nil }

type blacklistErrorAuthService struct {
	authServiceStub
}

func (blacklistErrorAuthService) IsTokenBlacklisted(context.Context, string) (bool, error) {
	return false, errors.New("redis unavailable")
}

func TestAuthMiddlewareRejectsNonHS256Tokens(t *testing.T) {
	secret := "secret"
	token := jwt.NewWithClaims(jwt.SigningMethodHS384, jwt.MapClaims{
		"sub": "user-1",
		"exp": time.Now().Add(time.Hour).Unix(),
	})
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatal(err)
	}

	nextCalled := false
	handler := AuthMiddleware(secret, authServiceStub{})(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		nextCalled = true
	}))
	req := httptest.NewRequest("GET", "/api/auth/me", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if nextCalled {
		t.Fatal("expected non-HS256 token to be rejected")
	}
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

func TestAPIKeyHasNoInsecureFallback(t *testing.T) {
	t.Setenv("API_KEYS", "")

	if isValidAPIKey("dev-key-12345") {
		t.Fatal("expected development fallback API key to be rejected")
	}
}

func TestAuthMiddlewareFailsClosedWhenBlacklistCheckFails(t *testing.T) {
	secret := "secret"
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": "user-1",
		"jti": "token-1",
		"exp": time.Now().Add(time.Hour).Unix(),
	})
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatal(err)
	}

	handler := AuthMiddleware(secret, blacklistErrorAuthService{})(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Fatal("expected request to be rejected when Redis is unavailable")
	}))
	req := httptest.NewRequest("GET", "/api/auth/me", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}
