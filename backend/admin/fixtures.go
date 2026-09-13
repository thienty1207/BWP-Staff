package admin

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type developmentDepartment struct {
	code        string
	name        string
	description string
}

type developmentLocation struct {
	code        string
	name        string
	description string
}

const developmentDepartmentDescription = "Development department seed data"

var developmentDepartments = []developmentDepartment{
	{code: "CON", name: "Concierge", description: developmentDepartmentDescription},
	{code: "DA", name: "Damaged Asset", description: developmentDepartmentDescription},
	{code: "FB", name: "F&B", description: developmentDepartmentDescription},
	{code: "FIN", name: "Finance Request", description: developmentDepartmentDescription},
	{code: "FO", name: "Front Office", description: developmentDepartmentDescription},
	{code: "HK", name: "Housekeeping", description: developmentDepartmentDescription},
	{code: "HKPPM", name: "Housekeeping PPM", description: developmentDepartmentDescription},
	{code: "IT", name: "IT", description: developmentDepartmentDescription},
	{code: "KIT", name: "Kitchen", description: developmentDepartmentDescription},
	{code: "LDRY", name: "Laundry", description: developmentDepartmentDescription},
	{code: "LF", name: "Lost & Found", description: developmentDepartmentDescription},
	{code: "MAINT", name: "Maintenance", description: developmentDepartmentDescription},
	{code: "REC", name: "REC", description: developmentDepartmentDescription},
	{code: "SEC", name: "Security", description: developmentDepartmentDescription},
}

var developmentLocations = []developmentLocation{
	{code: "LOBBY", name: "Lobby", description: "Development location seed data"},
	{code: "BALLROOM", name: "Ballroom", description: "Development location seed data"},
	{code: "BACK-OFFICE", name: "Back Office", description: "Development location seed data"},
	{code: "ROOM-8020", name: "Room 8020", description: "Development location seed data"},
	{code: "ROOM-7309", name: "Room 7309", description: "Development location seed data"},
	{code: "VILLA", name: "Villa", description: "Development location seed data"},
}

// SeedDevelopmentFixtures inserts only local development reference data.
// Callers must enforce the development-only configuration guard.
func SeedDevelopmentFixtures(ctx context.Context, pool *pgxpool.Pool) error {
	transaction, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin development fixtures seed: %w", err)
	}

	for _, department := range developmentDepartments {
		if _, err := transaction.Exec(ctx, `
INSERT INTO departments (code, name, description)
VALUES ($1, $2, $3)
ON CONFLICT (code) DO UPDATE
SET name = EXCLUDED.name,
    description = EXCLUDED.description,
    is_active = TRUE,
    updated_at = NOW()
WHERE departments.description = $3`, department.code, department.name, department.description); err != nil {
			_ = transaction.Rollback(ctx)
			return fmt.Errorf("insert development department %s: %w", department.code, err)
		}
	}

	if _, err := transaction.Exec(ctx, `
UPDATE departments
SET is_active = FALSE,
    updated_at = NOW()
WHERE code = ANY($1::text[])
  AND description = $2`, []string{"ENG", "HR"}, developmentDepartmentDescription); err != nil {
		_ = transaction.Rollback(ctx)
		return fmt.Errorf("deactivate legacy development departments: %w", err)
	}

	for _, location := range developmentLocations {
		if _, err := transaction.Exec(ctx, `
INSERT INTO locations (code, name, description)
VALUES ($1, $2, $3)
ON CONFLICT (code) DO NOTHING`, location.code, location.name, location.description); err != nil {
			_ = transaction.Rollback(ctx)
			return fmt.Errorf("insert development location %s: %w", location.code, err)
		}
	}

	if err := transaction.Commit(ctx); err != nil {
		return fmt.Errorf("commit development fixtures seed: %w", err)
	}
	return nil
}
