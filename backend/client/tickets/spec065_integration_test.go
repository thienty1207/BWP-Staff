package tickets_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/thienty1207/BWP-Staff/backend/admin"
)

func TestSPEC065DevelopmentSeedCreatesExactly155BWPLocations(t *testing.T) {
	pool, ctx := openTicketsTestPool(t)
	if err := admin.SeedDevelopmentFixtures(ctx, pool); err != nil {
		t.Fatalf("seed development fixtures: %v", err)
	}

	var count int
	if err := pool.QueryRow(ctx, `
SELECT COUNT(*)
FROM locations
WHERE description = 'Development location seed data: BWP'`).Scan(&count); err != nil {
		t.Fatalf("count BWP locations: %v", err)
	}
	if count != 155 {
		t.Fatalf("BWP fixture count = %d, want 155", count)
	}

	for _, name := range []string{"Nurse Office", "RECEIVING OFFICE", "Sales Room", "Security Office", "Server Room", "Server Room- 96 Villas", "Staffhouse"} {
		var active bool
		if err := pool.QueryRow(ctx, `
SELECT is_active
FROM locations
WHERE description = 'Development location seed data: BWP' AND name = $1`, name).Scan(&active); err != nil {
			t.Fatalf("read BWP location %q: %v", name, err)
		}
		if !active {
			t.Fatalf("BWP location %q is inactive", name)
		}
	}
}

