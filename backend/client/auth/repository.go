package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (repository *Repository) FindUserByUsername(ctx context.Context, username string) (userRecord, error) {
	if repository == nil || repository.pool == nil {
		return userRecord{}, fmt.Errorf("auth repository is not configured")
	}

	var user userRecord
	err := repository.pool.QueryRow(ctx, `
SELECT
    u.id,
    u.username,
    u.employee_code,
    u.password_hash,
    u.full_name,
    u.role::text,
    u.avatar_url,
    u.is_active,
    d.id,
    d.code,
    d.name
FROM users AS u
JOIN departments AS d ON d.id = u.department_id
WHERE u.username = $1
LIMIT 1`, username).Scan(
		&user.ID,
		&user.Username,
		&user.EmployeeCode,
		&user.PasswordHash,
		&user.FullName,
		&user.Role,
		&user.AvatarURL,
		&user.IsActive,
		&user.Department.ID,
		&user.Department.Code,
		&user.Department.Name,
	)
	if err != nil {
		return userRecord{}, err
	}
	return user, nil
}

func (repository *Repository) CreateSession(ctx context.Context, session sessionRecord) (int64, error) {
	if repository == nil || repository.pool == nil {
		return 0, fmt.Errorf("auth repository is not configured")
	}

	transaction, err := repository.pool.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("begin auth session transaction: %w", err)
	}
	defer func() { _ = transaction.Rollback(ctx) }()

	var userAgent any
	if session.UserAgent != nil {
		userAgent = *session.UserAgent
	}
	var ipAddress any
	if session.IPAddress != nil {
		ipAddress = *session.IPAddress
	}

	var sessionID int64
	if err := transaction.QueryRow(ctx, `
INSERT INTO auth_sessions (
    session_token_hash,
    user_id,
    expires_at,
    last_used_at,
    user_agent,
    ip_address,
    created_at
)
VALUES ($1, $2, $3, NULL, $4, $5::inet, $6)
RETURNING id`,
		session.TokenHash,
		session.UserID,
		session.ExpiresAt,
		userAgent,
		ipAddress,
		session.CreatedAt,
	).Scan(&sessionID); err != nil {
		return 0, fmt.Errorf("insert auth session: %w", err)
	}

	result, err := transaction.Exec(ctx, `
UPDATE users
SET last_login_at = $1
WHERE id = $2
  AND is_active = TRUE`, session.CreatedAt, session.UserID)
	if err != nil {
		return 0, fmt.Errorf("update last login time: %w", err)
	}
	if result.RowsAffected() != 1 {
		return 0, errUserNotActive
	}

	if err := transaction.Commit(ctx); err != nil {
		return 0, fmt.Errorf("commit auth session transaction: %w", err)
	}
	return sessionID, nil
}

func (repository *Repository) FindValidSession(ctx context.Context, tokenHash string) (Principal, error) {
	if repository == nil || repository.pool == nil {
		return Principal{}, fmt.Errorf("auth repository is not configured")
	}

	var (
		principal Principal
		user      userRecord
		expiresAt time.Time
	)
	err := repository.pool.QueryRow(ctx, `
SELECT
    s.id,
    s.expires_at,
    u.id,
    u.username,
    u.employee_code,
    u.full_name,
    u.role::text,
    u.avatar_url,
    d.id,
    d.code,
    d.name
FROM auth_sessions AS s
JOIN users AS u ON u.id = s.user_id
JOIN departments AS d ON d.id = u.department_id
WHERE s.session_token_hash = $1
  AND s.revoked_at IS NULL
  AND s.expires_at > NOW()
  AND u.is_active = TRUE
LIMIT 1`, tokenHash).Scan(
		&principal.SessionID,
		&expiresAt,
		&user.ID,
		&user.Username,
		&user.EmployeeCode,
		&user.FullName,
		&user.Role,
		&user.AvatarURL,
		&user.Department.ID,
		&user.Department.Code,
		&user.Department.Name,
	)
	if err != nil {
		return Principal{}, err
	}
	principal.User = identityFromUser(user)
	principal.SessionExpiresAt = expiresAt
	return principal, nil
}

func (repository *Repository) RevokeSession(ctx context.Context, tokenHash string) error {
	if repository == nil || repository.pool == nil {
		return fmt.Errorf("auth repository is not configured")
	}

	_, err := repository.pool.Exec(ctx, `
UPDATE auth_sessions
SET revoked_at = NOW()
WHERE session_token_hash = $1
  AND revoked_at IS NULL`, tokenHash)
	if err != nil {
		return fmt.Errorf("revoke auth session: %w", err)
	}
	return nil
}
