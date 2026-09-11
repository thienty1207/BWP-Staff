package security

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
)

const sessionTokenBytes = 32

// GenerateSessionToken creates a cryptographically random opaque token and its
// stable database representation.
func GenerateSessionToken() (string, string, error) {
	bytes := make([]byte, sessionTokenBytes)
	if _, err := rand.Read(bytes); err != nil {
		return "", "", fmt.Errorf("generate session token: %w", err)
	}
	rawToken := base64.RawURLEncoding.EncodeToString(bytes)
	return rawToken, HashSessionToken(rawToken), nil
}

// HashSessionToken returns the lowercase hexadecimal SHA-256 representation
// used in auth_sessions.session_token_hash.
func HashSessionToken(rawToken string) string {
	digest := sha256.Sum256([]byte(rawToken))
	return hex.EncodeToString(digest[:])
}

// IsValidSessionToken accepts only the canonical encoding emitted by
// GenerateSessionToken.
func IsValidSessionToken(rawToken string) bool {
	decoded, err := base64.RawURLEncoding.DecodeString(rawToken)
	return err == nil && len(decoded) == sessionTokenBytes && base64.RawURLEncoding.EncodeToString(decoded) == rawToken
}
