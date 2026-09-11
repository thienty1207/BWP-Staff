package auth

import (
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/thienty1207/BWP-Staff/backend/shared/security"
)

func TestUnknownUserVerificationUsesValidDummyArgon2idHash(t *testing.T) {
	if !verifyLoginPassword(nil, "bwp-sonasea-dummy-password") {
		t.Fatal("expected unknown-user verification to use the fixed dummy Argon2id hash")
	}
}

func TestKnownUserVerificationUsesStoredArgon2idHash(t *testing.T) {
	hash, err := security.HashPassword("test-only-correct-password")
	if err != nil {
		t.Fatalf("hash test password: %v", err)
	}
	user := &userRecord{PasswordHash: hash}
	if !verifyLoginPassword(user, "test-only-correct-password") {
		t.Fatal("expected stored password to verify")
	}
	if verifyLoginPassword(user, "test-only-wrong-password") {
		t.Fatal("wrong password unexpectedly verified")
	}
}

func TestSessionCookieContract(t *testing.T) {
	expiresAt := time.Now().Add(12 * time.Hour)
	cookie := sessionCookie("test-only-token", expiresAt, false)
	if cookie.Name != sessionCookieName || cookie.Value != "test-only-token" {
		t.Fatal("session cookie has unexpected name or value")
	}
	if !cookie.HTTPOnly || cookie.Path != "/" || cookie.Domain != "" {
		t.Fatal("session cookie does not use the required host-only HttpOnly path contract")
	}
	if cookie.SameSite != fiber.CookieSameSiteLaxMode || cookie.Secure {
		t.Fatal("development session cookie has unexpected SameSite or Secure settings")
	}
	if cookie.MaxAge <= 0 || !cookie.Expires.After(time.Now()) {
		t.Fatal("session cookie does not carry a usable expiry")
	}

	productionCookie := sessionCookie("test-only-token", expiresAt, true)
	if !productionCookie.Secure {
		t.Fatal("non-development session cookie must be Secure")
	}
}

func TestLoginInputValidationPreservesPasswordWhitespace(t *testing.T) {
	input, err := validateLoginRequest(loginRequest{
		Username: "  hothienty  ",
		Password: " password-with-space ",
	})
	if err != nil {
		t.Fatalf("validate login request: %v", err)
	}
	if input.Username != "hothienty" {
		t.Fatalf("expected trimmed username, got %q", input.Username)
	}
	if input.Password != " password-with-space " {
		t.Fatal("password whitespace was modified")
	}

	for _, invalid := range []loginRequest{
		{Username: "", Password: "password"},
		{Username: strings.Repeat("u", maxUsernameCharacters+1), Password: "password"},
		{Username: "username", Password: ""},
		{Username: "username", Password: strings.Repeat("p", maxPasswordBytes+1)},
	} {
		if _, err := validateLoginRequest(invalid); err == nil {
			t.Fatal("expected invalid login request to be rejected")
		}
	}
}

func TestLoginInputValidationUsesUnicodeCharacterLimit(t *testing.T) {
	cases := []struct {
		name     string
		username string
		wantErr  bool
	}{
		{name: "50 ASCII characters accepted", username: strings.Repeat("u", 50)},
		{name: "51 ASCII characters rejected", username: strings.Repeat("u", 51), wantErr: true},
		{name: "50 Unicode characters accepted", username: strings.Repeat("é", 50)},
		{name: "51 Unicode characters rejected", username: strings.Repeat("é", 51), wantErr: true},
	}

	if len(cases[2].username) <= 50 {
		t.Fatal("test fixture must contain more than 50 UTF-8 bytes")
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			_, err := validateLoginRequest(loginRequest{
				Username: testCase.username,
				Password: "password",
			})
			if (err != nil) != testCase.wantErr {
				t.Fatalf("validateLoginRequest(%q) error=%v, wantErr=%t", testCase.username, err, testCase.wantErr)
			}
		})
	}
}
