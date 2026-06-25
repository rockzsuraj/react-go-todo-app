package auth

import (
	"crypto/rand"
	"encoding/base64"
)

// GenerateRefreshToken creates a cryptographically secure random token
// suitable for use as a refresh token. Returns a base64url-encoded string.
func GenerateRefreshToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
