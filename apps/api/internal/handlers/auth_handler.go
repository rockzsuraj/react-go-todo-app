package handlers

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
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

/* ===================== GOOGLE CALLBACK ===================== */

func GoogleCallback(w http.ResponseWriter, r *http.Request) {
	appCfg := config.LoadAppConfig()

	// 🔍 Detect mobile client via explicit header set by the mobile app
	isMobile := r.Header.Get("X-Client-Type") == "mobile"
	slog.Info("callback request", "is_mobile", isMobile)

	stateCookie, err := r.Cookie(oauthStateCookieName)
	state := r.URL.Query().Get("state")
	if isMobile {
		// For mobile setups, ensure the state parameter exists at minimum
		if state == "" {
			middleware.SendJSONErrorWithCode(w, http.StatusUnauthorized, "ERR_INVALID_OAUTH_STATE", "Missing OAuth state parameter")
			return
		}
	} else {
		// Strict cookie validation remains intact for standard Web clients
		if err != nil || state == "" || subtle.ConstantTimeCompare([]byte(stateCookie.Value), []byte(state)) != 1 {
			middleware.SendJSONErrorWithCode(w, http.StatusUnauthorized, "ERR_INVALID_OAUTH_STATE", "Invalid OAuth state cookie")
			return
		}
	}

	if err == nil {
		http.SetCookie(w, &http.Cookie{
			Name:     oauthStateCookieName,
			Value:    "",
			Path:     "/api/auth/callback/google",
			HttpOnly: true,
			Secure:   appCfg.Env == "production",
			SameSite: http.SameSiteLaxMode,
			MaxAge:   -1,
		})
	}

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

	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	tok, err := oauthCfg.Exchange(context.Background(), code)
	if err != nil {
		middleware.SendError(w, err)
		return
	}

	client := oauthCfg.Client(context.Background(), tok)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		middleware.SendError(w, err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var g struct {
		ID      string `json:"id"`
		Email   string `json:"email"`
		Name    string `json:"name"`
		Picture string `json:"picture"`
	}

	if err := json.Unmarshal(body, &g); err != nil {
		middleware.SendError(w, err)
		return
	}

	// 🔥 Persist user
	user, err := authService.HandleGoogleLogin(
		r.Context(),
		g.ID,
		g.Email,
		g.Name,
		g.Picture,
	)
	if err != nil {
		middleware.SendError(w, err)
		return
	}

	// 🔐 Access token (7 days)
	jti := uuid.NewString()
	claims := jwt.MapClaims{
		"sub":  user.ID,
		"jti":  jti,
		"role": user.Role,
		"exp":  time.Now().Add(7 * 24 * time.Hour).Unix(),
	}

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	accessToken, err := jwtToken.SignedString([]byte(appCfg.JWTSecret))
	if err != nil {
		middleware.SendError(w, err)
		return
	}

	// 🔁 Refresh token (30 days)
	refreshToken, err := generateRefreshToken()
	if err != nil {
		middleware.SendError(w, err)
		return
	}

	refreshID := uuid.NewString()

	err = authService.StoreRefreshToken(
		r.Context(),
		refreshID,
		user.ID,
		refreshToken,
		time.Now().Add(30*24*time.Hour),
	)
	if err != nil {
		middleware.SendError(w, err)
		return
	}

	/* ===================== MOBILE: Redirect back to Android app ===================== */
	if isMobile {
		// Construct your custom application deep link address
		mobileRedirectURL := fmt.Sprintf("todoapp://oauth/callback?access_token=%s&refresh_token=%s", accessToken, refreshToken)
		http.Redirect(w, r, mobileRedirectURL, http.StatusTemporaryRedirect)
		return
	}

	/* ===================== WEB: set cookies on redirect, no tokens in URL ===================== */
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

	redirectURL := appCfg.FrontendURL + "/oauth/callback"
	http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
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
