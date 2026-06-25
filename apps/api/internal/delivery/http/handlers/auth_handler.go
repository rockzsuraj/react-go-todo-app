package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	infraAuth "react-todos/apps/api/internal/infrastructure/auth"
	"react-todos/apps/api/internal/infrastructure/config"
	"react-todos/apps/api/internal/delivery/http/dto"
	"react-todos/apps/api/internal/delivery/http/middleware"
	"react-todos/apps/api/internal/infrastructure/repository"
	"react-todos/apps/api/internal/domain/models"
	"react-todos/apps/api/internal/domain/services"
	usecaseAuth "react-todos/apps/api/internal/usecase/auth"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// AuthHandler holds the dependencies for all authentication-related endpoints.
// Config is injected once at startup — no per-request os.Getenv calls.
type AuthHandler struct {
	service        services.AuthServicer
	cfg            config.AppConfig
	oauthStateRepo *repository.OAuthStateRepository
	googleVerifier *infraAuth.GoogleVerifier
}

// NewAuthHandler constructs an AuthHandler with all dependencies resolved at startup.
func NewAuthHandler(
	service services.AuthServicer,
	cfg config.AppConfig,
	stateRepo *repository.OAuthStateRepository,
) *AuthHandler {
	return &AuthHandler{
		service:        service,
		cfg:            cfg,
		oauthStateRepo: stateRepo,
		googleVerifier: infraAuth.NewGoogleVerifier(cfg.GoogleClientID),
	}
}

/* ===================== UTILS ===================== */

func generateOAuthState() (string, error) {
	// Reuses the same 32-byte cryptographically secure random generator.
	return infraAuth.GenerateRefreshToken()
}

/* ===================== GOOGLE LOGIN (Web only) ===================== */

func (h *AuthHandler) GoogleLogin(w http.ResponseWriter, r *http.Request) {
	oauthCfg := &oauth2.Config{
		ClientID:     h.cfg.GoogleClientID,
		ClientSecret: h.cfg.GoogleClientSecret,
		RedirectURL:  h.cfg.GoogleRedirectURL,
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}

	state, err := generateOAuthState()
	if err != nil {
		middleware.SendError(w, err)
		return
	}

	if h.oauthStateRepo == nil {
		middleware.SendJSONErrorWithCode(w, http.StatusServiceUnavailable, "ERR_SERVICE_UNAVAILABLE", "Auth service unavailable")
		return
	}

	if err := h.oauthStateRepo.Store(r.Context(), state, 10*time.Minute); err != nil {
		slog.Error("failed to store oauth state", "error", err)
		middleware.SendError(w, err)
		return
	}

	slog.Info("oauth state stored", "state_prefix", state[:8])
	authURL := oauthCfg.AuthCodeURL(state, oauth2.AccessTypeOffline)
	http.Redirect(w, r, authURL, http.StatusFound)
}

/* ===================== GOOGLE CALLBACK (Web only) ===================== */

func (h *AuthHandler) GoogleCallback(w http.ResponseWriter, r *http.Request) {
	if h.oauthStateRepo == nil {
		middleware.SendJSONErrorWithCode(w, http.StatusServiceUnavailable, "ERR_SERVICE_UNAVAILABLE", "Auth service unavailable")
		return
	}

	code := r.URL.Query().Get("code")
	if code == "" {
		middleware.SendJSONErrorWithCode(w, http.StatusUnauthorized, "ERR_MISSING_CODE", "Missing authorization code")
		return
	}

	state := r.URL.Query().Get("state")
	if state == "" {
		middleware.SendJSONErrorWithCode(w, http.StatusBadRequest, "ERR_MISSING_STATE", "Missing state parameter")
		return
	}

	ok, err := h.oauthStateRepo.Consume(r.Context(), state)
	if err != nil || !ok {
		prefix := state
		if len(prefix) > 8 {
			prefix = prefix[:8]
		}
		slog.Warn("invalid or expired oauth state", "state_prefix", prefix)
		middleware.SendJSONErrorWithCode(w, http.StatusUnauthorized, "ERR_INVALID_STATE", "Invalid or expired state")
		return
	}
	slog.Info("oauth state validated", "state_prefix", state[:8])

	accessToken, refreshToken, err := h.exchangeGoogleCode(r.Context(), code, h.cfg.GoogleRedirectURL)
	if err != nil {
		slog.Error("google code exchange failed", "error", err)
		middleware.SendJSONErrorWithCode(w, http.StatusUnauthorized, "ERR_OAUTH_EXCHANGE", "OAuth code exchange failed")
		return
	}

	h.setAuthCookies(w, accessToken, refreshToken)
	http.Redirect(w, r, h.cfg.FrontendURL+"/oauth/callback", http.StatusTemporaryRedirect)
}

