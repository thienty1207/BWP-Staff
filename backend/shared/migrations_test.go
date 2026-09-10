package shared

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSpec01MigrationDirectoryContainsSchemaMigrationsOnly(t *testing.T) {
	required := []string{
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

	for _, name := range required {
		if _, err := os.Stat(filepath.Join("..", "migrations", name)); err != nil {
			t.Fatalf("missing migration %s: %v", name, err)
		}
	}

	legacySeed := filepath.Join("..", "migrations", "0018_seed_development.sql")
	if _, err := os.Stat(legacySeed); err == nil {
		t.Fatalf("development fixture migration must not remain in the normal migration directory")
	} else if !os.IsNotExist(err) {
		t.Fatalf("check development fixture migration: %v", err)
	}

	migrations, err := readMigrations(filepath.Join("..", "migrations"))
	if err != nil {
		t.Fatalf("read migrations: %v", err)
	}
	if len(migrations) != len(required) {
		t.Fatalf("expected %d schema migrations, got %d", len(required), len(migrations))
	}
	for index, migration := range migrations {
		if migration.name != required[index] {
			t.Fatalf("migration %d is %s, want %s", index, migration.name, required[index])
		}
	}
}