func TestSPEC065DevelopmentSeedMatchesExactBWPCodeNameMap(t *testing.T) {
	pool, ctx := openTicketsTestPool(t)
	if err := admin.SeedDevelopmentFixtures(ctx, pool); err != nil {
		t.Fatalf("seed development fixtures: %v", err)
	}

	expected := map[string]string{
		"BWP-AREA-001": "BOD Office",
		"BWP-AREA-002": "BWP - Asian kitchen",
		"BWP-AREA-003": "BWP - Back Office/FO",
		"BWP-AREA-004": "BWP - Bakery kitchen",
		"BWP-AREA-005": "BWP - Ballroom",
		"BWP-AREA-006": "BWP - Basemant cold storage",
		"BWP-AREA-007": "BWP - Beach Bar",
		"BWP-AREA-008": "BWP - Bell Desk",
		"BWP-AREA-009": "BWP - BOH Essence",
		"BWP-AREA-010": "BWP - Boiler room",
		"BWP-AREA-011": "BWP - BTS room",
		"BWP-AREA-012": "BWP - Buffet counter",
		"BWP-AREA-013": "BWP - Canteen",
		"BWP-AREA-014": "BWP - Capentry work",
		"BWP-AREA-015": "BWP - cold kitchen",
		"BWP-AREA-016": "BWP - Cooling tower",
		"BWP-AREA-017": "BWP - Cview Bar",
		"BWP-AREA-018": "BWP - Cview Kitchen",
		"BWP-AREA-019": "BWP - Eng store 2B",
		"BWP-AREA-020": "BWP - Eng store 2C",
		"BWP-AREA-021": "BWP - Eng store 3B",
		"BWP-AREA-022": "BWP - Eng store 3C",
		"BWP-AREA-023": "BWP - Eng store 4B",
		"BWP-AREA-024": "BWP - Eng store 4C",
		"BWP-AREA-025": "BWP - Eng store 5B",
		"BWP-AREA-026": "BWP - Eng store 5C",
		"BWP-AREA-027": "BWP - Eng store 6B",
		"BWP-AREA-028": "BWP - Eng store 6C",
		"BWP-AREA-029": "BWP - Eng store 7B",
		"BWP-AREA-030": "BWP - Eng store 7C",
		"BWP-AREA-031": "BWP - Eng store 8B",
		"BWP-AREA-032": "BWP - Eng store 8C",
		"BWP-AREA-033": "BWP - Eng store 9B",
		"BWP-AREA-034": "BWP - Eng store 9C",
		"BWP-AREA-035": "BWP - EPS room",
		"BWP-AREA-036": "BWP - Essence Kitchen",
		"BWP-AREA-037": "BWP - Essence restaurant",
		"BWP-AREA-038": "BWP - Essences Bar",
		"BWP-AREA-039": "BWP - European kitchen",
		"BWP-AREA-040": "BWP - Fan room",
		"BWP-AREA-041": "BWP - FB Office",
		"BWP-AREA-042": "BWP - Female Locker",
		"BWP-AREA-043": "BWP - FO Pantry",
		"BWP-AREA-044": "BWP - FO Reception",
		"BWP-AREA-045": "BWP - Generator room",
		"BWP-AREA-046": "BWP - GRO Counter",
		"BWP-AREA-047": "BWP - Guest elevator wing 1",
		"BWP-AREA-048": "BWP - Guest elevator wing 3",
		"BWP-AREA-049": "BWP - Gym",
		"BWP-AREA-050": "BWP - HK OFFICE",
		"BWP-AREA-051": "BWP - HK Pantry Wing 2-2nd Floor",
		"BWP-AREA-052": "BWP - HK Pantry Wing 2-2nd Floorr",
		"BWP-AREA-053": "BWP - HK Pantry Wing 2-3nd Floor",
		"BWP-AREA-054": "BWP - HK Pantry Wing 2-4th Floor",
		"BWP-AREA-055": "BWP - HK Pantry Wing 2-5th Floor",
		"BWP-AREA-056": "BWP - HK Pantry Wing 2-6th Floor",
		"BWP-AREA-057": "BWP - HK Pantry Wing 2-7th Floor",
		"BWP-AREA-058": "BWP - HK Pantry Wing 2-8th Floor",
		"BWP-AREA-059": "BWP - HK Pantry Wing 2-9th Floor",
		"BWP-AREA-060": "BWP - HK Pantry Wing 3-3th Floor",
		"BWP-AREA-061": "BWP - HK Pantry Wing 3-4th Floor",
		"BWP-AREA-062": "BWP - HK Pantry Wing 3-5th Floor",
		"BWP-AREA-063": "BWP - HK Pantry Wing 3-6th Floor",
		"BWP-AREA-064": "BWP - HK Pantry Wing 3-7th Floor",
		"BWP-AREA-065": "BWP - HK Pantry Wing 3-8th Floor",
		"BWP-AREA-066": "BWP - HK Pantry Wing 3-9th Floor",
		"BWP-AREA-067": "BWP - HK Store 2A1",
		"BWP-AREA-068": "BWP - HK Store 2A3",
		"BWP-AREA-069": "BWP - HK Store 3A1",
		"BWP-AREA-070": "BWP - HK Store 3A3",
		"BWP-AREA-071": "BWP - HK Store 4A1",
		"BWP-AREA-072": "BWP - HK Store 4A3",
		"BWP-AREA-073": "BWP - HK Store 5A1",
		"BWP-AREA-074": "BWP - HK Store 5A3",
		"BWP-AREA-075": "BWP - HK Store 6A1",
		"BWP-AREA-076": "BWP - HK Store 6A3",
		"BWP-AREA-077": "BWP - HK Store 7A1",
		"BWP-AREA-078": "BWP - HK Store 7A3",
		"BWP-AREA-079": "BWP - HK Store 8A1",
		"BWP-AREA-080": "BWP - HK Store 8A3",
		"BWP-AREA-081": "BWP - HK Store 9A1",
		"BWP-AREA-082": "BWP - HK Store 9A3",
		"BWP-AREA-083": "BWP - Ice machine room",
		"BWP-AREA-084": "BWP - In front of Ballroom",
		"BWP-AREA-085": "BWP - Kid's Club",
		"BWP-AREA-086": "BWP - Kid's Playground",
		"BWP-AREA-087": "BWP - Kitchen office",
		"BWP-AREA-088": "BWP - Lagoon",
		"BWP-AREA-089": "BWP - Lagoon pump room",
		"BWP-AREA-090": "BWP - Lagoon Swimming Pool",
		"BWP-AREA-091": "BWP - Laundry room",
		"BWP-AREA-092": "BWP - Lobby",
		"BWP-AREA-093": "BWP - Main kitchen",
		"BWP-AREA-094": "BWP - Main Swimming Pool",
		"BWP-AREA-095": "BWP - Mainpool MEP Room",
		"BWP-AREA-096": "BWP - Male Locker",
		"BWP-AREA-097": "BWP - Medium voltage room",
		"BWP-AREA-098": "BWP - Meeting room",
		"BWP-AREA-099": "BWP - Oasis bar",
		"BWP-AREA-100": "BWP - Oasis Bathroom",
		"BWP-AREA-101": "BWP - Oasis pool",
		"BWP-AREA-102": "BWP - Oasis pool bar",
		"BWP-AREA-103": "BWP - Oasis Swimming Pool",
		"BWP-AREA-104": "BWP - Operator Room",
		"BWP-AREA-105": "BWP - Outside Lobby",
		"BWP-AREA-106": "BWP - PA Store-1st Floor",
		"BWP-AREA-107": "BWP - PA Store-M Floor",
		"BWP-AREA-108": "BWP - Pastry kitchen.",
		"BWP-AREA-109": "BWP - PS/DS room",
		"BWP-AREA-110": "BWP - Pump filter room",
		"BWP-AREA-111": "BWP - Recreation Store",
		"BWP-AREA-112": "BWP - RES Office",
		"BWP-AREA-113": "BWP - Romantic Dinner",
		"BWP-AREA-114": "BWP - Rooftop Fan",
		"BWP-AREA-115": "BWP - Spa",
		"BWP-AREA-116": "BWP - Staff restroom",
		"BWP-AREA-117": "BWP - Staffhouse 1",
		"BWP-AREA-118": "BWP - Staffhouse 2",
		"BWP-AREA-119": "BWP - Toilet Ballroom",
		"BWP-AREA-120": "BWP - Toilet Basement",
		"BWP-AREA-121": "BWP - Toilet C view",
		"BWP-AREA-122": "BWP - Toilet Essence",
		"BWP-AREA-123": "BWP - Toilet Lobby",
		"BWP-AREA-124": "BWP - Toilet Oasis",
		"BWP-AREA-125": "BWP - Toilet Staff leve 1",
		"BWP-AREA-126": "BWP - Toilet Staff leve M",
		"BWP-AREA-127": "BWP - Transformer room",
		"BWP-AREA-128": "BWP - Uniform room",
		"BWP-AREA-129": "BWP - Wastewater treatmant room",
		"BWP-AREA-130": "BWP - Workshop",
		"BWP-AREA-131": "BWP (16 Villas & B- 1st floor)",
		"BWP-AREA-132": "BWP Codotel (M- Rooftop floor)",
		"BWP-AREA-133": "BWP- MSB room",
		"BWP-AREA-134": "BWP- Oasis Kitchen",
		"BWP-AREA-135": "BWP-Butchery Area",
		"BWP-AREA-136": "BWP-Kitchen Office",
		"BWP-AREA-137": "BWP-Receiving Area",
		"BWP-AREA-138": "BWP-Recreation Office",
		"BWP-AREA-139": "BWP-Steward Area",
		"BWP-AREA-140": "CCTV Room",
		"BWP-AREA-141": "ENG-OFFICE",
		"BWP-AREA-142": "Executive Office",
		"BWP-AREA-143": "FIN DOCUMENT STORE",
		"BWP-AREA-144": "FIN OFFICE",
		"BWP-AREA-145": "GENERAL STORE",
		"BWP-AREA-146": "HR Office",
		"BWP-AREA-147": "IT Office",
		"BWP-AREA-148": "MasterKey Audit",
		"BWP-AREA-149": "Nurse Office",
		"BWP-AREA-150": "RECEIVING OFFICE",
		"BWP-AREA-151": "Sales Room",
		"BWP-AREA-152": "Security Office",
		"BWP-AREA-153": "Server Room",
		"BWP-AREA-154": "Server Room- 96 Villas",
		"BWP-AREA-155": "Staffhouse",
	}

	rows, err := pool.Query(ctx, `
SELECT code, name, is_active, id
FROM locations
WHERE description = 'Development location seed data: BWP'`)
	if err != nil {
		t.Fatalf("query exact BWP fixture map: %v", err)
	}
	defer rows.Close()
	actual := make(map[string]string, len(expected))
	for rows.Next() {
		var code, name string
		var active bool
		var id int64
		if err := rows.Scan(&code, &name, &active, &id); err != nil {
			t.Fatalf("scan exact BWP fixture row: %v", err)
		}
		if !active || id <= 0 {
			t.Fatalf("BWP fixture row is not active or has invalid ID: code=%q active=%t id=%d", code, active, id)
		}
		actual[code] = name
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("iterate exact BWP fixture map: %v", err)
	}
	if len(actual) != len(expected) {
		t.Fatalf("exact BWP fixture map has %d unique codes, want %d", len(actual), len(expected))
	}
	for code, expectedName := range expected {
		if actual[code] != expectedName {
			t.Fatalf("BWP code %q has name %q, want %q", code, actual[code], expectedName)
		}
	}
}

