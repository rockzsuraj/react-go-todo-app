package handlers

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"react-todos/apps/api/internal/config"
	"react-todos/apps/api/internal/dto"
	"react-todos/apps/api/internal/middleware"
	"react-todos/apps/api/internal/repository"
	"react-todos/apps/api/internal/services"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var authService services.AuthServicer
var oauthStateStore *repository.OAuthStateRepository

func InitAuthHandlers(service services.AuthServicer, stateRepo *repository.OAuthStateRepository) {
	authService = service
	oauthStateStore = stateRepo
}

/* ===================== UTILS ===================== */

func generateRefreshToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func generateOAuthState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

/* ===================== GOOGLE LOGIN ===================== */

func GoogleLogin(w http.ResponseWriter, r *http.Request) {
	appCfg := config.LoadAppConfig()

	oauthCfg := &oauth2.Config{
		ClientID:     appCfg.GoogleClientID,
		ClientSecret: appCfg.GoogleClientSecret,
		RedirectURL:  appCfg.GoogleRedirectURL,
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

	var authURL string

	// Mobile PKCE flow: if code_challenge is present, it's an Android app request.
	// We pass the challenge through to Google and skip the server-side state store
	// (state validation is replaced by the PKCE verifier check on token exchange).
	codeChallenge := r.URL.Query().Get("code_challenge")
	mobileRedirectURI := r.URL.Query().Get("redirect_uri")
	if codeChallenge != "" {
		challengeMethod := r.URL.Query().Get("code_challenge_method")
		if challengeMethod == "" {
			challengeMethod = "S256"
		}
		// For mobile, use the app's custom-scheme redirect URI so Google bounces back through our callback
		if mobileRedirectURI != "" {
			oauthCfg.RedirectURL = appCfg.GoogleRedirectURL // backend callback handles the relay
		}
		authURL = oauthCfg.AuthCodeURL(
			state,
			oauth2.AccessTypeOffline,
			oauth2.SetAuthURLParam("code_challenge", codeChallenge),
			oauth2.SetAuthURLParam("code_challenge_method", challengeMethod),
		)
	} else {
		// Web flow: store state for CSRF protection
		if oauthStateStore == nil {
			middleware.SendJSONErrorWithCode(w, http.StatusServiceUnavailable, "ERR_SERVICE_UNAVAILABLE", "Auth service unavailable")
			return
		}
		if err := oauthStateStore.Store(r.Context(), state, 10*time.Minute); err != nil {
			slog.Error("failed to store oauth state", "error", err)
			middleware.SendError(w, err)
			return
		}
		slog.Info("oauth state stored", "state_prefix", state[:8])
		authURL = oauthCfg.AuthCodeURL(state, oauth2.AccessTypeOffline)
	}

	http.Redirect(w, r, authURL, http.StatusFound)
}

/* ===================== GOOGLE CALLBACK (Web + Mobile) ===================== */

func GoogleCallback(w http.ResponseWriter, r *http.Request) {
	appCfg := config.LoadAppConfig()

	code := r.URL.Query().Get("code")
	if code == "" {
		middleware.SendJSONErrorWithCode(w, http.StatusUnauthorized, "ERR_MISSING_CODE", "Missing authorization code")
		return
	}

	state := r.URL.Query().Get("state")

	// Mobile PKCE flow: the app launched a browser pointing at /api/auth/google/login
	// with a code_challenge. Google redirects here. If the state is absent or not
	// found in our store, it came from the mobile path (no server-side state stored).
	isMobile := false
	if state == "" {
		isMobile = true
	} else {
		if oauthStateStore == nil {
			middleware.SendJSONErrorWithCode(w, http.StatusServiceUnavailable, "ERR_SERVICE_UNAVAILABLE", "Auth service unavailable")
			return
		}
		ok, err := oauthStateStore.Consume(r.Context(), state)
		if err != nil || !ok {
			// State not found — treat as mobile PKCE callback
			prefix := state
			if len(prefix) > 8 {
				prefix = prefix[:8]
			}
			slog.Warn("state not found in store, treating as mobile", "state_prefix", prefix)
			isMobile = true
		} else {
			slog.Info("oauth state validated", "state_prefix", state[:8])
		}
	}

	if isMobile {
		// PKCE browser flow: redirect back to the app with the code via custom scheme.
		// The app captures this deep link and POSTs code + code_verifier to
		// /api/auth/mobile/google for the actual token exchange.
		deepLink := fmt.Sprintf("%s?code=%s", appCfg.MobileRedirectURI, code)
		http.Redirect(w, r, deepLink, http.StatusTemporaryRedirect)
		return
	}

	accessToken, refreshToken, err := exchangeGoogleCode(r.Context(), appCfg, code, appCfg.GoogleRedirectURL)
	if err != nil {
		slog.Error("google code exchange failed", "error", err)
		middleware.SendJSONErrorWithCode(w, http.StatusUnauthorized, "ERR_OAUTH_EXCHANGE", "OAuth code exchange failed")
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    accessToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   appCfg.Env == "production",
		SameSite: http.SameSiteLaxMode,
		MaxAge:   15 * 60,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   appCfg.Env == "production",
		SameSite: http.SameSiteLaxMode,
		MaxAge:   30 * 24 * 60 * 60,
	})
	http.Redirect(w, r, appCfg.FrontendURL+"/oauth/callback", http.StatusTemporaryRedirect)
}

