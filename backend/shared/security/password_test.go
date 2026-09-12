package security

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"testing"
)

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

func TestVerifyPasswordRejectsUnsafeArgon2Parameters(t *testing.T) {
	salt := bytes.Repeat([]byte{1}, 16)
	key := bytes.Repeat([]byte{2}, 32)
	longSalt := bytes.Repeat([]byte{3}, 65)
	longKey := bytes.Repeat([]byte{4}, 33)

	tests := []struct {
		name        string
		memory      uint64
		iterations  uint64
		parallelism uint64
		salt        []byte
		key         []byte
	}{
		{name: "memory over policy cap", memory: 65537, iterations: 1, parallelism: 1, salt: salt, key: key},
		{name: "iterations over policy cap", memory: 1, iterations: 4, parallelism: 1, salt: salt, key: key},
		{name: "parallelism over policy cap", memory: 1, iterations: 1, parallelism: 5, salt: salt, key: key},
		{name: "key over policy cap", memory: 1, iterations: 1, parallelism: 1, salt: salt, key: longKey},
		{name: "salt over policy cap", memory: 1, iterations: 1, parallelism: 1, salt: longSalt, key: key},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			encoded := testPHC(test.memory, test.iterations, test.parallelism, test.salt, test.key)
			if VerifyPassword("test-only-password", encoded) {
				t.Fatal("unsafe Argon2 parameters unexpectedly verified")
			}
		})
	}
}

func TestVerifyPasswordRejectsMalformedPHC(t *testing.T) {
	for _, encoded := range []string{
		"",
		"not-a-phc-hash",
		"$argon2i$v=19$m=65536,t=3,p=4$AQ$Ag",
		"$argon2id$v=18$m=65536,t=3,p=4$AQ$Ag",
		"$argon2id$v=19$m=65536,t=3,p=4$%%%$Ag",
		"$argon2id$v=19$m=1,t=1,p=1,x=1$AQ$Ag",
		"$argon2id$v=19$m=1,m=1,t=1,p=1$AQ$Ag",
	} {
		if VerifyPassword("test-only-password", encoded) {
			t.Fatalf("malformed PHC unexpectedly verified: %q", encoded)
		}
	}
}

func testPHC(memory, iterations, parallelism uint64, salt, key []byte) string {
	return fmt.Sprintf(
		"$argon2id$v=19$m=%d,t=%d,p=%d$%s$%s",
		memory,
		iterations,
		parallelism,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(key),
	)
}