func TestSPEC065DevelopmentSeedIsIdempotentAndCollisionSafe(t *testing.T) {
	pool, ctx := openTicketsTestPool(t)
	if err := admin.SeedDevelopmentFixtures(ctx, pool); err != nil {
		t.Fatalf("first seed development fixtures: %v", err)
	}

	var fixtureID int64
	if err := pool.QueryRow(ctx, `
SELECT id
FROM locations
WHERE code = 'BWP-AREA-001' AND description = 'Development location seed data: BWP'`).Scan(&fixtureID); err != nil {
		t.Fatalf("read BWP fixture ID: %v", err)
	}
	if err := admin.SeedDevelopmentFixtures(ctx, pool); err != nil {
		t.Fatalf("second seed development fixtures: %v", err)
	}

	var secondID int64
	if err := pool.QueryRow(ctx, `SELECT id FROM locations WHERE code = 'BWP-AREA-001'`).Scan(&secondID); err != nil {
		t.Fatalf("read BWP fixture ID after second seed: %v", err)
	}
	if secondID != fixtureID {
		t.Fatalf("BWP fixture ID changed from %d to %d", fixtureID, secondID)
	}

	if _, err := pool.Exec(ctx, `DELETE FROM locations WHERE code = 'BWP-AREA-001'`); err != nil {
		t.Fatalf("remove BWP fixture before collision setup: %v", err)
	}
	if _, err := pool.Exec(ctx, `
INSERT INTO locations (code, name, description)
VALUES ('BWP-AREA-001', 'Non-fixture collision', 'real location')`); err != nil {
		t.Fatalf("insert non-fixture collision: %v", err)
	}
	if err := admin.SeedDevelopmentFixtures(ctx, pool); err == nil {
		t.Fatal("expected non-fixture BWP code collision to fail")
	}

	var name string
	if err := pool.QueryRow(ctx, `SELECT name FROM locations WHERE code = 'BWP-AREA-001'`).Scan(&name); err != nil {
		t.Fatalf("read collision row: %v", err)
	}
	if name != "Non-fixture collision" {
		t.Fatalf("collision row was overwritten with %q", name)
	}
	var rolledBackCount int
	if err := pool.QueryRow(ctx, `
SELECT COUNT(*)
FROM locations
WHERE description = 'Development location seed data: BWP'`).Scan(&rolledBackCount); err != nil {
		t.Fatalf("count rolled-back BWP rows: %v", err)
	}
	if rolledBackCount != 154 {
		t.Fatalf("collision changed the existing BWP rows after rollback: count=%d want=154", rolledBackCount)
	}
}

