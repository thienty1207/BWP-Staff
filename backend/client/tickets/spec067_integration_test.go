package tickets_test

import (
	"net/http"
	"strconv"
	"testing"

	"github.com/thienty1207/BWP-Staff/backend/admin"
)

const spec067BWPRoomsDescription = "Development location seed data: BWP Rooms"

func TestSPEC067DevelopmentSeedMatchesExactBWPRoomSet(t *testing.T) {
	pool, ctx := openTicketsTestPool(t)
	if err := admin.SeedDevelopmentFixtures(ctx, pool); err != nil {
		t.Fatalf("seed development fixtures: %v", err)
	}

	expectedRooms := expectedSPEC067BWPRoomNumbers()
	if len(expectedRooms) != 564 {
		t.Fatalf("test fixture contains %d approved room numbers, want 564", len(expectedRooms))
	}

	var count, distinctCodes int
	if err := pool.QueryRow(ctx, `
SELECT COUNT(*), COUNT(DISTINCT code)
FROM locations
WHERE description = $1`, spec067BWPRoomsDescription).Scan(&count, &distinctCodes); err != nil {
		t.Fatalf("count BWP Rooms fixtures: %v", err)
	}
	if count != 564 || distinctCodes != 564 {
		t.Fatalf("BWP Rooms fixture count/distinct codes = %d/%d, want 564/564", count, distinctCodes)
	}

	rows, err := pool.Query(ctx, `
SELECT code, name, is_active, id
FROM locations
WHERE description = $1`, spec067BWPRoomsDescription)
	if err != nil {
		t.Fatalf("query BWP Rooms fixtures: %v", err)
	}
	defer rows.Close()

	type roomRow struct {
		name   string
		active bool
		id     int64
	}
	actual := make(map[string]roomRow, 564)
	for rows.Next() {
		var code, name string
		var active bool
		var id int64
		if err := rows.Scan(&code, &name, &active, &id); err != nil {
			t.Fatalf("scan BWP Rooms fixture: %v", err)
		}
		if _, exists := actual[code]; exists {
			t.Fatalf("duplicate BWP Rooms code %q", code)
		}
		actual[code] = roomRow{name: name, active: active, id: id}
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("iterate BWP Rooms fixtures: %v", err)
	}

	expectedByFloor := map[byte]int{'2': 72, '3': 75, '4': 75, '5': 65, '6': 60, '7': 75, '8': 76, '9': 66}
	actualByFloor := make(map[byte]int, len(expectedByFloor))
	for _, roomNumber := range expectedRooms {
		name := strconv.Itoa(roomNumber)
		code := "BWP-ROOM-" + name
		row, exists := actual[code]
		if !exists {
			t.Fatalf("approved BWP room %d is missing", roomNumber)
		}
		if row.name != name {
			t.Fatalf("BWP room %s has visible name %q, want %q", code, row.name, name)
		}
		if !row.active || row.id <= 0 {
			t.Fatalf("BWP room %s is inactive or has invalid ID: active=%t id=%d", code, row.active, row.id)
		}
		actualByFloor[name[0]]++
	}
	if len(actual) != len(expectedRooms) {
		t.Fatalf("BWP Rooms has %d unique rows, want %d", len(actual), len(expectedRooms))
	}
	for code := range actual {
		if len(code) < len("BWP-ROOM-") || code[:len("BWP-ROOM-")] != "BWP-ROOM-" {
			t.Fatalf("unexpected BWP Rooms code %q", code)
		}
		if _, err := strconv.Atoi(code[len("BWP-ROOM-"):]); err != nil {
			t.Fatalf("BWP Rooms code %q does not contain a numeric room: %v", code, err)
		}
	}
	for floor, want := range expectedByFloor {
		if actualByFloor[floor] != want {
			t.Fatalf("BWP %cxxx room count = %d, want %d", floor, actualByFloor[floor], want)
		}
	}

	for code, row := range actual {
		if row.name == "2013" || row.name == "3013" || row.name == "4013" || row.name == "5013" || row.name == "6013" || row.name == "7021" || row.name == "8021" || row.name == "9020" {
			t.Fatalf("unapproved inferred room %q was seeded as %q", row.name, code)
		}
	}
}