/* ===================== MOBILE GOOGLE AUTH (Android/iOS) ===================== */

// MobileGoogleAuth receives the authorization code from the Android app after
// the user completes the Google consent in a Chrome Custom Tab. The app captures
// the deep-link redirect (todoapp://oauth/callback?code=X), then POSTs the code
// here along with the PKCE code_verifier. Tokens are returned in the JSON body —
// never in a URL.
func MobileGoogleAuth(w http.ResponseWriter, r *http.Request) {
	appCfg := config.LoadAppConfig()

	var req struct {
		Code         string `json:"code"`
		CodeVerifier string `json:"code_verifier"`
		RedirectURI  string `json:"redirect_uri"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Code == "" {
		middleware.SendJSONErrorWithCode(w, http.StatusBadRequest, "ERR_INVALID_REQUEST", "code is required")
		return
	}

	// The Android client uses the custom-scheme URI registered in its manifest.
	// It must match exactly what was used when requesting the auth code.
	redirectURI := req.RedirectURI
	if redirectURI == "" {
		redirectURI = appCfg.MobileRedirectURI // e.g. "todoapp://oauth/callback"
	}

	accessToken, refreshToken, err := exchangeGoogleCode(r.Context(), appCfg, req.Code, redirectURI, oauth2.SetAuthURLParam("code_verifier", req.CodeVerifier))
	if err != nil {
		slog.Error("mobile google auth failed", "error", err)
		middleware.SendJSONErrorWithCode(w, http.StatusUnauthorized, "ERR_OAUTH_EXCHANGE", "OAuth code exchange failed")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(dto.SuccessResponse(map[string]string{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	}))
}

/* ===================== SHARED: code exchange + token issuance ===================== */

func exchangeGoogleCode(ctx context.Context, appCfg config.AppConfig, code, redirectURI string, opts ...oauth2.AuthCodeOption) (accessToken, refreshToken string, err error) {
	oauthCfg := &oauth2.Config{
		ClientID:     appCfg.GoogleClientID,
		ClientSecret: appCfg.GoogleClientSecret,
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

	u, err := authService.HandleGoogleLogin(ctx, g.ID, g.Email, g.Name, g.Picture)
	if err != nil {
		return "", "", err
	}

	jti := uuid.NewString()
	claims := jwt.MapClaims{
		"sub":  u.ID,
		"jti":  jti,
		"role": u.Role,
		"iat":  time.Now().Unix(),
		"exp":  time.Now().Add(15 * time.Minute).Unix(), // short-lived: refresh token handles renewal
	}
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	accessToken, err = jwtToken.SignedString([]byte(appCfg.JWTSecret))
	if err != nil {
		return "", "", err
	}

	refreshToken, err = generateRefreshToken()
	if err != nil {
		return "", "", err
	}

	err = authService.StoreRefreshToken(ctx, uuid.NewString(), u.ID, refreshToken, time.Now().Add(30*24*time.Hour))
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

/* ===================== AUTH ME ===================== */

func AuthMe(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	if userID == "" {
		middleware.SendJSONErrorWithCode(w, http.StatusUnauthorized, "ERR_UNAUTHORIZED", "Unauthorized")
		return
	}

	user, err := authService.GetUserByID(r.Context(), userID)
	if err != nil {
		middleware.SendError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	response := dto.SuccessResponse(map[string]interface{}{
		"user": dto.NewUserResponse(user),
	})
	_ = json.NewEncoder(w).Encode(response)
}

/* ===================== LOGOUT ===================== */

func Logout(w http.ResponseWriter, r *http.Request) {
	appCfg := config.LoadAppConfig()

	// Delete refresh token
	if cookie, err := r.Cookie("refresh_token"); err == nil {
		authService.DeleteRefreshToken(r.Context(), cookie.Value)
	}

	// Blacklist the current JWT (if present)
	tokenStr := middleware.ExtractToken(r)
	if tokenStr != "" {
		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
			if t.Method != jwt.SigningMethodHS256 {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return []byte(appCfg.JWTSecret), nil
		}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
		if err == nil && token.Valid {
			if claims, ok := token.Claims.(jwt.MapClaims); ok {
				if jti, ok := claims["jti"].(string); ok {
					if exp, ok := claims["exp"].(float64); ok {
						expiresAt := time.Unix(int64(exp), 0)
						authService.BlacklistToken(r.Context(), jti, expiresAt)
					}
				}
			}
		}
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   appCfg.Env == "production",
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   appCfg.Env == "production",
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	response := dto.SuccessResponse(map[string]string{"message": "Logged out successfully"})
	_ = json.NewEncoder(w).Encode(response)
}

/* ===================== REFRESH ===================== */

func RefreshToken(w http.ResponseWriter, r *http.Request) {
	appCfg := config.LoadAppConfig()
	slog.Debug("Processing refresh token request")

	var refreshToken string

	// 🌐 Web cookie
	if c, err := r.Cookie("refresh_token"); err == nil {
		refreshToken = c.Value
		slog.Debug("Found refresh token in cookie", "token_length", len(refreshToken))
	} else {
		slog.Debug("No refresh token cookie found")
	}

	// 📱 Mobile header
	if refreshToken == "" {
		if h := r.Header.Get("Authorization"); strings.HasPrefix(h, "Bearer ") {
			refreshToken = strings.TrimPrefix(h, "Bearer ")
			slog.Debug("Found refresh token in authorization header", "token_length", len(refreshToken))
		}
	}

	if refreshToken == "" {
		slog.Warn("No refresh token provided in request")
		middleware.SendJSONErrorWithCode(w, http.StatusUnauthorized, "ERR_MISSING_TOKEN", "No refresh token provided")
		return
	}

	// Validate and rotate refresh token
	userID, newRefreshToken, err := authService.ValidateAndRotateRefreshToken(
		r.Context(),
		refreshToken,
	)
	if err != nil {
		slog.Error("Refresh token validation failed", "error", err)
		errorCode := "ERR_INVALID_TOKEN"
		errorMessage := "Invalid refresh token"

		// Provide more specific error messages for common cases
		switch err.Error() {
		case "token expired":
			errorCode = "ERR_TOKEN_EXPIRED"
			errorMessage = "Refresh token has expired"
		case "invalid token":
			errorCode = "ERR_INVALID_TOKEN"
			errorMessage = "Invalid refresh token"
		}

		// Clear cookies on failure to prevent browser from sending bad credentials again
		http.SetCookie(w, &http.Cookie{
			Name:     "token",
			Value:    "",
			Path:     "/",
			HttpOnly: true,
			Secure:   appCfg.Env == "production",
			SameSite: http.SameSiteLaxMode,
			MaxAge:   -1,
		})
		http.SetCookie(w, &http.Cookie{
			Name:     "refresh_token",
			Value:    "",
			Path:     "/",
			HttpOnly: true,
			Secure:   appCfg.Env == "production",
			SameSite: http.SameSiteLaxMode,
			MaxAge:   -1,
		})

		middleware.SendJSONErrorWithCode(w, http.StatusUnauthorized, errorCode, errorMessage)
		return
	}

	user, err := authService.GetUserByID(r.Context(), userID)
	if err != nil {
		slog.Error("Failed to get user by ID", "user_id", userID, "error", err)
		middleware.SendError(w, err)
		return
	}

	// Generate new access token
	jti := uuid.NewString()
	claims := jwt.MapClaims{
		"sub":  userID,
		"jti":  jti,
		"role": user.Role,
		"iat":  time.Now().Unix(),
		"exp":  time.Now().Add(15 * time.Minute).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	accessToken, err := token.SignedString([]byte(appCfg.JWTSecret))
	if err != nil {
		slog.Error("Failed to sign access token", "error", err)
		middleware.SendError(w, err)
		return
	}

	// 🌐 Web: Set new access token cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    accessToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   appCfg.Env == "production",
		SameSite: http.SameSiteLaxMode,
		MaxAge:   7 * 24 * 60 * 60,
	})

	// 🌐 Web: Rotate refresh token cookie if a new one was generated
	if newRefreshToken != "" {
		http.SetCookie(w, &http.Cookie{
			Name:     "refresh_token",
			Value:    newRefreshToken,
			Path:     "/",
			HttpOnly: true,
			Secure:   appCfg.Env == "production",
			SameSite: http.SameSiteLaxMode,
			MaxAge:   30 * 24 * 60 * 60,
		})
		slog.Debug("Set new refresh token cookie")
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	respData := map[string]string{"access_token": accessToken}
	if newRefreshToken != "" {
		respData["refresh_token"] = newRefreshToken
	}
	if err := json.NewEncoder(w).Encode(dto.SuccessResponse(respData)); err != nil {
		slog.Error("Failed to encode refresh response", "error", err)
	}

	slog.Info("Successfully refreshed token", "user_id", userID)
}
