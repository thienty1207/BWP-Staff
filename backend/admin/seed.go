package admin

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/thienty1207/BWP-Staff/backend/config"
	"golang.org/x/crypto/argon2"
)

const (
	argon2Memory      uint32 = 64 * 1024
	argon2Iterations  uint32 = 3
	argon2Parallelism uint8  = 4
	argon2KeyLength   uint32 = 32
	argon2SaltLength         = 16
)

type SeedOutcome string

const (
	SeedCreated       SeedOutcome = "created"
	SeedAlreadyExists SeedOutcome = "already_present"
)

func HashPassword(password string) (string, error) {
	salt := make([]byte, argon2SaltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("generate password salt: %w", err)
	}
	key := argon2.IDKey([]byte(password), salt, argon2Iterations, argon2Memory, argon2Parallelism, argon2KeyLength)
	encode := base64.RawStdEncoding.EncodeToString
	return fmt.Sprintf("$argon2id$v=19$m=%d,t=%d,p=%d$%s$%s", argon2Memory, argon2Iterations, argon2Parallelism, encode(salt), encode(key)), nil
}

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
		value, err := strconv.ParseUint(pair[1], 10, 32)
		if err != nil {
			return false
		}
		parameters[pair[0]] = value
	}
	memory, okMemory := parameters["m"]
	iterations, okIterations := parameters["t"]
	parallelism, okParallelism := parameters["p"]
	if !okMemory || !okIterations || !okParallelism || memory == 0 || iterations == 0 || parallelism == 0 || parallelism > 255 {
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
	if err != nil || len(salt) == 0 {
		return false
	}
	expected, err := decode(parts[5])
	if err != nil || len(expected) == 0 {
		return false
	}
	actual := argon2.IDKey([]byte(password), salt, uint32(iterations), uint32(memory), uint8(parallelism), uint32(len(expected)))
	return subtle.ConstantTimeCompare(actual, expected) == 1
}

func SeedDevelopmentAdmin(ctx context.Context, pool *pgxpool.Pool, admin config.SeedAdmin) (SeedOutcome, error) {
	transaction, err := pool.Begin(ctx)
	if err != nil {
		return "", fmt.Errorf("begin development seed: %w", err)
	}

	var existingID int64
	err = transaction.QueryRow(ctx, "SELECT id FROM users WHERE username = $1", admin.Username).Scan(&existingID)
	if err == nil {
		if err := transaction.Commit(ctx); err != nil {
			return "", fmt.Errorf("finish existing development seed: %w", err)
		}
		return SeedAlreadyExists, nil
	}
	if err != pgx.ErrNoRows {
		_ = transaction.Rollback(ctx)
		return "", fmt.Errorf("check development admin: %w", err)
	}

	var departmentID int64
	err = transaction.QueryRow(ctx, "SELECT id FROM departments WHERE code = $1 AND is_active = TRUE", admin.DepartmentCode).Scan(&departmentID)
	if err == pgx.ErrNoRows {
		_ = transaction.Rollback(ctx)
		return "", fmt.Errorf("development department %q is missing", admin.DepartmentCode)
	}
	if err != nil {
		_ = transaction.Rollback(ctx)
		return "", fmt.Errorf("find development department: %w", err)
	}

	passwordHash, err := HashPassword(admin.Password)
	if err != nil {
		_ = transaction.Rollback(ctx)
		return "", err
	}

	err = transaction.QueryRow(ctx, `
INSERT INTO users (
    username,
    employee_code,
    email,
    password_hash,
    full_name,
    department_id,
    role
)
VALUES ($1, $2, NULL, $3, $4, $5, 'admin'::user_role)
ON CONFLICT (username) DO NOTHING
RETURNING id`,
		admin.Username,
		admin.EmployeeCode,
		passwordHash,
		admin.FullName,
		departmentID,
	).Scan(&existingID)
	if err == pgx.ErrNoRows {
		if rollbackErr := transaction.Rollback(ctx); rollbackErr != nil {
			return "", fmt.Errorf("rollback existing development seed: %w", rollbackErr)
		}
		return SeedAlreadyExists, nil
	}
	if err != nil {
		_ = transaction.Rollback(ctx)
		return "", fmt.Errorf("insert development admin: %w", err)
	}

	if _, err := transaction.Exec(ctx, `
INSERT INTO user_preferences (user_id)
VALUES ($1)
ON CONFLICT (user_id) DO NOTHING`, existingID); err != nil {
		_ = transaction.Rollback(ctx)
		return "", fmt.Errorf("insert development admin preferences: %w", err)
	}
	if err := transaction.Commit(ctx); err != nil {
		return "", fmt.Errorf("commit development seed: %w", err)
	}
	return SeedCreated, nil
}
