package admin

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/thienty1207/BWP-Staff/backend/config"
	"github.com/thienty1207/BWP-Staff/backend/shared/security"
)

type SeedOutcome string

const (
	SeedCreated       SeedOutcome = "created"
	SeedAlreadyExists SeedOutcome = "already_present"
)

func SeedDevelopmentAdmin(ctx context.Context, pool *pgxpool.Pool, admin config.SeedAdmin) (SeedOutcome, error) {
	if len(admin.Password) > security.MaxPasswordBytes {
		return "", fmt.Errorf("development admin password exceeds %d bytes", security.MaxPasswordBytes)
	}

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

	passwordHash, err := security.HashPassword(admin.Password)
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
