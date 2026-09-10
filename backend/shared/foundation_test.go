package shared_test

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/thienty1207/BWP-Staff/backend/admin"
	"github.com/thienty1207/BWP-Staff/backend/config"
	"github.com/thienty1207/BWP-Staff/backend/shared"
	"github.com/thienty1207/BWP-Staff/backend/shared/security"
)

func TestSPEC01PostgreSQLFoundation(t *testing.T) {
	pool, ctx := openSPEC01Pool(t)

	migrationsDirectory := filepath.Join("..", "migrations")
	if err := shared.RunMigrations(ctx, pool, migrationsDirectory, 5*time.Second); err != nil {
		t.Fatalf("run migrations on a clean schema: %v", err)
	}
	if err := shared.RunMigrations(ctx, pool, migrationsDirectory, 5*time.Second); err != nil {
		t.Fatalf("run migrations a second time: %v", err)
	}

	t.Run("pgxpool pings and migrations are clean", func(t *testing.T) {
		if err := pool.Ping(ctx); err != nil {
			t.Fatalf("ping PostgreSQL pool: %v", err)
		}

		var migrationCount int
		if err := pool.QueryRow(ctx, "SELECT COUNT(*) FROM _sqlx_migrations WHERE success = TRUE").Scan(&migrationCount); err != nil {
			t.Fatalf("count successful migrations: %v", err)
		}
		if migrationCount != 19 {
			t.Fatalf("expected 19 successful migrations, got %d", migrationCount)
		}

		var departmentCount, locationCount int
		if err := pool.QueryRow(ctx, "SELECT COUNT(*) FROM departments").Scan(&departmentCount); err != nil {
			t.Fatalf("count departments after normal migrations: %v", err)
		}
		if err := pool.QueryRow(ctx, "SELECT COUNT(*) FROM locations").Scan(&locationCount); err != nil {
			t.Fatalf("count locations after normal migrations: %v", err)
		}
		if departmentCount != 0 || locationCount != 0 {
			t.Fatalf("normal migrations inserted development fixtures: departments=%d locations=%d", departmentCount, locationCount)
		}
	})

	t.Run("required tables exist", func(t *testing.T) {
		requiredTables := []string{
			"departments",
			"users",
			"user_preferences",
			"auth_sessions",
			"locations",
			"tickets",
			"ticket_activity",
			"ticket_messages",
			"message_attachments",
			"checklist_items",
			"announcements",
			"staff_meals",
			"notifications",
			"audit_logs",
		}

		var tableCount int
		if err := pool.QueryRow(ctx, `
SELECT COUNT(*)
FROM information_schema.tables
WHERE table_schema = current_schema()
  AND table_type = 'BASE TABLE'
  AND table_name = ANY($1)`, requiredTables).Scan(&tableCount); err != nil {
			t.Fatalf("count foundation tables: %v", err)
		}
		if tableCount != len(requiredTables) {
			t.Fatalf("expected %d foundation tables, got %d", len(requiredTables), tableCount)
		}
	})

	t.Run("required enum values exist", func(t *testing.T) {
		expected := map[string][]string{
			"user_role": {
				"admin",
				"staff",
			},
			"ticket_status": {
				"pending",
				"accepted",
				"closed",
			},
			"notification_type": {
				"ticket_created",
				"ticket_accepted",
				"ticket_assigned",
				"ticket_closed",
				"new_message",
				"announcement",
				"system",
			},
			"audit_action": {
				"create",
				"update",
				"delete",
				"login",
				"logout",
				"accept",
				"assign",
				"close",
				"publish",
				"unpublish",
			},
		}

		actual := make(map[string][]string, len(expected))
		rows, err := pool.Query(ctx, `
SELECT pg_type.typname, pg_enum.enumlabel::text
FROM pg_type
JOIN pg_namespace ON pg_namespace.oid = pg_type.typnamespace
JOIN pg_enum ON pg_enum.enumtypid = pg_type.oid
WHERE pg_namespace.nspname = current_schema()
  AND pg_type.typname = ANY($1)
ORDER BY pg_type.typname, pg_enum.enumsortorder`, []string{
			"user_role",
			"ticket_status",
			"notification_type",
			"audit_action",
		})
		if err != nil {
			t.Fatalf("read enum values: %v", err)
		}
		defer rows.Close()
		for rows.Next() {
			var enumName, enumValue string
			if err := rows.Scan(&enumName, &enumValue); err != nil {
				t.Fatalf("scan enum value: %v", err)
			}
			actual[enumName] = append(actual[enumName], enumValue)
		}
		if err := rows.Err(); err != nil {
			t.Fatalf("read enum values: %v", err)
		}
		if !reflect.DeepEqual(actual, expected) {
			t.Fatalf("unexpected enum values: got=%v want=%v", actual, expected)
		}
	})

	t.Run("required indexes exist", func(t *testing.T) {
		requiredIndexes := []string{
			"idx_tickets_status_created_at",
			"idx_tickets_requester_created_at",
			"idx_tickets_assigned_status",
			"idx_tickets_department_status_created",
			"idx_tickets_location_created_at",
			"idx_ticket_messages_ticket_created",
			"idx_notifications_user_unread_created",
			"idx_announcements_published_at",
			"idx_announcements_author_created_at",
			"idx_auth_sessions_active_user_expires_at",
			"idx_staff_meals_valid_range",
			"idx_audit_logs_entity_created",
		}

		rows, err := pool.Query(ctx, `
SELECT indexname
FROM pg_indexes
WHERE schemaname = current_schema()
  AND indexname = ANY($1)`, requiredIndexes)
		if err != nil {
			t.Fatalf("read foundation indexes: %v", err)
		}
		defer rows.Close()
		found := make(map[string]bool, len(requiredIndexes))
		for rows.Next() {
			var indexName string
			if err := rows.Scan(&indexName); err != nil {
				t.Fatalf("scan foundation index: %v", err)
			}
			found[indexName] = true
		}
		if err := rows.Err(); err != nil {
			t.Fatalf("read foundation indexes: %v", err)
		}
		for _, indexName := range requiredIndexes {
			if !found[indexName] {
				t.Errorf("missing required index %s", indexName)
			}
		}
	})

	var departmentID int64
	if err := pool.QueryRow(ctx, "SELECT id FROM departments WHERE code = 'IT'").Scan(&departmentID); err != pgx.ErrNoRows {
		if err != nil {
			t.Fatalf("unexpected IT department before explicit seed: %v", err)
		}
		t.Fatal("development department exists before explicit seed")
	}

	t.Run("development fixtures use explicit seed workflow", func(t *testing.T) {
		if err := admin.SeedDevelopmentFixtures(ctx, pool); err != nil {
			t.Fatalf("seed development fixtures: %v", err)
		}
		if err := admin.SeedDevelopmentFixtures(ctx, pool); err != nil {
			t.Fatalf("seed development fixtures a second time: %v", err)
		}

		var departmentCount, locationCount int
		if err := pool.QueryRow(ctx, "SELECT COUNT(*) FROM departments").Scan(&departmentCount); err != nil {
			t.Fatalf("count seeded departments: %v", err)
		}
		if err := pool.QueryRow(ctx, "SELECT COUNT(*) FROM locations").Scan(&locationCount); err != nil {
			t.Fatalf("count seeded locations: %v", err)
		}
		if departmentCount != 6 || locationCount != 6 {
			t.Fatalf("unexpected development fixture counts: departments=%d locations=%d", departmentCount, locationCount)
		}

		var namedLocations int
		if err := pool.QueryRow(ctx, `
SELECT COUNT(*)
FROM locations
WHERE name = ANY($1)`, []string{"Lobby", "Ballroom", "Back Office", "Room 8020", "Room 7309", "Villa"}).Scan(&namedLocations); err != nil {
			t.Fatalf("count named development locations: %v", err)
		}
		if namedLocations != 6 {
			t.Fatalf("expected all six named development locations, got %d", namedLocations)
		}
	})

	var seededAdminID int64
	t.Run("development admin password is Argon2id", func(t *testing.T) {
		outcome, err := admin.SeedDevelopmentAdmin(ctx, pool, config.SeedAdmin{
			Username:       "hothienty",
			Password:       "test-only-password",
			EmployeeCode:   "IT-ADMIN-001",
			FullName:       "Ho Thien Ty",
			DepartmentCode: "IT",
		})
		if err != nil {
			t.Fatalf("seed development admin: %v", err)
		}
		if outcome != admin.SeedCreated {
			t.Fatalf("expected development admin to be created, got %s", outcome)
		}

		if err := pool.QueryRow(ctx, `
SELECT id
FROM users
WHERE username = $1`, "hothienty").Scan(&seededAdminID); err != nil {
			t.Fatalf("read seeded admin id: %v", err)
		}
		var passwordHash string
		if err := pool.QueryRow(ctx, `
SELECT password_hash
FROM users
WHERE id = $1`, seededAdminID).Scan(&passwordHash); err != nil {
			t.Fatalf("read seeded admin password hash: %v", err)
		}
		if !strings.HasPrefix(passwordHash, "$argon2id$v=19$") {
			t.Fatal("development admin password is not Argon2id")
		}
		if !security.VerifyPassword("test-only-password", passwordHash) {
			t.Fatal("development admin password does not verify")
		}
		if security.VerifyPassword("wrong-password", passwordHash) {
			t.Fatal("wrong development admin password unexpectedly verified")
		}

		outcome, err = admin.SeedDevelopmentAdmin(ctx, pool, config.SeedAdmin{
			Username:       "hothienty",
			Password:       "test-only-password",
			EmployeeCode:   "IT-ADMIN-001",
			FullName:       "Ho Thien Ty",
			DepartmentCode: "IT",
		})
		if err != nil {
			t.Fatalf("repeat development admin seed: %v", err)
		}
		if outcome != admin.SeedAlreadyExists {
			t.Fatalf("expected repeat development admin seed to be idempotent, got %s", outcome)
		}
	})

	if err := pool.QueryRow(ctx, "SELECT id FROM departments WHERE code = 'IT'").Scan(&departmentID); err != nil {
		t.Fatalf("read seeded IT department: %v", err)
	}

	t.Run("invalid foreign key is rejected", func(t *testing.T) {
		_, err := pool.Exec(ctx, `
INSERT INTO tickets (requester_id, department_id, title)
VALUES ($1, $2, $3)`, int64(9_999_999), departmentID, "invalid requester")
		assertPostgreSQLErrorCode(t, err, "23503")
	})

	t.Run("duplicate usernames are rejected", func(t *testing.T) {
		insertTestUser(t, ctx, pool, "spec01-username", "SPEC01-USERNAME-1", nil)
		_, err := pool.Exec(ctx, `
INSERT INTO users (username, employee_code, email, password_hash, full_name, department_id, role)
VALUES ($1, $2, NULL, $3, $4, $5, 'staff'::user_role)`,
			"spec01-username",
			"SPEC01-USERNAME-2",
			mustTestPasswordHash(t),
			"Duplicate Username",
			departmentID,
		)
		assertPostgreSQLErrorCode(t, err, "23505")
	})

	t.Run("duplicate employee codes are rejected", func(t *testing.T) {
		insertTestUser(t, ctx, pool, "spec01-employee-1", "SPEC01-EMPLOYEE", nil)
		_, err := pool.Exec(ctx, `
INSERT INTO users (username, employee_code, email, password_hash, full_name, department_id, role)
VALUES ($1, $2, NULL, $3, $4, $5, 'staff'::user_role)`,
			"spec01-employee-2",
			"SPEC01-EMPLOYEE",
			mustTestPasswordHash(t),
			"Duplicate Employee",
			departmentID,
		)
		assertPostgreSQLErrorCode(t, err, "23505")
	})

	t.Run("duplicate non-null emails are rejected", func(t *testing.T) {
		email := "Spec01-Email@example.invalid"
		insertTestUser(t, ctx, pool, "spec01-email-1", "SPEC01-EMAIL-1", &email)
		duplicateEmail := "spec01-email@example.invalid"
		_, err := pool.Exec(ctx, `
INSERT INTO users (username, employee_code, email, password_hash, full_name, department_id, role)
VALUES ($1, $2, $3, $4, $5, $6, 'staff'::user_role)`,
			"spec01-email-2",
			"SPEC01-EMAIL-2",
			duplicateEmail,
			mustTestPasswordHash(t),
			"Duplicate Email",
			departmentID,
		)
		assertPostgreSQLErrorCode(t, err, "23505")
	})

	t.Run("invalid ticket status is rejected", func(t *testing.T) {
		_, err := pool.Exec(ctx, `
INSERT INTO tickets (requester_id, department_id, title, status)
VALUES ($1, $2, $3, $4)`, seededAdminID, departmentID, "invalid status", "not-a-ticket-status")
		assertPostgreSQLErrorCode(t, err, "22P02")
	})
}

