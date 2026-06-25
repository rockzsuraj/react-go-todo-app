package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// GoogleClaims contains the verified fields from a Google ID token.
type GoogleClaims struct {
	Sub        string // Google user ID — stable, unique per Google account
	Email      string
	Name       string
	PictureURL string
	IssuedAt   time.Time
}

// GoogleVerifier verifies Google ID tokens by calling Google's tokeninfo endpoint.
// This requires no extra dependencies — only the standard net/http client.
//
// For production with high traffic, consider caching Google's public certs and
// verifying locally. For a personal/demo project, the tokeninfo endpoint is simple
// and reliable.
type GoogleVerifier struct {
	clientID   string
	httpClient *http.Client
}

func NewGoogleVerifier(clientID string) *GoogleVerifier {
	return &GoogleVerifier{
		clientID:   clientID,
		httpClient: &http.Client{Timeout: 5 * time.Second},
	}
}

// Verify validates the ID token against Google's tokeninfo endpoint and returns
// the claims. The audience (aud) is validated against our client ID to prevent
// token substitution attacks.
func (g *GoogleVerifier) Verify(ctx context.Context, rawIDToken string) (*GoogleClaims, error) {
	endpoint := "https://oauth2.googleapis.com/tokeninfo?id_token=" + url.QueryEscape(rawIDToken)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("building tokeninfo request: %w", err)
	}

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("calling tokeninfo: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("invalid id token (tokeninfo status %d)", resp.StatusCode)
	}

	var info struct {
		Aud     string `json:"aud"`
		Sub     string `json:"sub"`
		Email   string `json:"email"`
		Name    string `json:"name"`
		Picture string `json:"picture"`
		Iat     string `json:"iat"`
		Exp     string `json:"exp"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, fmt.Errorf("decoding tokeninfo response: %w", err)
	}

	// Validate audience — must match our OAuth client ID
	if info.Aud != g.clientID {
		return nil, fmt.Errorf("id token audience mismatch: got %q, want %q", info.Aud, g.clientID)
	}

	if info.Sub == "" || info.Email == "" {
		return nil, fmt.Errorf("id token missing required claims (sub, email)")
	}

	return &GoogleClaims{
		Sub:        info.Sub,
		Email:      info.Email,
		Name:       info.Name,
		PictureURL: info.Picture,
	}, nil
}
