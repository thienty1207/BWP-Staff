package security

import "testing"

func TestHashPasswordUsesArgon2idAndVerifies(t *testing.T) {
	hash, err := HashPassword("test-only-password")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	if len(hash) < 20 || hash[:10] != "$argon2id$" {
		t.Fatalf("expected Argon2id PHC hash, got %q", hash)
	}
	if !VerifyPassword("test-only-password", hash) {
		t.Fatal("expected original password to verify")
	}
	if VerifyPassword("wrong-password", hash) {
		t.Fatal("wrong password unexpectedly verified")
	}
}
