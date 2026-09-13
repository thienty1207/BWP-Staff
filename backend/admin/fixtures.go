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
const developmentBWPLocationDescription = "Development location seed data: BWP"

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

var developmentBWPLocations = []developmentLocation{
	{code: "BWP-AREA-001", name: "BOD Office", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-002", name: "BWP - Asian kitchen", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-003", name: "BWP - Back Office/FO", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-004", name: "BWP - Bakery kitchen", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-005", name: "BWP - Ballroom", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-006", name: "BWP - Basemant cold storage", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-007", name: "BWP - Beach Bar", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-008", name: "BWP - Bell Desk", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-009", name: "BWP - BOH Essence", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-010", name: "BWP - Boiler room", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-011", name: "BWP - BTS room", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-012", name: "BWP - Buffet counter", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-013", name: "BWP - Canteen", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-014", name: "BWP - Capentry work", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-015", name: "BWP - cold kitchen", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-016", name: "BWP - Cooling tower", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-017", name: "BWP - Cview Bar", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-018", name: "BWP - Cview Kitchen", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-019", name: "BWP - Eng store 2B", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-020", name: "BWP - Eng store 2C", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-021", name: "BWP - Eng store 3B", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-022", name: "BWP - Eng store 3C", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-023", name: "BWP - Eng store 4B", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-024", name: "BWP - Eng store 4C", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-025", name: "BWP - Eng store 5B", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-026", name: "BWP - Eng store 5C", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-027", name: "BWP - Eng store 6B", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-028", name: "BWP - Eng store 6C", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-029", name: "BWP - Eng store 7B", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-030", name: "BWP - Eng store 7C", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-031", name: "BWP - Eng store 8B", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-032", name: "BWP - Eng store 8C", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-033", name: "BWP - Eng store 9B", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-034", name: "BWP - Eng store 9C", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-035", name: "BWP - EPS room", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-036", name: "BWP - Essence Kitchen", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-037", name: "BWP - Essence restaurant", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-038", name: "BWP - Essences Bar", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-039", name: "BWP - European kitchen", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-040", name: "BWP - Fan room", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-041", name: "BWP - FB Office", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-042", name: "BWP - Female Locker", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-043", name: "BWP - FO Pantry", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-044", name: "BWP - FO Reception", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-045", name: "BWP - Generator room", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-046", name: "BWP - GRO Counter", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-047", name: "BWP - Guest elevator wing 1", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-048", name: "BWP - Guest elevator wing 3", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-049", name: "BWP - Gym", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-050", name: "BWP - HK OFFICE", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-051", name: "BWP - HK Pantry Wing 2-2nd Floor", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-052", name: "BWP - HK Pantry Wing 2-2nd Floorr", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-053", name: "BWP - HK Pantry Wing 2-3nd Floor", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-054", name: "BWP - HK Pantry Wing 2-4th Floor", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-055", name: "BWP - HK Pantry Wing 2-5th Floor", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-056", name: "BWP - HK Pantry Wing 2-6th Floor", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-057", name: "BWP - HK Pantry Wing 2-7th Floor", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-058", name: "BWP - HK Pantry Wing 2-8th Floor", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-059", name: "BWP - HK Pantry Wing 2-9th Floor", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-060", name: "BWP - HK Pantry Wing 3-3th Floor", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-061", name: "BWP - HK Pantry Wing 3-4th Floor", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-062", name: "BWP - HK Pantry Wing 3-5th Floor", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-063", name: "BWP - HK Pantry Wing 3-6th Floor", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-064", name: "BWP - HK Pantry Wing 3-7th Floor", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-065", name: "BWP - HK Pantry Wing 3-8th Floor", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-066", name: "BWP - HK Pantry Wing 3-9th Floor", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-067", name: "BWP - HK Store 2A1", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-068", name: "BWP - HK Store 2A3", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-069", name: "BWP - HK Store 3A1", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-070", name: "BWP - HK Store 3A3", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-071", name: "BWP - HK Store 4A1", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-072", name: "BWP - HK Store 4A3", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-073", name: "BWP - HK Store 5A1", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-074", name: "BWP - HK Store 5A3", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-075", name: "BWP - HK Store 6A1", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-076", name: "BWP - HK Store 6A3", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-077", name: "BWP - HK Store 7A1", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-078", name: "BWP - HK Store 7A3", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-079", name: "BWP - HK Store 8A1", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-080", name: "BWP - HK Store 8A3", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-081", name: "BWP - HK Store 9A1", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-082", name: "BWP - HK Store 9A3", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-083", name: "BWP - Ice machine room", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-084", name: "BWP - In front of Ballroom", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-085", name: "BWP - Kid's Club", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-086", name: "BWP - Kid's Playground", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-087", name: "BWP - Kitchen office", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-088", name: "BWP - Lagoon", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-089", name: "BWP - Lagoon pump room", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-090", name: "BWP - Lagoon Swimming Pool", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-091", name: "BWP - Laundry room", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-092", name: "BWP - Lobby", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-093", name: "BWP - Main kitchen", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-094", name: "BWP - Main Swimming Pool", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-095", name: "BWP - Mainpool MEP Room", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-096", name: "BWP - Male Locker", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-097", name: "BWP - Medium voltage room", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-098", name: "BWP - Meeting room", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-099", name: "BWP - Oasis bar", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-100", name: "BWP - Oasis Bathroom", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-101", name: "BWP - Oasis pool", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-102", name: "BWP - Oasis pool bar", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-103", name: "BWP - Oasis Swimming Pool", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-104", name: "BWP - Operator Room", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-105", name: "BWP - Outside Lobby", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-106", name: "BWP - PA Store-1st Floor", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-107", name: "BWP - PA Store-M Floor", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-108", name: "BWP - Pastry kitchen.", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-109", name: "BWP - PS/DS room", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-110", name: "BWP - Pump filter room", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-111", name: "BWP - Recreation Store", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-112", name: "BWP - RES Office", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-113", name: "BWP - Romantic Dinner", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-114", name: "BWP - Rooftop Fan", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-115", name: "BWP - Spa", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-116", name: "BWP - Staff restroom", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-117", name: "BWP - Staffhouse 1", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-118", name: "BWP - Staffhouse 2", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-119", name: "BWP - Toilet Ballroom", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-120", name: "BWP - Toilet Basement", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-121", name: "BWP - Toilet C view", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-122", name: "BWP - Toilet Essence", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-123", name: "BWP - Toilet Lobby", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-124", name: "BWP - Toilet Oasis", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-125", name: "BWP - Toilet Staff leve 1", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-126", name: "BWP - Toilet Staff leve M", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-127", name: "BWP - Transformer room", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-128", name: "BWP - Uniform room", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-129", name: "BWP - Wastewater treatmant room", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-130", name: "BWP - Workshop", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-131", name: "BWP (16 Villas & B- 1st floor)", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-132", name: "BWP Codotel (M- Rooftop floor)", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-133", name: "BWP- MSB room", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-134", name: "BWP- Oasis Kitchen", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-135", name: "BWP-Butchery Area", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-136", name: "BWP-Kitchen Office", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-137", name: "BWP-Receiving Area", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-138", name: "BWP-Recreation Office", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-139", name: "BWP-Steward Area", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-140", name: "CCTV Room", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-141", name: "ENG-OFFICE", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-142", name: "Executive Office", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-143", name: "FIN DOCUMENT STORE", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-144", name: "FIN OFFICE", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-145", name: "GENERAL STORE", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-146", name: "HR Office", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-147", name: "IT Office", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-148", name: "MasterKey Audit", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-149", name: "Nurse Office", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-150", name: "RECEIVING OFFICE", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-151", name: "Sales Room", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-152", name: "Security Office", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-153", name: "Server Room", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-154", name: "Server Room- 96 Villas", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-155", name: "Staffhouse", description: developmentBWPLocationDescription},
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
	for _, location := range developmentBWPLocations {
		if err := seedBWPLocation(ctx, transaction, location); err != nil {
			_ = transaction.Rollback(ctx)
			return fmt.Errorf("seed BWP location %s: %w", location.code, err)
		}
	}

	if err := transaction.Commit(ctx); err != nil {
		return fmt.Errorf("commit development fixtures seed: %w", err)
	}
	return nil
}

func seed96VillasLocation(ctx context.Context, transaction pgx.Tx, location developmentLocation) error {
	return seedMarkedLocation(ctx, transaction, location, "96 Villas")
}

func seedBWPLocation(ctx context.Context, transaction pgx.Tx, location developmentLocation) error {
	return seedMarkedLocation(ctx, transaction, location, "BWP")
}

func seedMarkedLocation(ctx context.Context, transaction pgx.Tx, location developmentLocation, fixtureName string) error {
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
		return fmt.Errorf("%s location code %q collides with non-fixture row", fixtureName, location.code)
	}
	return nil
}
