package handlers

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
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
	"react-todos/apps/api/internal/services"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var authService services.AuthServicer

const oauthStateCookieName = "oauth_state"

func InitAuthHandlers(service services.AuthServicer) {
	authService = service
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

	http.SetCookie(w, &http.Cookie{
		Name:     oauthStateCookieName,
		Value:    state,
		Path:     "/api/auth/callback/google",
		HttpOnly: true,
		Secure:   appCfg.Env == "production",
		SameSite: http.SameSiteLaxMode,
		MaxAge:   10 * 60,
	})

	authURL := oauthCfg.AuthCodeURL(state, oauth2.AccessTypeOffline)
	http.Redirect(w, r, authURL, http.StatusFound)
}

/* ===================== GOOGLE CALLBACK (Web only) ===================== */

func GoogleCallback(w http.ResponseWriter, r *http.Request) {
	appCfg := config.LoadAppConfig()

	stateCookie, err := r.Cookie(oauthStateCookieName)
	state := r.URL.Query().Get("state")
	if err != nil || state == "" || subtle.ConstantTimeCompare([]byte(stateCookie.Value), []byte(state)) != 1 {
		middleware.SendJSONErrorWithCode(w, http.StatusUnauthorized, "ERR_INVALID_OAUTH_STATE", "Invalid OAuth state")
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     oauthStateCookieName,
		Value:    "",
		Path:     "/api/auth/callback/google",
		HttpOnly: true,
		Secure:   appCfg.Env == "production",
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})

	code := r.URL.Query().Get("code")
	if code == "" {
		middleware.SendJSONErrorWithCode(w, http.StatusUnauthorized, "ERR_MISSING_CODE", "Missing authorization code")
		return
	}

	accessToken, refreshToken, user, err := exchangeGoogleCode(r.Context(), appCfg, code, appCfg.GoogleRedirectURL)
	if err != nil {
		middleware.SendError(w, err)
		return
	}

	secure := appCfg.Env == "production"
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    accessToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   7 * 24 * 60 * 60,
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
	_ = user

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

	accessToken, refreshToken, user, err := exchangeGoogleCode(r.Context(), appCfg, req.Code, redirectURI, oauth2.SetAuthURLParam("code_verifier", req.CodeVerifier))
	if err != nil {
		slog.Error("mobile google auth failed", "error", err)
		middleware.SendJSONErrorWithCode(w, http.StatusUnauthorized, "ERR_OAUTH_EXCHANGE", "OAuth code exchange failed")
		return
	}
	_ = user

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(dto.SuccessResponse(map[string]string{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	}))
}

/* ===================== SHARED: code exchange + token issuance ===================== */

func exchangeGoogleCode(ctx context.Context, appCfg config.AppConfig, code, redirectURI string, opts ...oauth2.AuthCodeOption) (accessToken, refreshToken string, user interface{}, err error) {
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
		return "", "", nil, err
	}

	resp, err := oauthCfg.Client(ctx, tok).Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return "", "", nil, err
	}
	defer resp.Body.Close()

	var g struct {
		ID      string `json:"id"`
		Email   string `json:"email"`
		Name    string `json:"name"`
		Picture string `json:"picture"`
	}
	if err = json.NewDecoder(resp.Body).Decode(&g); err != nil {
		return "", "", nil, err
	}

	u, err := authService.HandleGoogleLogin(ctx, g.ID, g.Email, g.Name, g.Picture)
	if err != nil {
		return "", "", nil, err
	}

	jti := uuid.NewString()
	claims := jwt.MapClaims{
		"sub":  u.ID,
		"jti":  jti,
		"role": u.Role,
		"exp":  time.Now().Add(7 * 24 * time.Hour).Unix(),
	}
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	accessToken, err = jwtToken.SignedString([]byte(appCfg.JWTSecret))
	if err != nil {
		return "", "", nil, err
	}

	refreshToken, err = generateRefreshToken()
	if err != nil {
		return "", "", nil, err
	}

	err = authService.StoreRefreshToken(ctx, uuid.NewString(), u.ID, refreshToken, time.Now().Add(30*24*time.Hour))
	if err != nil {
		return "", "", nil, err
	}

	return accessToken, refreshToken, u, nil
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
		"exp":  time.Now().Add(7 * 24 * time.Hour).Unix(),
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

	// 📱 Mobile response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := dto.SuccessResponse(map[string]string{"access_token": accessToken})
	if newRefreshToken != "" {
		response.Data.(map[string]string)["refresh_token"] = newRefreshToken
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		slog.Error("Failed to encode refresh response", "error", err)
	}

	slog.Info("Successfully refreshed token", "user_id", userID)
}
