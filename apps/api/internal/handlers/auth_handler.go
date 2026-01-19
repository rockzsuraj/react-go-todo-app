package handlers

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"react-todos/apps/api/internal/config"
	"react-todos/apps/api/internal/middleware"
	"react-todos/apps/api/internal/services"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var authService *services.AuthService

func InitAuthHandlers(service *services.AuthService) {
	authService = service
}

func generateRefreshToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// GoogleLogin redirects the user to Google's OAuth2 consent screen.
func GoogleLogin(w http.ResponseWriter, r *http.Request) {
	appCfg := config.LoadAppConfig()
	fmt.Printf("GoogleLogin: %v\n", appCfg)

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

	// Allow callers to provide a `redirect` query param (e.g. frontend URL).
	// We encode it into the OAuth `state` so it is returned to our callback.
	redirect := r.URL.Query().Get("redirect")
	state := ""
	if redirect != "" {
		state = url.QueryEscape(redirect)
	} else {
		state = "state"
	}

	authURL := oauthCfg.AuthCodeURL(state, oauth2.AccessTypeOffline)
	http.Redirect(w, r, authURL, http.StatusFound)
}

// GoogleCallback handles OAuth callback, persists user, returns JWT
func GoogleCallback(w http.ResponseWriter, r *http.Request) {
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

	code := r.URL.Query().Get("code")
	if code == "" {
		middleware.SendError(w, middleware.ErrUnauthorized)
		return
	}

	// Exchange code for token
	tok, err := oauthCfg.Exchange(context.Background(), code)
	if err != nil {
		middleware.SendError(w, err)
		return
	}

	// Fetch Google profile
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

	// 🔐 Create JWT
	claims := jwt.MapClaims{
		"sub":  user.ID,
		"role": user.Role,
		"exp":  time.Now().Add(15 * time.Minute).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(appCfg.JWTSecret))
	if err != nil {
		middleware.SendError(w, err)
		return
	}

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

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		Path:     "/api/auth/refresh",
		HttpOnly: true,
		Secure:   appCfg.Env == "production",
		SameSite: http.SameSiteStrictMode,
		MaxAge:   30 * 24 * 60 * 60,
	})

	// 🍪 SET COOKIE (web users)
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    signed,
		Path:     "/",
		HttpOnly: true,
		Secure:   appCfg.Env == "production",
		SameSite: http.SameSiteLaxMode,
		MaxAge:   15 * 60,
	})

	// 🔁 If browser login → redirect. Prefer an explicit `redirect` query
	// parameter; otherwise attempt to read redirect from the OAuth `state`.
	redirect := r.URL.Query().Get("redirect")
	if redirect == "" {
		// Try to use state (we encoded redirect as state in GoogleLogin)
		stateParam := r.URL.Query().Get("state")
		if stateParam != "" {
			if u, err := url.QueryUnescape(stateParam); err == nil {
				redirect = u
			}
		}
	}

	if redirect != "" {
		// Ensure redirect is a valid URL to avoid open-redirects.
		if parsed, err := url.Parse(redirect); err == nil && parsed.Scheme != "" {
			http.Redirect(
				w,
				r,
				redirect+"?token="+signed,
				http.StatusTemporaryRedirect,
			)
			return
		}
	}

	// 📦 Otherwise return JSON (API / SPA / mobile)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"token": signed,
		"user": map[string]any{
			"id":      user.ID,
			"email":   user.Email,
			"name":    user.Name,
			"picture": user.Picture,
		},
	})
}

func AuthMe(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	if userID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	// Fetch full user from service
	user, err := authService.GetUserByID(r.Context(), userID)
	if err != nil {
		middleware.SendError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"user": map[string]any{
			"id":        user.ID,
			"email":     user.Email,
			"name":      user.Name,
			"picture":   user.Picture,
			"role":      user.Role,
			"is_active": user.IsActive,
		},
	})
}

func Logout(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie("refresh_token"); err == nil {
		authService.DeleteRefreshToken(r.Context(), cookie.Value)
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/api/auth/refresh",
		HttpOnly: true,
		MaxAge:   -1,
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
	})

	w.WriteHeader(http.StatusOK)
}

func RefreshToken(w http.ResponseWriter, r *http.Request) {
	appCfg := config.LoadAppConfig()

	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	userID, err := authService.ValidateAndRotateRefreshToken(
		r.Context(),
		cookie.Value,
	)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	claims := jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(15 * time.Minute).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(appCfg.JWTSecret))
	if err != nil {
		middleware.SendError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"token": signed,
	})
}