func TestSPEC065DevelopmentSeedCollisionRollsBackFreshSeed(t *testing.T) {
	pool, ctx := openTicketsTestPool(t)
	insertLocation(t, pool, ctx, "BWP-AREA-001", "Existing non-fixture location")

	if err := admin.SeedDevelopmentFixtures(ctx, pool); err == nil {
		t.Fatal("expected fresh BWP seed collision to fail")
	}

	var bwpCount, villasCount, departmentCount int
	if err := pool.QueryRow(ctx, `
SELECT (SELECT COUNT(*) FROM locations WHERE description = 'Development location seed data: BWP'),
       (SELECT COUNT(*) FROM locations WHERE description = 'Development location seed data: 96 Villas'),
       (SELECT COUNT(*) FROM departments WHERE description = 'Development department seed data')`).Scan(&bwpCount, &villasCount, &departmentCount); err != nil {
		t.Fatalf("count rolled-back development fixtures: %v", err)
	}
	if bwpCount != 0 || villasCount != 0 || departmentCount != 0 {
		t.Fatalf("fresh collision did not roll back all fixture writes: BWP=%d Villas=%d departments=%d", bwpCount, villasCount, departmentCount)
	}
	var name string
	if err := pool.QueryRow(ctx, `SELECT name FROM locations WHERE code = 'BWP-AREA-001'`).Scan(&name); err != nil {
		t.Fatalf("read preserved collision row: %v", err)
	}
	if name != "Existing non-fixture location" {
		t.Fatalf("collision row changed to %q", name)
	}
}

