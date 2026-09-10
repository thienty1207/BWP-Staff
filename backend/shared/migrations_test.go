package shared

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSpec01MigrationsRemainSequential(t *testing.T) {
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
		"0018_seed_development.sql",
		"0019_announcement_author_index.sql",
	}

	for _, name := range required {
		if _, err := os.Stat(filepath.Join("..", "migrations", name)); err != nil {
			t.Fatalf("missing migration %s: %v", name, err)
		}
	}
}
