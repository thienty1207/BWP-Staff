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

var developmentDepartments = []developmentDepartment{
	{code: "IT", name: "IT Department", description: "Development department seed data"},
	{code: "HK", name: "Housekeeping", description: "Development department seed data"},
	{code: "FO", name: "Front Office", description: "Development department seed data"},
	{code: "ENG", name: "Engineering", description: "Development department seed data"},
	{code: "HR", name: "Human Resources", description: "Development department seed data"},
	{code: "FB", name: "Food & Beverage", description: "Development department seed data"},
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
ON CONFLICT (code) DO NOTHING`, department.code, department.name, department.description); err != nil {
			_ = transaction.Rollback(ctx)
			return fmt.Errorf("insert development department %s: %w", department.code, err)
		}
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