func TestSPEC065DevelopmentSeedPreserves96VillasAndUnrelatedLocations(t *testing.T) {
	pool, ctx := openTicketsTestPool(t)
	unrelatedID := insertLocation(t, pool, ctx, "SPEC065-UNRELATED", "Unrelated real location")
	if err := admin.SeedDevelopmentFixtures(ctx, pool); err != nil {
		t.Fatalf("first seed development fixtures: %v", err)
	}

	before96 := snapshotLocationsByDescription(t, pool, ctx, "Development location seed data: 96 Villas")
	if len(before96) != 141 {
		t.Fatalf("96 Villas fixture count = %d, want 141", len(before96))
	}
	var fixtureID int64
	if err := pool.QueryRow(ctx, `SELECT id FROM locations WHERE code = 'BWP-AREA-001'`).Scan(&fixtureID); err != nil {
		t.Fatalf("read BWP fixture ID before owned update: %v", err)
	}
	if _, err := pool.Exec(ctx, `UPDATE locations SET name = 'stale BWP name', is_active = FALSE WHERE id = $1`, fixtureID); err != nil {
		t.Fatalf("make owned BWP fixture stale: %v", err)
	}

	if err := admin.SeedDevelopmentFixtures(ctx, pool); err != nil {
		t.Fatalf("second seed development fixtures: %v", err)
	}

	var updatedName string
	var updatedActive bool
	var updatedID int64
	if err := pool.QueryRow(ctx, `SELECT id, name, is_active FROM locations WHERE code = 'BWP-AREA-001'`).Scan(&updatedID, &updatedName, &updatedActive); err != nil {
		t.Fatalf("read owned BWP fixture after update: %v", err)
	}
	if updatedID != fixtureID || updatedName != "BOD Office" || !updatedActive {
		t.Fatalf("owned BWP fixture was not safely refreshed: id=%d name=%q active=%t", updatedID, updatedName, updatedActive)
	}

	after96 := snapshotLocationsByDescription(t, pool, ctx, "Development location seed data: 96 Villas")
	if len(after96) != len(before96) {
		t.Fatalf("96 Villas fixture count changed from %d to %d", len(before96), len(after96))
	}
	for code, name := range before96 {
		if after96[code] != name {
			t.Fatalf("96 Villas location %q changed from %q to %q", code, name, after96[code])
		}
	}
	var unrelatedName string
	if err := pool.QueryRow(ctx, `SELECT name FROM locations WHERE id = $1`, unrelatedID).Scan(&unrelatedName); err != nil {
		t.Fatalf("read unrelated location %d: %v", unrelatedID, err)
	}
	if unrelatedName != "Unrelated real location" {
		t.Fatalf("unrelated location was changed to %q", unrelatedName)
	}
}

