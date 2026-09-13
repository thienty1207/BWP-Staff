package admin

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
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
const development96VillasLocationDescription = "Development location seed data: 96 Villas"

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

var development96VillasLocations = []developmentLocation{
	{code: "96V", name: "96 Villas", description: development96VillasLocationDescription},
	{code: "96BWV-AREA-001", name: "96-BWV - Asian kitchen", description: development96VillasLocationDescription},
	{code: "96BWV-AREA-002", name: "96-BWV - Auxiliary Swimming Pool", description: development96VillasLocationDescription},
	{code: "96BWV-AREA-003", name: "96-BWV - Bathroom", description: development96VillasLocationDescription},
	{code: "96BWV-AREA-004", name: "96-BWV - Buffet counter", description: development96VillasLocationDescription},
	{code: "96BWV-AREA-005", name: "96-BWV - Cold kitchen", description: development96VillasLocationDescription},
	{code: "96BWV-AREA-006", name: "96-BWV - Eng Fire Pump Room", description: development96VillasLocationDescription},
	{code: "96BWV-AREA-007", name: "96-BWV - Eng Mainpool MEP Room", description: development96VillasLocationDescription},
	{code: "96BWV-AREA-008", name: "96-BWV - Eng MEP Room", description: development96VillasLocationDescription},
	{code: "96BWV-AREA-009", name: "96-BWV - Eng Subpool MEP Room", description: development96VillasLocationDescription},
	{code: "96BWV-AREA-010", name: "96-BWV - Eng Water Treatment Room", description: development96VillasLocationDescription},
	{code: "96BWV-AREA-011", name: "96-BWV - Eng Well Water treatment Room", description: development96VillasLocationDescription},
	{code: "96BWV-AREA-012", name: "96-BWV - Eng workshop", description: development96VillasLocationDescription},
	{code: "96BWV-AREA-013", name: "96-BWV - European kitchen", description: development96VillasLocationDescription},
	{code: "96BWV-AREA-014", name: "96-BWV - Extra Pool", description: development96VillasLocationDescription},
	{code: "96BWV-AREA-015", name: "96-BWV - Female Locker", description: development96VillasLocationDescription},
	{code: "96BWV-AREA-016", name: "96-BWV - FO Reception", description: development96VillasLocationDescription},
	{code: "96BWV-AREA-017", name: "96-BWV - Generator Room", description: development96VillasLocationDescription},
	{code: "96BWV-AREA-018", name: "96-BWV - Gym", description: development96VillasLocationDescription},
	{code: "96BWV-AREA-019", name: "96-BWV - HK Store", description: development96VillasLocationDescription},
	{code: "96BWV-AREA-020", name: "96-BWV - Kid club", description: development96VillasLocationDescription},
	{code: "96BWV-AREA-021", name: "96-BWV - Kid's Club", description: development96VillasLocationDescription},
	{code: "96BWV-AREA-022", name: "96-BWV - Kid's Playground", description: development96VillasLocationDescription},
	{code: "96BWV-AREA-023", name: "96-BWV - Lobby", description: development96VillasLocationDescription},
	{code: "96BWV-AREA-024", name: "96-BWV - Lobby Lounge", description: development96VillasLocationDescription},
	{code: "96BWV-AREA-025", name: "96-BWV - Main kitchen", description: development96VillasLocationDescription},
	{code: "96BWV-AREA-026", name: "96-BWV - Main Pool", description: development96VillasLocationDescription},
	{code: "96BWV-AREA-027", name: "96-BWV - Male Locker", description: development96VillasLocationDescription},
	{code: "96BWV-AREA-028", name: "96-BWV - Outside", description: development96VillasLocationDescription},
	{code: "96BWV-AREA-029", name: "96-BWV - PA Store", description: development96VillasLocationDescription},
	{code: "96BWV-AREA-030", name: "96-BWV - Pastry kitchen", description: development96VillasLocationDescription},
	{code: "96BWV-AREA-031", name: "96-BWV - Spa", description: development96VillasLocationDescription},
	{code: "96BWV-AREA-032", name: "96-BWV - Toilet Gym", description: development96VillasLocationDescription},
	{code: "96BWV-AREA-033", name: "96-BWV - Toilet Hồ bơi phụ", description: development96VillasLocationDescription},
	{code: "96BWV-AREA-034", name: "96-BWV - Toilet Lobby", description: development96VillasLocationDescription},
	{code: "96BWV-AREA-035", name: "96-BWV - Toilet Spa", description: development96VillasLocationDescription},
	{code: "96BWV-AREA-036", name: "96-BWV - Toilet Tropicana", description: development96VillasLocationDescription},
	{code: "96BWV-AREA-037", name: "96-BWV - Tropicana Bar", description: development96VillasLocationDescription},
	{code: "96BWV-AREA-038", name: "96-BWV - Tropicana Kitchen", description: development96VillasLocationDescription},
	{code: "96BWV-AREA-039", name: "96-BWV - Tropicana Restaurant", description: development96VillasLocationDescription},
	{code: "96BWV-AREA-040", name: "96-BWV-Kitchen Office", description: development96VillasLocationDescription},
	{code: "96BWV-AREA-041", name: "96-BWV-Steward", description: development96VillasLocationDescription},
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

	for _, location := range development96VillasLocations {
		if err := seed96VillasLocation(ctx, transaction, location); err != nil {
			_ = transaction.Rollback(ctx)
			return fmt.Errorf("seed 96 Villas location %s: %w", location.code, err)
		}
	}
	for room := 1001; room <= 1099; room++ {
		location := developmentLocation{
			code:        fmt.Sprintf("96BWV-ROOM-%d", room),
			name:        fmt.Sprintf("%d", room),
			description: development96VillasLocationDescription,
		}
		if err := seed96VillasLocation(ctx, transaction, location); err != nil {
			_ = transaction.Rollback(ctx)
			return fmt.Errorf("seed 96 Villas room %d: %w", room, err)
		}
	}

	if err := transaction.Commit(ctx); err != nil {
		return fmt.Errorf("commit development fixtures seed: %w", err)
	}
	return nil
}

func seed96VillasLocation(ctx context.Context, transaction pgx.Tx, location developmentLocation) error {
	result, err := transaction.Exec(ctx, `
INSERT INTO locations (code, name, description, is_active)
VALUES ($1, $2, $3, TRUE)
ON CONFLICT (code) DO UPDATE
SET name = EXCLUDED.name,
    description = EXCLUDED.description,
    is_active = TRUE,
    updated_at = NOW()
WHERE locations.description = $3`, location.code, location.name, location.description)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("96 Villas location code %q collides with non-fixture row", location.code)
	}
	return nil
}
