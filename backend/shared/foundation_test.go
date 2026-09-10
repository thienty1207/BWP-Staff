package shared

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/thienty1207/BWP-Staff/backend/config"
)

func TestFoundationAgainstExplicitTestDatabase(t *testing.T) {
	databaseURL := os.Getenv("DATABASE_TEST_URL")
	if databaseURL == "" {
		t.Skip("set DATABASE_TEST_URL to run the PostgreSQL foundation contract")
	}
	t.Setenv("DATABASE_URL", databaseURL)
	t.Setenv("SEED_DEVELOPMENT_DATA", "false")

	settings, err := config.Load()
	if err != nil {
		t.Fatalf("load test database config: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	pool, err := Connect(ctx, settings)
	if err != nil {
		t.Fatalf("connect test database: %v", err)
	}
	defer pool.Close()

	if err := RunMigrations(ctx, pool, filepath.Join("..", "migrations")); err != nil {
		t.Fatalf("run migrations: %v", err)
	}

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
WHERE table_schema = 'public' AND table_name = ANY($1)`, requiredTables).Scan(&tableCount); err != nil {
		t.Fatalf("count foundation tables: %v", err)
	}
	if tableCount != len(requiredTables) {
		t.Fatalf("expected %d foundation tables, got %d", len(requiredTables), tableCount)
	}

	var migrationCount int
	if err := pool.QueryRow(ctx, "SELECT COUNT(*) FROM _sqlx_migrations WHERE success = TRUE").Scan(&migrationCount); err != nil {
		t.Fatalf("count applied migrations: %v", err)
	}
	if migrationCount != 19 {
		t.Fatalf("expected 19 applied migrations, got %d", migrationCount)
	}

	var ticketStatuses []string
	rows, err := pool.Query(ctx, `
SELECT enumlabel::text
FROM pg_enum
JOIN pg_type ON pg_type.oid = pg_enum.enumtypid
WHERE pg_type.typname = 'ticket_status'
ORDER BY enumsortorder`)
	if err != nil {
		t.Fatalf("read ticket statuses: %v", err)
	}
	for rows.Next() {
		var status string
		if err := rows.Scan(&status); err != nil {
			rows.Close()
			t.Fatalf("scan ticket status: %v", err)
		}
		ticketStatuses = append(ticketStatuses, status)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		t.Fatalf("read ticket statuses: %v", err)
	}
	rows.Close()
	if len(ticketStatuses) != 3 || ticketStatuses[0] != "pending" || ticketStatuses[1] != "accepted" || ticketStatuses[2] != "closed" {
		t.Fatalf("unexpected ticket statuses: %v", ticketStatuses)
	}

	var indexCount int
	if err := pool.QueryRow(ctx, `
SELECT COUNT(*)
FROM pg_indexes
WHERE schemaname = 'public'
  AND indexname = ANY($1)`, []string{
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
	}).Scan(&indexCount); err != nil {
		t.Fatalf("count foundation indexes: %v", err)
	}
	if indexCount != 12 {
		t.Fatalf("expected 12 foundation indexes, got %d", indexCount)
	}
}