func TestSPEC065LocationSearchUsesAuthenticationAndPostgresRanking(t *testing.T) {
	pool, ctx := openTicketsTestPool(t)
	if err := admin.SeedDevelopmentFixtures(ctx, pool); err != nil {
		t.Fatalf("seed development fixtures: %v", err)
	}
	insertLocation(t, pool, ctx, "SPEC065-PRIORITY-EXACT", "Priority")
	insertLocation(t, pool, ctx, "SPEC065-PRIORITY-PREFIX", "Priority Search")
	insertLocation(t, pool, ctx, "SPEC065-PRIORITY-TOKEN", "Desk Priority Panel")
	insertLocation(t, pool, ctx, "SPEC065-PRIORITY-CONTAINS", "Maintenance Panel Priority")
	inactiveID := insertLocation(t, pool, ctx, "SPEC065-PRIORITY-INACTIVE", "Inactive Priority")
	if _, err := pool.Exec(ctx, `UPDATE locations SET is_active = FALSE WHERE id = $1`, inactiveID); err != nil {
		t.Fatalf("deactivate inactive search location: %v", err)
	}

	var departmentID int64
	if err := pool.QueryRow(ctx, `SELECT id FROM departments WHERE code = 'IT'`).Scan(&departmentID); err != nil {
		t.Fatalf("read IT department: %v", err)
	}
	userID := insertUser(t, pool, ctx, "spec065-search-user", "SPEC065-SEARCH", "SPEC-06.5 Search User", departmentID)
	token := insertSession(t, pool, ctx, userID)
	server := newTicketsApp(pool)

	unauthenticated := requestTickets(t, server, "/api/v1/locations?q=server&limit=10", "")
	unauthenticated.Body.Close()
	if unauthenticated.StatusCode != http.StatusUnauthorized {
		t.Fatalf("unauthenticated location search: expected 401, got %d", unauthenticated.StatusCode)
	}

	response := requestTickets(t, server, "/api/v1/locations?q=%20SeRvEr%20&limit=10", token)
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		t.Fatalf("location search: expected 200, got %d", response.StatusCode)
	}
	var body struct {
		Locations []spec06LookupLocation `json:"locations"`
	}
	decodeTicketResponse(t, response, &body)
	if len(body.Locations) < 2 || body.Locations[0].Name != "Server Room" || body.Locations[1].Name != "Server Room- 96 Villas" {
		t.Fatalf("unexpected server search ranking: %+v", body.Locations)
	}
	if len(body.Locations) > 10 {
		t.Fatalf("default search limit returned %d rows, want at most 10", len(body.Locations))
	}
	seenIDs := make(map[int64]struct{}, len(body.Locations))
	for _, location := range body.Locations {
		if location.ID <= 0 {
			t.Fatalf("location has invalid real ID: %+v", location)
		}
		if _, exists := seenIDs[location.ID]; exists {
			t.Fatalf("location search returned duplicate ID %d", location.ID)
		}
		seenIDs[location.ID] = struct{}{}
	}

	searchCases := []struct {
		path             string
		mustIncludeNames []string
		wantFirst        string
		wantCount        int
	}{
		{path: "/api/v1/locations?q=96%20v&limit=30", mustIncludeNames: []string{"96 Villas", "Server Room- 96 Villas"}, wantFirst: "96 Villas", wantCount: -1},
		{path: "/api/v1/locations?q=OASIS&limit=30", mustIncludeNames: []string{"BWP - Oasis bar", "BWP - Oasis Bathroom", "BWP - Oasis pool", "BWP - Oasis pool bar", "BWP - Oasis Swimming Pool", "BWP- Oasis Kitchen", "BWP - Toilet Oasis"}, wantCount: -1},
		{path: "/api/v1/locations?q=bwp%20-%20lobby&limit=10", mustIncludeNames: []string{"BWP - Lobby"}, wantFirst: "BWP - Lobby", wantCount: 1},
		{path: "/api/v1/locations?q=does-not-exist&limit=30", wantCount: 0},
		{path: "/api/v1/locations?q=bwp", wantCount: 10},
	}
	for _, searchCase := range searchCases {
		searchResponse := requestTickets(t, server, searchCase.path, token)
		if searchResponse.StatusCode != http.StatusOK {
			searchResponse.Body.Close()
			t.Fatalf("%s returned status %d, want 200", searchCase.path, searchResponse.StatusCode)
		}
		var searchBody struct {
			Locations []spec06LookupLocation `json:"locations"`
		}
		decodeTicketResponse(t, searchResponse, &searchBody)
		searchResponse.Body.Close()
		if searchCase.wantCount >= 0 && len(searchBody.Locations) != searchCase.wantCount {
			t.Fatalf("%s returned %d rows, want %d", searchCase.path, len(searchBody.Locations), searchCase.wantCount)
		}
		if searchCase.wantFirst != "" && (len(searchBody.Locations) == 0 || searchBody.Locations[0].Name != searchCase.wantFirst) {
			t.Fatalf("%s first result = %+v, want %q", searchCase.path, searchBody.Locations, searchCase.wantFirst)
		}
		for _, expectedName := range searchCase.mustIncludeNames {
			found := false
			for _, location := range searchBody.Locations {
				if location.Name == expectedName {
					found = true
					break
				}
			}
			if !found {
				t.Fatalf("%s did not return expected location %q: %+v", searchCase.path, expectedName, searchBody.Locations)
			}
		}
	}

	priorityResponse := requestTickets(t, server, "/api/v1/locations?q=priority&limit=30", token)
	defer priorityResponse.Body.Close()
	if priorityResponse.StatusCode != http.StatusOK {
		t.Fatalf("priority search: expected 200, got %d", priorityResponse.StatusCode)
	}
	var priorityBody struct {
		Locations []spec06LookupLocation `json:"locations"`
	}
	decodeTicketResponse(t, priorityResponse, &priorityBody)
	if len(priorityBody.Locations) != 4 {
		t.Fatalf("unexpected priority search count: %+v", priorityBody.Locations)
	}
	expectedPriorityNames := []string{"Priority", "Priority Search", "Desk Priority Panel", "Maintenance Panel Priority"}
	for index, expectedName := range expectedPriorityNames {
		if priorityBody.Locations[index].Name != expectedName {
			t.Fatalf("priority search rank %d = %q, want %q", index, priorityBody.Locations[index].Name, expectedName)
		}
	}

	limitedResponse := requestTickets(t, server, "/api/v1/locations?q=priority&limit=1", token)
	defer limitedResponse.Body.Close()
	if limitedResponse.StatusCode != http.StatusOK {
		t.Fatalf("limited location search: expected 200, got %d", limitedResponse.StatusCode)
	}
	var limitedBody struct {
		Locations []spec06LookupLocation `json:"locations"`
	}
	decodeTicketResponse(t, limitedResponse, &limitedBody)
	if len(limitedBody.Locations) != 1 || limitedBody.Locations[0].Name != "Priority" {
		t.Fatalf("unexpected limited location search: %+v", limitedBody.Locations)
	}

	invalidLimit := requestTickets(t, server, "/api/v1/locations?limit=31", token)
	invalidLimit.Body.Close()
	if invalidLimit.StatusCode != http.StatusBadRequest {
		t.Fatalf("invalid location limit: expected 400, got %d", invalidLimit.StatusCode)
	}
}

func snapshotLocationsByDescription(t *testing.T, pool *pgxpool.Pool, ctx context.Context, description string) map[string]string {
	t.Helper()
	rows, err := pool.Query(ctx, `
SELECT code, name
FROM locations
WHERE description = $1`, description)
	if err != nil {
		t.Fatalf("snapshot locations for %q: %v", description, err)
	}
	defer rows.Close()
	snapshot := make(map[string]string)
	for rows.Next() {
		var code, name string
		if err := rows.Scan(&code, &name); err != nil {
			t.Fatalf("scan location snapshot for %q: %v", description, err)
		}
		snapshot[code] = name
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("iterate location snapshot for %q: %v", description, err)
	}
	return snapshot
}
