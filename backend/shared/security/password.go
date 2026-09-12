package security

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"

	"golang.org/x/crypto/argon2"
)

const (
	// MaxPasswordBytes is the shared input bound for password-bearing flows.
	MaxPasswordBytes = 1024

	argon2Memory         uint32 = 64 * 1024
	argon2Iterations     uint32 = 3
	argon2Parallelism    uint8  = 4
	argon2KeyLength      uint32 = 32
	argon2SaltLength            = 16
	argon2MaxMemory             = 64 * 1024
	argon2MaxIterations         = 3
	argon2MaxParallelism        = 4
	argon2MaxKeyLength          = 32
	argon2MaxSaltLength         = 64
)

// HashPassword returns an Argon2id PHC string suitable for password_hash.
func HashPassword(password string) (string, error) {
	salt := make([]byte, argon2SaltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("generate password salt: %w", err)
	}
	key := argon2.IDKey([]byte(password), salt, argon2Iterations, argon2Memory, argon2Parallelism, argon2KeyLength)
	encode := base64.RawStdEncoding.EncodeToString
	return fmt.Sprintf("$argon2id$v=19$m=%d,t=%d,p=%d$%s$%s", argon2Memory, argon2Iterations, argon2Parallelism, encode(salt), encode(key)), nil
}

// VerifyPassword verifies a password against an Argon2id PHC string.
func VerifyPassword(password, encoded string) bool {
	parts := strings.Split(encoded, "$")
	if len(parts) != 6 || parts[1] != "argon2id" || parts[2] != "v=19" {
		return false
	}
	parameters := make(map[string]uint64, 3)
	for _, item := range strings.Split(parts[3], ",") {
		pair := strings.SplitN(item, "=", 2)
		if len(pair) != 2 {
			return false
		}
		if pair[0] != "m" && pair[0] != "t" && pair[0] != "p" {
			return false
		}
		if _, exists := parameters[pair[0]]; exists {
			return false
		}
		value, err := strconv.ParseUint(pair[1], 10, 32)
		if err != nil {
			return false
		}
		parameters[pair[0]] = value
	}
	memory, okMemory := parameters["m"]
	iterations, okIterations := parameters["t"]
	parallelism, okParallelism := parameters["p"]
	if !okMemory || !okIterations || !okParallelism || memory == 0 || iterations == 0 || parallelism == 0 {
		return false
	}
	if memory > argon2MaxMemory || iterations > argon2MaxIterations || parallelism > argon2MaxParallelism {
		return false
	}
	decode := func(value string) ([]byte, error) {
		decoded, err := base64.RawStdEncoding.DecodeString(value)
		if err != nil {
			return base64.StdEncoding.DecodeString(value)
		}
		return decoded, nil
	}
	salt, err := decode(parts[4])
	if err != nil || len(salt) == 0 || len(salt) > argon2MaxSaltLength {
		return false
	}
	expected, err := decode(parts[5])
	if err != nil || len(expected) == 0 || len(expected) > argon2MaxKeyLength {
		return false
	}
	actual := argon2.IDKey([]byte(password), salt, uint32(iterations), uint32(memory), uint8(parallelism), uint32(len(expected)))
	return subtle.ConstantTimeCompare(actual, expected) == 1
}