func openSPEC01Pool(t *testing.T) (*pgxpool.Pool, context.Context) {
	t.Helper()

	databaseURL := strings.TrimSpace(os.Getenv("DATABASE_TEST_URL"))
	if databaseURL == "" {
		t.Skip("set DATABASE_TEST_URL to run the PostgreSQL foundation contract")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	schema := fmt.Sprintf("spec01_%d", time.Now().UnixNano())
	quotedSchema := pgx.Identifier{schema}.Sanitize()

	adminConnection, err := pgx.Connect(ctx, databaseURL)
	if err != nil {
		cancel()
		t.Fatalf("connect PostgreSQL test database: %v", err)
	}
	if _, err := adminConnection.Exec(ctx, "CREATE SCHEMA "+quotedSchema); err != nil {
		_ = adminConnection.Close(ctx)
		cancel()
		t.Fatalf("create isolated PostgreSQL test schema: %v", err)
	}
	if err := adminConnection.Close(ctx); err != nil {
		cancel()
		t.Fatalf("close PostgreSQL setup connection: %v", err)
	}

	var pool *pgxpool.Pool
	cleanup := func() {
		if pool != nil {
			pool.Close()
			pool = nil
		}
		cleanupContext, cleanupCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cleanupCancel()
		cleanupConnection, cleanupErr := pgx.Connect(cleanupContext, databaseURL)
		if cleanupErr != nil {
			t.Errorf("connect PostgreSQL cleanup connection: %v", cleanupErr)
			cancel()
			return
		}
		defer cleanupConnection.Close(cleanupContext)
		if _, cleanupErr := cleanupConnection.Exec(cleanupContext, "DROP SCHEMA "+quotedSchema+" CASCADE"); cleanupErr != nil {
			t.Errorf("drop isolated PostgreSQL test schema: %v", cleanupErr)
		}
		cancel()
	}

	parsedURL, err := url.Parse(databaseURL)
	if err != nil {
		cleanup()
		t.Fatalf("parse PostgreSQL test URL: %v", err)
	}
	query := parsedURL.Query()
	query.Del("options")
	encodedQuery := query.Encode()
	encodedOptions := strings.ReplaceAll(url.QueryEscape("-c search_path="+schema+",public"), "+", "%20")
	if encodedQuery != "" {
		encodedQuery += "&"
	}
	parsedURL.RawQuery = encodedQuery + "options=" + encodedOptions

	t.Setenv("DATABASE_URL", parsedURL.String())
	t.Setenv("DATABASE_MAX_CONNECTIONS", "2")
	t.Setenv("DATABASE_MIN_CONNECTIONS", "0")
	t.Setenv("DATABASE_ACQUIRE_TIMEOUT_SECONDS", "5")
	t.Setenv("SEED_DEVELOPMENT_DATA", "false")

	settings, err := config.Load()
	if err != nil {
		cleanup()
		t.Fatalf("load isolated PostgreSQL test config: %v", err)
	}
	p, err := shared.Connect(ctx, settings)
	if err != nil {
		cleanup()
		t.Fatalf("connect isolated PostgreSQL pool: %v", err)
	}
	pool = p
	t.Cleanup(cleanup)

	var currentSchema string
	if err := pool.QueryRow(ctx, "SELECT current_schema()").Scan(&currentSchema); err != nil {
		t.Fatalf("read isolated PostgreSQL schema: %v", err)
	}
	if currentSchema != schema {
		t.Fatalf("PostgreSQL test pool is not isolated: got schema %q, want %q", currentSchema, schema)
	}
	return pool, ctx
}

func insertTestUser(t *testing.T, ctx context.Context, pool *pgxpool.Pool, username, employeeCode string, email *string) int64 {
	t.Helper()
	var id int64
	if err := pool.QueryRow(ctx, `
INSERT INTO users (username, employee_code, email, password_hash, full_name, department_id, role)
VALUES ($1, $2, $3, $4, $5, (SELECT id FROM departments WHERE code = 'IT'), 'staff'::user_role)
RETURNING id`, username, employeeCode, email, mustTestPasswordHash(t), "SPEC-01 Test User").Scan(&id); err != nil {
		t.Fatalf("insert test user %s: %v", username, err)
	}
	return id
}

func mustTestPasswordHash(t *testing.T) string {
	t.Helper()
	hash, err := security.HashPassword("test-only-password")
	if err != nil {
		t.Fatalf("hash test password: %v", err)
	}
	return hash
}

func assertPostgreSQLErrorCode(t *testing.T, err error, expectedCode string) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected PostgreSQL error %s", expectedCode)
	}
	var postgresError *pgconn.PgError
	if !errors.As(err, &postgresError) {
		t.Fatalf("expected PostgreSQL error %s, got %T: %v", expectedCode, err, err)
	}
	if postgresError.Code != expectedCode {
		t.Fatalf("expected PostgreSQL error %s, got %s: %s", expectedCode, postgresError.Code, postgresError.Message)
	}
}