func TestSPEC067DevelopmentSeedIsIdempotentAndPreservesExistingData(t *testing.T) {
	pool, ctx := openTicketsTestPool(t)
	unrelatedID := insertLocation(t, pool, ctx, "SPEC067-UNRELATED", "SPEC-06.7 unrelated location")
	if err := admin.SeedDevelopmentFixtures(ctx, pool); err != nil {
		t.Fatalf("first seed development fixtures: %v", err)
	}

	bwpAreasBefore := snapshotLocationsByDescription(t, pool, ctx, "Development location seed data: BWP")
	if len(bwpAreasBefore) != 155 {
		t.Fatalf("BWP area fixture count = %d, want 155", len(bwpAreasBefore))
	}
	villasBefore := snapshotLocationsByDescription(t, pool, ctx, "Development location seed data: 96 Villas")
	if len(villasBefore) != 141 {
		t.Fatalf("96 Villas fixture count = %d, want 141", len(villasBefore))
	}

	var roomID int64
	if err := pool.QueryRow(ctx, `SELECT id FROM locations WHERE code = 'BWP-ROOM-2001'`).Scan(&roomID); err != nil {
		t.Fatalf("read BWP room ID: %v", err)
	}
	if _, err := pool.Exec(ctx, `
UPDATE locations
SET name = 'stale room name', is_active = FALSE
WHERE id = $1`, roomID); err != nil {
		t.Fatalf("make owned BWP room stale: %v", err)
	}

	if err := admin.SeedDevelopmentFixtures(ctx, pool); err != nil {
		t.Fatalf("second seed development fixtures: %v", err)
	}

	var refreshedID int64
	var refreshedName string
	var refreshedActive bool
	if err := pool.QueryRow(ctx, `
SELECT id, name, is_active
FROM locations
WHERE code = 'BWP-ROOM-2001'`).Scan(&refreshedID, &refreshedName, &refreshedActive); err != nil {
		t.Fatalf("read refreshed BWP room: %v", err)
	}
	if refreshedID != roomID || refreshedName != "2001" || !refreshedActive {
		t.Fatalf("owned BWP room was not refreshed safely: id=%d name=%q active=%t", refreshedID, refreshedName, refreshedActive)
	}

	var roomCount, distinctRoomCodes int
	if err := pool.QueryRow(ctx, `
SELECT COUNT(*), COUNT(DISTINCT code)
FROM locations
WHERE description = $1`, spec067BWPRoomsDescription).Scan(&roomCount, &distinctRoomCodes); err != nil {
		t.Fatalf("count BWP Rooms after second seed: %v", err)
	}
	if roomCount != 564 || distinctRoomCodes != 564 {
		t.Fatalf("BWP Rooms after second seed = %d/%d, want 564/564", roomCount, distinctRoomCodes)
	}

	bwpAreasAfter := snapshotLocationsByDescription(t, pool, ctx, "Development location seed data: BWP")
	assertLocationSnapshotEqual(t, "BWP areas", bwpAreasBefore, bwpAreasAfter)
	villasAfter := snapshotLocationsByDescription(t, pool, ctx, "Development location seed data: 96 Villas")
	assertLocationSnapshotEqual(t, "96 Villas", villasBefore, villasAfter)

	var unrelatedName string
	if err := pool.QueryRow(ctx, `SELECT name FROM locations WHERE id = $1`, unrelatedID).Scan(&unrelatedName); err != nil {
		t.Fatalf("read unrelated location: %v", err)
	}
	if unrelatedName != "SPEC-06.7 unrelated location" {
		t.Fatalf("unrelated location changed to %q", unrelatedName)
	}
}

func TestSPEC067DevelopmentSeedCollisionRollsBackAllWrites(t *testing.T) {
	pool, ctx := openTicketsTestPool(t)
	collisionID := insertLocation(t, pool, ctx, "BWP-ROOM-2014", "Existing non-fixture room")

	if err := admin.SeedDevelopmentFixtures(ctx, pool); err == nil {
		t.Fatal("expected non-fixture BWP Rooms code collision to fail")
	}

	var bwpRooms, bwpAreas, villas, departments int
	if err := pool.QueryRow(ctx, `
SELECT
    (SELECT COUNT(*) FROM locations WHERE description = $1),
    (SELECT COUNT(*) FROM locations WHERE description = 'Development location seed data: BWP'),
    (SELECT COUNT(*) FROM locations WHERE description = 'Development location seed data: 96 Villas'),
    (SELECT COUNT(*) FROM departments WHERE description = 'Development department seed data')`, spec067BWPRoomsDescription).Scan(&bwpRooms, &bwpAreas, &villas, &departments); err != nil {
		t.Fatalf("count rolled-back fixtures: %v", err)
	}
	if bwpRooms != 0 || bwpAreas != 0 || villas != 0 || departments != 0 {
		t.Fatalf("collision left partial fixture writes: rooms=%d areas=%d villas=%d departments=%d", bwpRooms, bwpAreas, villas, departments)
	}

	var preservedID int64
	var preservedName string
	if err := pool.QueryRow(ctx, `SELECT id, name FROM locations WHERE code = 'BWP-ROOM-2014'`).Scan(&preservedID, &preservedName); err != nil {
		t.Fatalf("read collision row: %v", err)
	}
	if preservedID != collisionID || preservedName != "Existing non-fixture room" {
		t.Fatalf("collision row changed: id=%d name=%q", preservedID, preservedName)
	}
}

