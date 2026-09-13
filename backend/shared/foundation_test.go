package shared_test

import (
	"context"
	"errors"
	"fmt"
	"net"
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
	"github.com/joho/godotenv"
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

		expectedMigrationVersions := []int64{
			1, 2, 3, 4, 5, 6, 7, 8, 9,
			10, 11, 12, 13, 14, 15, 16, 17, 19, 20,
		}
		rows, err := pool.Query(ctx, `
SELECT version, success
FROM _sqlx_migrations
ORDER BY version`)
		if err != nil {
			t.Fatalf("read migration ledger: %v", err)
		}
		defer rows.Close()
		actualMigrationVersions := make([]int64, 0, len(expectedMigrationVersions))
		for rows.Next() {
			var version int64
			var success bool
			if err := rows.Scan(&version, &success); err != nil {
				t.Fatalf("scan migration ledger: %v", err)
			}
			if !success {
				t.Fatalf("migration %d was recorded as unsuccessful", version)
			}
			actualMigrationVersions = append(actualMigrationVersions, version)
		}
		if err := rows.Err(); err != nil {
			t.Fatalf("read migration ledger: %v", err)
		}
		if !reflect.DeepEqual(actualMigrationVersions, expectedMigrationVersions) {
			t.Fatalf("unexpected schema migration versions: got=%v want=%v", actualMigrationVersions, expectedMigrationVersions)
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
			"ticket_assigned_departments",
			"ticket_assigned_users",
			"ticket_attachments",
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
			"idx_tickets_department_status_created",
			"idx_tickets_location_created_at",
			"idx_ticket_assigned_departments_ticket_department",
			"idx_ticket_assigned_departments_department_ticket",
			"idx_ticket_assigned_users_ticket_user",
			"idx_ticket_assigned_users_user_ticket",
			"idx_ticket_attachments_ticket_created",
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

		var legacyIndexCount int
		if err := pool.QueryRow(ctx, `
SELECT COUNT(*)
FROM pg_indexes
WHERE schemaname = current_schema()
  AND indexname = 'idx_tickets_assigned_status'`).Scan(&legacyIndexCount); err != nil {
			t.Fatalf("check retired assignment index: %v", err)
		}
		if legacyIndexCount != 0 {
			t.Fatalf("retired assignment index still exists")
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
		const developmentDescription = "Development department seed data"
		legacyDepartments := []struct {
			code string
			name string
		}{
			{code: "IT", name: "IT Department"},
			{code: "FB", name: "Food & Beverage"},
			{code: "ENG", name: "Engineering"},
			{code: "HR", name: "Human Resources"},
		}
		legacyIDs := make(map[string]int64, len(legacyDepartments))
		for _, department := range legacyDepartments {
			var legacyID int64
			if err := pool.QueryRow(ctx, `
INSERT INTO departments (code, name, description)
VALUES ($1, $2, $3)
RETURNING id`, department.code, department.name, developmentDescription).Scan(&legacyID); err != nil {
				t.Fatalf("insert legacy development department %s: %v", department.code, err)
			}
			legacyIDs[department.code] = legacyID
		}
		var unrelatedID int64
		if err := pool.QueryRow(ctx, `
INSERT INTO departments (code, name, description, is_active)
VALUES ('SPEC01-REAL', 'Real Operations', 'Non-development department', TRUE)
RETURNING id`).Scan(&unrelatedID); err != nil {
			t.Fatalf("insert unrelated department: %v", err)
		}

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
		if departmentCount != 17 || locationCount != 6 {
			t.Fatalf("unexpected development fixture counts: departments=%d locations=%d", departmentCount, locationCount)
		}

		expectedDepartments := map[string]string{
			"CON":   "Concierge",
			"DA":    "Damaged Asset",
			"FB":    "F&B",
			"FIN":   "Finance Request",
			"FO":    "Front Office",
			"HK":    "Housekeeping",
			"HKPPM": "Housekeeping PPM",
			"IT":    "IT",
			"KIT":   "Kitchen",
			"LDRY":  "Laundry",
			"LF":    "Lost & Found",
			"MAINT": "Maintenance",
			"REC":   "REC",
			"SEC":   "Security",
		}
		rows, err := pool.Query(ctx, `
SELECT code, name, is_active
FROM departments
WHERE code = ANY($1::text[])
ORDER BY code`, []string{
			"CON", "DA", "FB", "FIN", "FO", "HK", "HKPPM", "IT", "KIT", "LDRY", "LF", "MAINT", "REC", "SEC",
		})
		if err != nil {
			t.Fatalf("read approved departments: %v", err)
		}
		defer rows.Close()
		activeDepartments := make(map[string]string, len(expectedDepartments))
		for rows.Next() {
			var code, name string
			var active bool
			if err := rows.Scan(&code, &name, &active); err != nil {
				t.Fatalf("scan approved department: %v", err)
			}
			if !active {
				continue
			}
			activeDepartments[code] = name
		}
		if err := rows.Err(); err != nil {
			t.Fatalf("iterate approved departments: %v", err)
		}
		if !reflect.DeepEqual(activeDepartments, expectedDepartments) {
			t.Fatalf("unexpected active development departments: got=%v want=%v", activeDepartments, expectedDepartments)
		}

		for _, department := range []struct {
			code string
			name string
		}{
			{code: "ENG", name: "Engineering"},
			{code: "HR", name: "Human Resources"},
		} {
			var active bool
			if err := pool.QueryRow(ctx, "SELECT is_active FROM departments WHERE code = $1", department.code).Scan(&active); err != nil {
				t.Fatalf("read legacy department %s: %v", department.code, err)
			}
			if active {
				t.Fatalf("legacy development department %s is still active", department.code)
			}
		}

		for _, department := range []struct {
			code string
			name string
			id   int64
		}{
			{code: "IT", name: "IT", id: legacyIDs["IT"]},
			{code: "FB", name: "F&B", id: legacyIDs["FB"]},
		} {
			var id int64
			var name string
			if err := pool.QueryRow(ctx, "SELECT id, name FROM departments WHERE code = $1", department.code).Scan(&id, &name); err != nil {
				t.Fatalf("read aligned department %s: %v", department.code, err)
			}
			if id != department.id || name != department.name {
				t.Fatalf("department %s was not updated in place: got id=%d name=%q want id=%d name=%q", department.code, id, name, department.id, department.name)
			}
		}

		var unrelatedCode string
		var unrelatedName, unrelatedDescription string
		var unrelatedActive bool
		if err := pool.QueryRow(ctx, `
SELECT code, name, description, is_active
FROM departments
WHERE id = $1`, unrelatedID).Scan(&unrelatedCode, &unrelatedName, &unrelatedDescription, &unrelatedActive); err != nil {
			t.Fatalf("read unrelated department: %v", err)
		}
		if unrelatedCode != "SPEC01-REAL" || unrelatedName != "Real Operations" || unrelatedDescription != "Non-development department" || !unrelatedActive {
			t.Fatalf("unrelated department changed during development seed: code=%q name=%q description=%q active=%t", unrelatedCode, unrelatedName, unrelatedDescription, unrelatedActive)
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

	t.Run("ticket request fields enforce priority and optional due time", func(t *testing.T) {
		dueAt := time.Date(2026, 9, 11, 12, 30, 0, 0, time.UTC)
		var ticketID int64
		if err := pool.QueryRow(ctx, `
INSERT INTO tickets (requester_id, department_id, title, priority, due_at)
VALUES ($1, $2, $3, $4, $5)
RETURNING id`, seededAdminID, departmentID, "Priority request", true, dueAt).Scan(&ticketID); err != nil {
			t.Fatalf("insert priority ticket: %v", err)
		}

		var priority bool
		var storedDueAt time.Time
		if err := pool.QueryRow(ctx, `
SELECT priority, due_at
FROM tickets
WHERE id = $1`, ticketID).Scan(&priority, &storedDueAt); err != nil {
			t.Fatalf("read priority ticket: %v", err)
		}
		if !priority {
			t.Fatal("priority ticket was stored as normal")
		}
		if !storedDueAt.Equal(dueAt) {
			t.Fatalf("unexpected due time: got=%s want=%s", storedDueAt, dueAt)
		}

		var defaultPriority, dueAtIsNull bool
		if err := pool.QueryRow(ctx, `
INSERT INTO tickets (requester_id, department_id, title)
VALUES ($1, $2, $3)
RETURNING priority, due_at IS NULL`, seededAdminID, departmentID, "Default request").Scan(&defaultPriority, &dueAtIsNull); err != nil {
			t.Fatalf("insert default ticket: %v", err)
		}
		if defaultPriority {
			t.Fatal("ticket priority defaulted to true")
		}
		if !dueAtIsNull {
			t.Fatal("ticket due time did not default to NULL")
		}
	})

	t.Run("ticket supports multiple department and user assignments", func(t *testing.T) {
		var secondDepartmentID int64
		if err := pool.QueryRow(ctx, "SELECT id FROM departments WHERE code = 'HK'").Scan(&secondDepartmentID); err != nil {
			t.Fatalf("read seeded HK department: %v", err)
		}
		firstUserID := insertTestUser(t, ctx, pool, "spec01-assignment-user-1", "SPEC01-ASSIGNMENT-1", nil)
		secondUserID := insertTestUser(t, ctx, pool, "spec01-assignment-user-2", "SPEC01-ASSIGNMENT-2", nil)

		var ticketID int64
		if err := pool.QueryRow(ctx, `
INSERT INTO tickets (requester_id, department_id, title)
VALUES ($1, $2, $3)
RETURNING id`, seededAdminID, departmentID, "Multi-assignment request").Scan(&ticketID); err != nil {
			t.Fatalf("insert multi-assignment ticket: %v", err)
		}

		for _, assignedDepartmentID := range []int64{departmentID, secondDepartmentID} {
			if _, err := pool.Exec(ctx, `
INSERT INTO ticket_assigned_departments (ticket_id, department_id)
VALUES ($1, $2)`, ticketID, assignedDepartmentID); err != nil {
				t.Fatalf("insert department assignment %d: %v", assignedDepartmentID, err)
			}
		}
		for _, assignedUserID := range []int64{firstUserID, secondUserID} {
			if _, err := pool.Exec(ctx, `
INSERT INTO ticket_assigned_users (ticket_id, user_id)
VALUES ($1, $2)`, ticketID, assignedUserID); err != nil {
				t.Fatalf("insert user assignment %d: %v", assignedUserID, err)
			}
		}

		var departmentAssignments, userAssignments int
		if err := pool.QueryRow(ctx, `
SELECT
    (SELECT COUNT(*) FROM ticket_assigned_departments WHERE ticket_id = $1),
    (SELECT COUNT(*) FROM ticket_assigned_users WHERE ticket_id = $1)`, ticketID).Scan(&departmentAssignments, &userAssignments); err != nil {
			t.Fatalf("count ticket assignments: %v", err)
		}
		if departmentAssignments != 2 || userAssignments != 2 {
			t.Fatalf("unexpected assignment counts: departments=%d users=%d", departmentAssignments, userAssignments)
		}

		_, err := pool.Exec(ctx, `
INSERT INTO ticket_assigned_departments (ticket_id, department_id)
VALUES ($1, $2)`, ticketID, departmentID)
		assertPostgreSQLErrorCode(t, err, "23505")

		_, err = pool.Exec(ctx, `
INSERT INTO ticket_assigned_users (ticket_id, user_id)
VALUES ($1, $2)`, ticketID, firstUserID)
		assertPostgreSQLErrorCode(t, err, "23505")

		var status string
		if err := pool.QueryRow(ctx, "SELECT status::text FROM tickets WHERE id = $1", ticketID).Scan(&status); err != nil {
			t.Fatalf("read multi-assignment ticket status: %v", err)
		}
		if status != "pending" {
			t.Fatalf("assignment changed ticket status: got=%s want=pending", status)
		}
	})

	t.Run("ticket and message attachments remain separate metadata", func(t *testing.T) {
		var ticketID int64
		if err := pool.QueryRow(ctx, `
INSERT INTO tickets (requester_id, department_id, title)
VALUES ($1, $2, $3)
RETURNING id`, seededAdminID, departmentID, "Attachment request").Scan(&ticketID); err != nil {
			t.Fatalf("insert attachment ticket: %v", err)
		}

		var ticketAttachmentID int64
		if err := pool.QueryRow(ctx, `
INSERT INTO ticket_attachments (
    ticket_id, uploaded_by, file_name, storage_key, public_url,
    mime_type, file_size_bytes, width, height
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING id`, ticketID, seededAdminID, "request.png", "tickets/request.png", "https://storage.invalid/request.png", "image/png", int64(2048), 1280, 720).Scan(&ticketAttachmentID); err != nil {
			t.Fatalf("insert ticket attachment metadata: %v", err)
		}

		var storedFileName string
		if err := pool.QueryRow(ctx, "SELECT file_name FROM ticket_attachments WHERE id = $1", ticketAttachmentID).Scan(&storedFileName); err != nil {
			t.Fatalf("read ticket attachment metadata: %v", err)
		}
		if storedFileName != "request.png" {
			t.Fatalf("unexpected ticket attachment file name: %s", storedFileName)
		}

		var messageID int64
		if err := pool.QueryRow(ctx, `
INSERT INTO ticket_messages (ticket_id, sender_id, content)
VALUES ($1, $2, $3)
RETURNING id`, ticketID, seededAdminID, "message attachment").Scan(&messageID); err != nil {
			t.Fatalf("insert ticket message: %v", err)
		}
		if _, err := pool.Exec(ctx, `
INSERT INTO message_attachments (
    message_id, file_name, storage_key, public_url, mime_type,
    file_size_bytes, width, height
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`, messageID, "chat.png", "messages/chat.png", "https://storage.invalid/chat.png", "image/png", int64(1024), 640, 480); err != nil {
			t.Fatalf("insert message attachment metadata: %v", err)
		}

		var ticketAttachmentCount, messageAttachmentCount int
		if err := pool.QueryRow(ctx, `
SELECT
    (SELECT COUNT(*) FROM ticket_attachments WHERE ticket_id = $1),
    (SELECT COUNT(*) FROM message_attachments WHERE message_id = $2)`, ticketID, messageID).Scan(&ticketAttachmentCount, &messageAttachmentCount); err != nil {
			t.Fatalf("count separated attachment metadata: %v", err)
		}
		if ticketAttachmentCount != 1 || messageAttachmentCount != 1 {
			t.Fatalf("unexpected separated attachment counts: ticket=%d message=%d", ticketAttachmentCount, messageAttachmentCount)
		}
	})

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

func TestSPEC01MigrationRunnerExecutesEveryDiscoveredMigration(t *testing.T) {
	pool, ctx := openSPEC01Pool(t)
	migrationsDirectory := t.TempDir()

	writeTestMigration(t, migrationsDirectory, "0001_create_runner_contract.sql", `
CREATE TABLE migration_runner_contract (
    id INTEGER PRIMARY KEY,
    marker TEXT NOT NULL
);`)
	writeTestMigration(t, migrationsDirectory, "0002_insert_runner_contract.sql", `
INSERT INTO migration_runner_contract (id, marker)
VALUES (1, 'executed');`)

	if err := shared.RunMigrations(ctx, pool, migrationsDirectory, 5*time.Second); err != nil {
		t.Fatalf("run test migrations: %v", err)
	}
	if err := shared.RunMigrations(ctx, pool, migrationsDirectory, 5*time.Second); err != nil {
		t.Fatalf("run test migrations a second time: %v", err)
	}

	var marker string
	if err := pool.QueryRow(ctx, "SELECT marker FROM migration_runner_contract WHERE id = 1").Scan(&marker); err != nil {
		t.Fatalf("read migration result: %v", err)
	}
	if marker != "executed" {
		t.Fatalf("unexpected migration result: got %q want %q", marker, "executed")
	}

	var rowCount int
	if err := pool.QueryRow(ctx, "SELECT COUNT(*) FROM migration_runner_contract").Scan(&rowCount); err != nil {
		t.Fatalf("count migration results: %v", err)
	}
	if rowCount != 1 {
		t.Fatalf("second migration run duplicated data: got %d rows", rowCount)
	}

	writeTestMigration(t, migrationsDirectory, "0002_insert_runner_contract.sql", `
INSERT INTO migration_runner_contract (id, marker)
VALUES (1, 'changed');`)
	if err := shared.RunMigrations(ctx, pool, migrationsDirectory, 5*time.Second); err == nil || !strings.Contains(err.Error(), "checksum mismatch") {
		t.Fatalf("expected migration checksum mismatch, got %v", err)
	}
}

func TestSPEC01MigrationBackfillsLegacyTicketAssignment(t *testing.T) {
	pool, ctx := openSPEC01Pool(t)
	sourceDirectory := filepath.Join("..", "migrations")
	migrationsDirectory := t.TempDir()
	legacyMigrationNames := []string{
		"0001_extensions.sql",
		"0002_enums.sql",
		"0003_departments.sql",
		"0004_users.sql",
		"0005_user_preferences.sql",
		"0006_auth_sessions.sql",
		"0007_locations.sql",
		"0008_tickets.sql",
		"0009_ticket_activity.sql",
		"0010_ticket_messages.sql",
		"0011_message_attachments.sql",
		"0012_checklist_items.sql",
		"0013_announcements.sql",
		"0014_staff_meals.sql",
		"0015_notifications.sql",
		"0016_audit_logs.sql",
		"0017_indexes.sql",
		"0019_announcement_author_index.sql",
	}
	for _, name := range legacyMigrationNames {
		copyMigration(t, sourceDirectory, migrationsDirectory, name)
	}

	if err := shared.RunMigrations(ctx, pool, migrationsDirectory, 5*time.Second); err != nil {
		t.Fatalf("run legacy migrations: %v", err)
	}

	var departmentID int64
	if err := pool.QueryRow(ctx, `
INSERT INTO departments (code, name, description)
VALUES ($1, $2, $3)
RETURNING id`, "SPEC01-UPGRADE-DEPT", "SPEC-01 Upgrade Department", "SPEC-01 upgrade test").Scan(&departmentID); err != nil {
		t.Fatalf("insert upgrade department: %v", err)
	}
	passwordHash := mustTestPasswordHash(t)
	var requesterID int64
	if err := pool.QueryRow(ctx, `
INSERT INTO users (username, employee_code, password_hash, full_name, department_id, role)
VALUES ($1, $2, $3, $4, $5, 'staff'::user_role)
RETURNING id`, "spec01-upgrade-requester", "SPEC01-UPGRADE-REQUESTER", passwordHash, "SPEC-01 Upgrade Requester", departmentID).Scan(&requesterID); err != nil {
		t.Fatalf("insert upgrade requester: %v", err)
	}
	var assignedUserID int64
	if err := pool.QueryRow(ctx, `
INSERT INTO users (username, employee_code, password_hash, full_name, department_id, role)
VALUES ($1, $2, $3, $4, $5, 'staff'::user_role)
RETURNING id`, "spec01-upgrade-assignee", "SPEC01-UPGRADE-ASSIGNEE", passwordHash, "SPEC-01 Upgrade Assignee", departmentID).Scan(&assignedUserID); err != nil {
		t.Fatalf("insert upgrade assignee: %v", err)
	}

	assignedAt := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	var ticketID int64
	if err := pool.QueryRow(ctx, `
INSERT INTO tickets (requester_id, department_id, title, assigned_to, assigned_at)
VALUES ($1, $2, $3, $4, $5)
RETURNING id`, requesterID, departmentID, "Legacy assignment ticket", assignedUserID, assignedAt).Scan(&ticketID); err != nil {
		t.Fatalf("insert legacy assignment ticket: %v", err)
	}

	copyMigration(t, sourceDirectory, migrationsDirectory, "0020_ticket_assignment_and_request_fields.sql")
	if err := shared.RunMigrations(ctx, pool, migrationsDirectory, 5*time.Second); err != nil {
		t.Fatalf("run ticket foundation amendment: %v", err)
	}

	var migratedUserID int64
	var migratedAssignedAt time.Time
	if err := pool.QueryRow(ctx, `
SELECT user_id, assigned_at
FROM ticket_assigned_users
WHERE ticket_id = $1`, ticketID).Scan(&migratedUserID, &migratedAssignedAt); err != nil {
		t.Fatalf("read migrated ticket assignment: %v", err)
	}
	if migratedUserID != assignedUserID {
		t.Fatalf("legacy assignment user changed: got=%d want=%d", migratedUserID, assignedUserID)
	}
	if !migratedAssignedAt.Equal(assignedAt) {
		t.Fatalf("legacy assignment timestamp changed: got=%s want=%s", migratedAssignedAt, assignedAt)
	}

	var legacyColumnCount int
	if err := pool.QueryRow(ctx, `
SELECT COUNT(*)
FROM information_schema.columns
WHERE table_schema = current_schema()
  AND table_name = 'tickets'
  AND column_name = ANY($1)`, []string{"assigned_to", "assigned_at"}).Scan(&legacyColumnCount); err != nil {
		t.Fatalf("check retired ticket columns: %v", err)
	}
	if legacyColumnCount != 0 {
		t.Fatalf("legacy ticket assignment columns still exist: %d", legacyColumnCount)
	}

	var legacyIndexCount int
	if err := pool.QueryRow(ctx, `
SELECT COUNT(*)
FROM pg_indexes
WHERE schemaname = current_schema()
  AND indexname = 'idx_tickets_assigned_status'`).Scan(&legacyIndexCount); err != nil {
		t.Fatalf("check retired assignment index after upgrade: %v", err)
	}
	if legacyIndexCount != 0 {
		t.Fatal("legacy assignment index still exists after upgrade")
	}

	var priority bool
	var dueAtIsNull bool
	if err := pool.QueryRow(ctx, `
SELECT priority, due_at IS NULL
FROM tickets
WHERE id = $1`, ticketID).Scan(&priority, &dueAtIsNull); err != nil {
		t.Fatalf("read amended ticket request fields: %v", err)
	}
	if priority || !dueAtIsNull {
		t.Fatalf("unexpected amended ticket defaults: priority=%t due_at_is_null=%t", priority, dueAtIsNull)
	}
}

func openSPEC01Pool(t *testing.T) (*pgxpool.Pool, context.Context) {
	t.Helper()

	loadLocalTestEnv(t)
	databaseURL := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if databaseURL == "" {
		t.Skip("set DATABASE_URL to run the PostgreSQL foundation contract")
	}
	if appEnv := strings.TrimSpace(os.Getenv("APP_ENV")); appEnv != "" && !strings.EqualFold(appEnv, "development") {
		t.Skipf("refusing PostgreSQL foundation test with APP_ENV=%q", appEnv)
	}

	parsedURL, err := url.Parse(databaseURL)
	if err != nil || !isLoopbackDatabaseURL(parsedURL) {
		t.Skip("refusing PostgreSQL foundation test unless DATABASE_URL points to a loopback PostgreSQL instance")
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
	t.Setenv("APP_ENV", "development")
	t.Setenv("FRONTEND_ORIGIN", "")
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

func loadLocalTestEnv(t *testing.T) {
	t.Helper()
	if strings.TrimSpace(os.Getenv("DATABASE_URL")) != "" {
		return
	}
	if err := godotenv.Load(filepath.Join("..", ".env")); err != nil && !os.IsNotExist(err) {
		t.Fatalf("load local backend/.env for PostgreSQL test: %v", err)
	}
}

func isLoopbackDatabaseURL(databaseURL *url.URL) bool {
	if databaseURL == nil || (databaseURL.Scheme != "postgres" && databaseURL.Scheme != "postgresql") {
		return false
	}
	host := strings.TrimSuffix(strings.ToLower(databaseURL.Hostname()), ".")
	if host == "localhost" || host == "127.0.0.1" || host == "::1" {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

func writeTestMigration(t *testing.T, directory, name, contents string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(directory, name), []byte(contents), 0o600); err != nil {
		t.Fatalf("write test migration %s: %v", name, err)
	}
}

func copyMigration(t *testing.T, sourceDirectory, destinationDirectory, name string) {
	t.Helper()
	contents, err := os.ReadFile(filepath.Join(sourceDirectory, name))
	if err != nil {
		t.Fatalf("read migration %s for upgrade test: %v", name, err)
	}
	if err := os.WriteFile(filepath.Join(destinationDirectory, name), contents, 0o600); err != nil {
		t.Fatalf("copy migration %s for upgrade test: %v", name, err)
	}
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
