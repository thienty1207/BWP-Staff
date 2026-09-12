package admin

import (
	"context"
	"strings"
	"testing"

	"github.com/thienty1207/BWP-Staff/backend/config"
	"github.com/thienty1207/BWP-Staff/backend/shared/security"
)

func TestSeedDevelopmentAdminRejectsOversizedPasswordBeforeDatabaseWork(t *testing.T) {
	_, err := SeedDevelopmentAdmin(context.Background(), nil, config.SeedAdmin{
		Username:       "test-only-admin",
		EmployeeCode:   "TEST-ONLY-ADMIN",
		Password:       strings.Repeat("p", security.MaxPasswordBytes+1),
		FullName:       "Test-only Admin",
		DepartmentCode: "IT",
	})
	if err == nil {
		t.Fatal("expected oversized development admin password to be rejected")
	}
}