func TestSPEC067BWPRoomsAreFindableThroughAuthenticatedLocationLookup(t *testing.T) {
	pool, ctx := openTicketsTestPool(t)
	if err := admin.SeedDevelopmentFixtures(ctx, pool); err != nil {
		t.Fatalf("seed development fixtures: %v", err)
	}
	departmentID := insertDepartment(t, pool, ctx, "SPEC067-LOOKUP", "SPEC-06.7 Lookup Department")
	userID := insertUser(t, pool, ctx, "spec067-lookup-user", "SPEC067-LOOKUP-USER", "SPEC-06.7 Lookup User", departmentID)
	token := insertSession(t, pool, ctx, userID)
	server := newTicketsApp(pool)

	unauthenticated := requestTickets(t, server, "/api/v1/locations?q=2001", "")
	unauthenticated.Body.Close()
	if unauthenticated.StatusCode != http.StatusUnauthorized {
		t.Fatalf("unauthenticated location lookup: expected 401, got %d", unauthenticated.StatusCode)
	}

	for _, roomNumber := range []int{2001, 2126, 2319, 3001, 3321, 4001, 4321, 5001, 5321, 6001, 6666, 7001, 7321, 8001, 8888, 9001, 9321, 9999} {
		path := "/api/v1/locations?q=" + strconv.Itoa(roomNumber)
		response := requestTickets(t, server, path, token)
		if response.StatusCode != http.StatusOK {
			response.Body.Close()
			t.Fatalf("%s returned status %d, want 200", path, response.StatusCode)
		}
		var body struct {
			Locations []spec06LookupLocation `json:"locations"`
		}
		decodeTicketResponse(t, response, &body)
		response.Body.Close()

		wantName := strconv.Itoa(roomNumber)
		wantCode := "BWP-ROOM-" + wantName
		found := false
		for _, location := range body.Locations {
			if location.Code != nil && *location.Code == wantCode {
				if location.ID <= 0 || location.Name != wantName {
					t.Fatalf("%s returned invalid room: %+v", path, location)
				}
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("%s did not return %s with numeric name: %+v", path, wantCode, body.Locations)
		}
	}

	for _, absentRoom := range []int{2013, 3013, 4013, 5013, 6013, 7021, 8021, 9020} {
		path := "/api/v1/locations?q=" + strconv.Itoa(absentRoom)
		response := requestTickets(t, server, path, token)
		if response.StatusCode != http.StatusOK {
			response.Body.Close()
			t.Fatalf("absent-room %s returned status %d, want 200", path, response.StatusCode)
		}
		var body struct {
			Locations []spec06LookupLocation `json:"locations"`
		}
		decodeTicketResponse(t, response, &body)
		response.Body.Close()
		unexpectedCode := "BWP-ROOM-" + strconv.Itoa(absentRoom)
		for _, location := range body.Locations {
			if location.Code != nil && *location.Code == unexpectedCode {
				t.Fatalf("absent room %d was returned as %q", absentRoom, *location.Code)
			}
		}
	}
}

func expectedSPEC067BWPRoomNumbers() []int {
	return []int{
		// 2xxx: 72
		2001, 2002, 2003, 2004, 2005, 2006, 2007, 2008, 2009, 2010, 2011, 2012, 2014, 2015, 2016, 2017, 2018, 2020, 2022, 2024, 2026, 2028,
		2100, 2101, 2102, 2103, 2104, 2105, 2106, 2107, 2108, 2109, 2110, 2111, 2112, 2114, 2115, 2116, 2117, 2118, 2120, 2122, 2124, 2126,
		2200, 2201, 2202, 2203, 2204, 2205, 2206, 2207, 2208, 2209, 2211, 2215, 2217,
		2300, 2301, 2302, 2303, 2304, 2305, 2306, 2307, 2308, 2309, 2310, 2311, 2315, 2317, 2319,

		// 3xxx: 75
		3001, 3002, 3003, 3004, 3005, 3006, 3007, 3008, 3009, 3010, 3011, 3012, 3014, 3015, 3016, 3017, 3018, 3020, 3022, 3024, 3026, 3028,
		3100, 3101, 3102, 3103, 3104, 3105, 3106, 3107, 3108, 3109, 3110, 3111, 3112, 3114, 3115, 3116, 3117, 3118, 3120, 3122, 3124, 3126,
		3200, 3201, 3202, 3203, 3204, 3205, 3206, 3207, 3208, 3209, 3211, 3215, 3217, 3219, 3221,
		3300, 3301, 3302, 3303, 3304, 3305, 3306, 3307, 3308, 3309, 3310, 3311, 3315, 3317, 3319, 3321,

		// 4xxx: 75
		4001, 4002, 4003, 4004, 4005, 4006, 4007, 4008, 4009, 4010, 4011, 4012, 4014, 4015, 4016, 4017, 4018, 4020, 4022, 4024, 4026, 4028,
		4100, 4101, 4102, 4103, 4104, 4105, 4106, 4107, 4108, 4109, 4110, 4111, 4112, 4114, 4115, 4116, 4117, 4118, 4120, 4122, 4124, 4126,
		4200, 4201, 4202, 4203, 4204, 4205, 4206, 4207, 4208, 4209, 4211, 4215, 4217, 4219, 4221,
		4300, 4301, 4302, 4303, 4304, 4305, 4306, 4307, 4308, 4309, 4310, 4311, 4315, 4317, 4319, 4321,

		// 5xxx: 65
		5001, 5002, 5003, 5004, 5005, 5006, 5007, 5008, 5009, 5010, 5011, 5012, 5014, 5015, 5016, 5017, 5018, 5020, 5022, 5024, 5026, 5028,
		5100, 5101, 5102, 5103, 5104, 5105, 5106, 5107, 5108, 5109, 5110, 5111, 5112, 5114, 5115, 5116, 5117, 5118, 5120, 5122, 5124,
		5209, 5211, 5215, 5217, 5219, 5221,
		5300, 5301, 5302, 5303, 5304, 5305, 5306, 5307, 5308, 5309, 5310, 5311, 5315, 5317, 5319, 5321,

		// 6xxx: 60
		6001, 6002, 6003, 6004, 6005, 6006, 6007, 6008, 6009, 6010, 6011, 6012, 6014, 6015, 6016, 6017, 6019,
		6100, 6101, 6102, 6103, 6104, 6105, 6106, 6107, 6108, 6109, 6110, 6111, 6112, 6114, 6115, 6117, 6119,
		6200, 6201, 6202, 6203, 6205, 6207, 6209, 6211, 6215, 6217, 6219, 6221,
		6300, 6301, 6302, 6303, 6304, 6305, 6307, 6309, 6311, 6315, 6317, 6319, 6321, 6666,

		// 7xxx: 75
		7001, 7002, 7003, 7004, 7005, 7006, 7007, 7008, 7009, 7010, 7011, 7012, 7014, 7015, 7016, 7017, 7018, 7019, 7020, 7022, 7024, 7026,
		7100, 7101, 7102, 7103, 7104, 7105, 7106, 7107, 7108, 7109, 7110, 7111, 7112, 7114, 7115, 7116, 7117, 7118, 7119, 7120, 7122, 7124,
		7200, 7201, 7202, 7203, 7204, 7205, 7206, 7207, 7208, 7209, 7211, 7215, 7217, 7219, 7221,
		7300, 7301, 7302, 7303, 7304, 7305, 7306, 7307, 7308, 7309, 7310, 7311, 7315, 7317, 7319, 7321,

		// 8xxx: 76
		8001, 8002, 8003, 8004, 8005, 8006, 8007, 8008, 8009, 8010, 8011, 8012, 8014, 8015, 8016, 8017, 8018, 8019, 8020, 8022, 8024, 8026,
		8100, 8101, 8102, 8103, 8104, 8105, 8106, 8107, 8108, 8109, 8110, 8111, 8112, 8114, 8115, 8116, 8117, 8118, 8119, 8120, 8122, 8124,
		8200, 8201, 8202, 8203, 8204, 8205, 8206, 8207, 8208, 8209, 8211, 8215, 8217, 8219, 8221,
		8300, 8301, 8302, 8303, 8304, 8305, 8306, 8307, 8308, 8309, 8310, 8311, 8315, 8317, 8319, 8321, 8888,

		// 9xxx: 66
		9001, 9002, 9003, 9004, 9005, 9006, 9007, 9008, 9009, 9010, 9011, 9012, 9014, 9015, 9017,
		9100, 9101, 9102, 9103, 9104, 9105, 9106, 9107, 9108, 9109, 9110, 9111,
		9200, 9201, 9202, 9203, 9205, 9207, 9209, 9211, 9215, 9217, 9219, 9221,
		9300, 9301, 9302, 9303, 9304, 9305, 9307, 9309, 9311, 9315, 9317, 9319, 9321,
		9501, 9502, 9503, 9504, 9505, 9506, 9507, 9508, 9509, 9510, 9511,
		9601, 9602, 9999,
	}
}

func assertLocationSnapshotEqual(t *testing.T, label string, before, after map[string]string) {
	t.Helper()
	if len(after) != len(before) {
		t.Fatalf("%s changed from %d rows to %d", label, len(before), len(after))
	}
	for code, beforeName := range before {
		if after[code] != beforeName {
			t.Fatalf("%s location %q changed from %q to %q", label, code, beforeName, after[code])
		}
	}
}
