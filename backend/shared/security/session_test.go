package security

import (
	"encoding/base64"
	"strings"
	"testing"
)

func TestGenerateSessionTokenReturnsOpaqueURLSafeTokenAndHash(t *testing.T) {
	rawToken, tokenHash, err := GenerateSessionToken()
	if err != nil {
		t.Fatalf("generate session token: %v", err)
	}
	if rawToken == "" || tokenHash == "" {
		t.Fatal("expected non-empty raw token and token hash")
	}
	if rawToken == tokenHash {
		t.Fatal("raw session token must differ from stored hash")
	}
	if strings.ContainsAny(rawToken, "+/=") {
		t.Fatal("raw session token is not URL-safe")
	}
	decoded, err := base64.RawURLEncoding.DecodeString(rawToken)
	if err != nil {
		t.Fatalf("decode raw session token: %v", err)
	}
	if len(decoded) != 32 {
		t.Fatalf("unexpected raw session token entropy: %d bytes", len(decoded))
	}
}

func TestSessionTokenHashIsDeterministic(t *testing.T) {
	const rawToken = "test-only-session-token"
	first := HashSessionToken(rawToken)
	second := HashSessionToken(rawToken)
	if first == "" || first != second {
		t.Fatal("same raw session token did not produce a stable hash")
	}
	if first == HashSessionToken("another-test-only-session-token") {
		t.Fatal("different raw session tokens produced the same hash")
	}
}

func TestGenerateSessionTokenProducesDifferentTokens(t *testing.T) {
	first, _, err := GenerateSessionToken()
	if err != nil {
		t.Fatalf("generate first session token: %v", err)
	}
	second, _, err := GenerateSessionToken()
	if err != nil {
		t.Fatalf("generate second session token: %v", err)
	}
	if first == second {
		t.Fatal("independent session token generations matched")
	}
}