/* ===================== MOBILE GOOGLE AUTH — Credential Manager (Android) ===================== */

// MobileGoogleAuth handles sign-in from the Android Credential Manager API.
// The Android app receives a Google ID token directly from the Credential Manager
// bottom sheet and POSTs it here. The backend verifies the token with Google,
// upserts the user, and returns our own JWT pair.
//
// Flow:
//
//	Android Credential Manager → Google ID token
//	  → POST /api/auth/mobile/google { "id_token": "..." }
//	  → backend verifies with Google tokeninfo endpoint
//	  → upsert user in DB
//	  → return { access_token, refresh_token }
func (h *AuthHandler) MobileGoogleAuth(w http.ResponseWriter, r *http.Request) {
	var req struct {
		IDToken string `json:"id_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.IDToken == "" {
		middleware.SendJSONErrorWithCode(w, http.StatusBadRequest, "ERR_INVALID_REQUEST", "id_token is required")
		return
	}

	// Verify the Google ID token against Google's tokeninfo endpoint.
	// Confirms the token is genuine, not expired, and issued for our client ID.
	claims, err := h.googleVerifier.Verify(r.Context(), req.IDToken)
	if err != nil {
		slog.Warn("google id token verification failed", "error", err, "ip", r.RemoteAddr)
		middleware.SendJSONErrorWithCode(w, http.StatusUnauthorized, "ERR_INVALID_TOKEN", "Invalid Google ID token")
		return
	}

	user, accessToken, refreshToken, err := h.issueTokensForGoogleUser(
		r.Context(),
		claims.Sub, claims.Email, claims.Name, claims.PictureURL,
	)
	if err != nil {
		slog.Error("failed to issue tokens for mobile google user", "error", err, "sub", claims.Sub)
		middleware.SendJSONErrorWithCode(w, http.StatusInternalServerError, "ERR_INTERNAL", "Failed to issue tokens")
		return
	}

	slog.Info("mobile google sign-in successful", "sub", claims.Sub, "email", claims.Email)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(dto.SuccessResponse(map[string]any{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"user":          dto.NewUserResponse(user),
	}))
}

/* ===================== SHARED: upsert user + issue JWT pair ===================== */

// issueTokensForGoogleUser upserts the Google user in the DB and issues
// a short-lived access JWT + long-lived refresh token. Used by both the
// web callback and the mobile Credential Manager endpoint.
func (h *AuthHandler) issueTokensForGoogleUser(
	ctx context.Context,
	googleUserID, email, name, picture string,
) (u *models.User, accessToken, refreshToken string, err error) {
	u, err = h.service.HandleGoogleLogin(ctx, googleUserID, email, name, picture)
	if err != nil {
		return nil, "", "", fmt.Errorf("upserting user: %w", err)
	}

	jti := uuid.NewString()
	claims := jwt.MapClaims{
		"sub":  u.ID,
		"jti":  jti,
		"role": u.Role,
		"iat":  time.Now().Unix(),
		"exp":  time.Now().Add(15 * time.Minute).Unix(),
	}
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	accessToken, err = jwtToken.SignedString([]byte(h.cfg.JWTSecret))
	if err != nil {
		return nil, "", "", fmt.Errorf("signing access token: %w", err)
	}

	refreshToken, err = infraAuth.GenerateRefreshToken()
	if err != nil {
		return nil, "", "", fmt.Errorf("generating refresh token: %w", err)
	}

	if err = h.service.StoreRefreshToken(ctx, uuid.NewString(), u.ID, refreshToken, time.Now().Add(30*24*time.Hour)); err != nil {
		return nil, "", "", fmt.Errorf("storing refresh token: %w", err)
	}

	return u, accessToken, refreshToken, nil
}

/* ===================== SHARED: OAuth code exchange (Web callback) ===================== */

func (h *AuthHandler) exchangeGoogleCode(ctx context.Context, code, redirectURI string, opts ...oauth2.AuthCodeOption) (accessToken, refreshToken string, err error) {
	oauthCfg := &oauth2.Config{
		ClientID:     h.cfg.GoogleClientID,
		ClientSecret: h.cfg.GoogleClientSecret,
		RedirectURL:  redirectURI,
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}

	tok, err := oauthCfg.Exchange(ctx, code, opts...)
	if err != nil {
		return "", "", err
	}

	resp, err := oauthCfg.Client(ctx, tok).Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	var g struct {
		ID      string `json:"id"`
		Email   string `json:"email"`
		Name    string `json:"name"`
		Picture string `json:"picture"`
	}
	if err = json.NewDecoder(resp.Body).Decode(&g); err != nil {
		return "", "", err
	}

	_, accessToken, refreshToken, err = h.issueTokensForGoogleUser(ctx, g.ID, g.Email, g.Name, g.Picture)
	return accessToken, refreshToken, err
}

/* ===================== SHARED: cookie helpers ===================== */

// setAuthCookies writes the access and refresh token as HttpOnly cookies.
func (h *AuthHandler) setAuthCookies(w http.ResponseWriter, accessToken, refreshToken string) {
	secure := h.cfg.Env == "production"
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    accessToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   15 * 60,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   30 * 24 * 60 * 60,
	})
}

// clearAuthCookies expires both auth cookies (used on logout / failed refresh).
func (h *AuthHandler) clearAuthCookies(w http.ResponseWriter) {
	secure := h.cfg.Env == "production"
	http.SetCookie(w, &http.Cookie{Name: "token", Value: "", Path: "/", HttpOnly: true, Secure: secure, SameSite: http.SameSiteLaxMode, MaxAge: -1})
	http.SetCookie(w, &http.Cookie{Name: "refresh_token", Value: "", Path: "/", HttpOnly: true, Secure: secure, SameSite: http.SameSiteLaxMode, MaxAge: -1})
}

/* ===================== AUTH ME ===================== */

func (h *AuthHandler) AuthMe(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	if userID == "" {
		middleware.SendJSONErrorWithCode(w, http.StatusUnauthorized, "ERR_UNAUTHORIZED", "Unauthorized")
		return
	}

	user, err := h.service.GetUserByID(r.Context(), userID)
	if err != nil {
		middleware.SendError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(dto.SuccessResponse(map[string]interface{}{
		"user": dto.NewUserResponse(user),
	}))
}

/* ===================== LOGOUT ===================== */

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	// Mobile: refresh token in X-Refresh-Token or Authorization header; Web: in cookie
	if c, err := r.Cookie("refresh_token"); err == nil {
		h.service.DeleteRefreshToken(r.Context(), c.Value)
	} else if xrf := r.Header.Get("X-Refresh-Token"); xrf != "" {
		h.service.DeleteRefreshToken(r.Context(), xrf)
	} else if hdr := r.Header.Get("Authorization"); strings.HasPrefix(hdr, "Bearer ") {
		h.service.DeleteRefreshToken(r.Context(), strings.TrimPrefix(hdr, "Bearer "))
	}

	// Blacklist the current access JWT by jti
	tokenStr := middleware.ExtractToken(r)
	if tokenStr != "" {
		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
			if t.Method != jwt.SigningMethodHS256 {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return []byte(h.cfg.JWTSecret), nil
		}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
		if err == nil && token.Valid {
			if claims, ok := token.Claims.(jwt.MapClaims); ok {
				if jti, ok := claims["jti"].(string); ok {
					if exp, ok := claims["exp"].(float64); ok {
						h.service.BlacklistToken(r.Context(), jti, time.Unix(int64(exp), 0))
					}
				}
			}
		}
	}

	// Clear web cookies (no-op on mobile)
	h.clearAuthCookies(w)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(dto.SuccessResponse(map[string]string{"message": "Logged out successfully"}))
}

/* ===================== REFRESH ===================== */

func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	slog.Debug("Processing refresh token request")

	var refreshToken string

	// Web: cookie
	if c, err := r.Cookie("refresh_token"); err == nil {
		refreshToken = c.Value
	}
	// Mobile: Authorization header
	if refreshToken == "" {
		if hdr := r.Header.Get("Authorization"); strings.HasPrefix(hdr, "Bearer ") {
			refreshToken = strings.TrimPrefix(hdr, "Bearer ")
		}
	}

	if refreshToken == "" {
		middleware.SendJSONErrorWithCode(w, http.StatusUnauthorized, "ERR_MISSING_TOKEN", "No refresh token provided")
		return
	}

	userID, newRefreshToken, err := h.service.ValidateAndRotateRefreshToken(r.Context(), refreshToken)
	if err != nil {
		slog.Error("refresh token validation failed", "error", err)
		errorCode, errorMessage := "ERR_INVALID_TOKEN", "Invalid refresh token"
		if errors.Is(err, usecaseAuth.ErrRefreshTokenExpired) {
			errorCode, errorMessage = "ERR_TOKEN_EXPIRED", "Refresh token has expired"
		}
		// Clear web cookies on failure
		h.clearAuthCookies(w)
		middleware.SendJSONErrorWithCode(w, http.StatusUnauthorized, errorCode, errorMessage)
		return
	}

	user, err := h.service.GetUserByID(r.Context(), userID)
	if err != nil {
		slog.Error("failed to get user by ID", "user_id", userID, "error", err)
		middleware.SendError(w, err)
		return
	}

	jti := uuid.NewString()
	claims := jwt.MapClaims{
		"sub":  userID,
		"jti":  jti,
		"role": user.Role,
		"iat":  time.Now().Unix(),
		"exp":  time.Now().Add(15 * time.Minute).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	accessToken, err := token.SignedString([]byte(h.cfg.JWTSecret))
	if err != nil {
		slog.Error("failed to sign access token", "error", err)
		middleware.SendError(w, err)
		return
	}

	// Web: set updated cookies — access token MaxAge must match the JWT expiry (15 min).
	// Refresh token retains its 30-day lifetime.
	if newRefreshToken != "" {
		h.setAuthCookies(w, accessToken, newRefreshToken)
	} else {
		// Rotation returned no new refresh token (shouldn't happen in normal flow,
		// but handled defensively — only refresh the access token cookie).
		secure := h.cfg.Env == "production"
		http.SetCookie(w, &http.Cookie{Name: "token", Value: accessToken, Path: "/", HttpOnly: true, Secure: secure, SameSite: http.SameSiteLaxMode, MaxAge: 15 * 60})
	}

	respData := map[string]string{"access_token": accessToken}
	if newRefreshToken != "" {
		respData["refresh_token"] = newRefreshToken
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(dto.SuccessResponse(respData)); err != nil {
		slog.Error("failed to encode refresh response", "error", err)
	}

	slog.Info("token refreshed successfully", "user_id", userID)
}
