package auth

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
)

// GetUserIDFromToken extracts the Cognito "sub" claim (user ID) from a JWT ID token
// without requiring an external JWT library.
func GetUserIDFromToken(idToken string) (string, error) {
	parts := strings.Split(idToken, ".")
	if len(parts) != 3 {
		return "", fmt.Errorf("invalid JWT format: expected 3 parts, got %d", len(parts))
	}

	// JWT payload is base64url-encoded without padding.
	payload := parts[1]

	// Add padding so that base64.URLEncoding can decode it.
	switch len(payload) % 4 {
	case 2:
		payload += "=="
	case 3:
		payload += "="
	}

	decoded, err := base64.URLEncoding.DecodeString(payload)
	if err != nil {
		return "", fmt.Errorf("failed to decode JWT payload: %w", err)
	}

	var claims map[string]interface{}
	if err := json.Unmarshal(decoded, &claims); err != nil {
		return "", fmt.Errorf("failed to parse JWT claims: %w", err)
	}

	sub, ok := claims["sub"].(string)
	if !ok || sub == "" {
		return "", fmt.Errorf("'sub' claim not found in JWT")
	}

	return sub, nil
}
