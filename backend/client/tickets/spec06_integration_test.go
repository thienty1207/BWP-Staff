package tickets_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/thienty1207/BWP-Staff/backend/admin"
)

type spec06LookupDepartment struct {
	ID   int64  `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
}

type spec06LookupLocation struct {
	ID   int64   `json:"id"`
	Code *string `json:"code"`
	Name string  `json:"name"`
}

func TestSPEC06LookupEndpointsRequireAuthAndReturnActiveRowsInStableOrder(t *testing.T) {
	pool, ctx := openTicketsTestPool(t)
	activeDepartmentFirst := insertDepartment(t, pool, ctx, "SPEC06-A", "SPEC06 Alpha")
	insertDepartment(t, pool, ctx, "SPEC06-I", "SPEC06 Hidden")
	activeDepartmentLast := insertDepartment(t, pool, ctx, "SPEC06-Z", "SPEC06 Omega")
	if _, err := pool.Exec(ctx, `UPDATE departments SET is_active = FALSE WHERE code = 'SPEC06-I'`); err != nil {
		t.Fatalf("deactivate lookup department: %v", err)
	}

	activeLocationFirst := insertLocation(t, pool, ctx, "SPEC06-LOBBY", "SPEC06 Lobby")
	var nullableLocationID int64
	if err := pool.QueryRow(ctx, `
INSERT INTO locations (code, name)
VALUES (NULL, 'SPEC06 No Code')
RETURNING id`).Scan(&nullableLocationID); err != nil {
		t.Fatalf("insert nullable lookup location: %v", err)
	}
	insertLocation(t, pool, ctx, "SPEC06-HIDDEN", "SPEC06 Hidden Location")
	if _, err := pool.Exec(ctx, `UPDATE locations SET is_active = FALSE WHERE code = 'SPEC06-HIDDEN'`); err != nil {
		t.Fatalf("deactivate lookup location: %v", err)
	}

	userID := insertUser(t, pool, ctx, "spec06-lookup-user", "SPEC06-LOOKUP", "SPEC-06 Lookup User", activeDepartmentFirst)
	token := insertSession(t, pool, ctx, userID)
	server := newTicketsApp(pool)

	for _, path := range []string{"/api/v1/departments", "/api/v1/locations"} {
		response := requestTickets(t, server, path, "")
		response.Body.Close()
		if response.StatusCode != http.StatusUnauthorized {
			t.Fatalf("%s: expected 401, got %d", path, response.StatusCode)
		}
	}

	departmentsResponse := requestTickets(t, server, "/api/v1/departments", token)
	defer departmentsResponse.Body.Close()
	if departmentsResponse.StatusCode != http.StatusOK {
		t.Fatalf("departments: expected 200, got %d", departmentsResponse.StatusCode)
	}
	var departmentsBody struct {
		Departments []spec06LookupDepartment `json:"departments"`
	}
	decodeTicketResponse(t, departmentsResponse, &departmentsBody)
	if len(departmentsBody.Departments) != 2 || departmentsBody.Departments[0].ID != activeDepartmentFirst || departmentsBody.Departments[1].ID != activeDepartmentLast {
		t.Fatalf("unexpected active department lookup rows: %+v", departmentsBody.Departments)
	}
	if departmentsBody.Departments[0].Name != "SPEC06 Alpha" || departmentsBody.Departments[1].Name != "SPEC06 Omega" {
		t.Fatalf("departments are not ordered by name then id: %+v", departmentsBody.Departments)
	}

	locationsResponse := requestTickets(t, server, "/api/v1/locations", token)
	defer locationsResponse.Body.Close()
	if locationsResponse.StatusCode != http.StatusOK {
		t.Fatalf("locations: expected 200, got %d", locationsResponse.StatusCode)
	}
	var locationsBody struct {
		Locations []spec06LookupLocation `json:"locations"`
	}
	decodeTicketResponse(t, locationsResponse, &locationsBody)
	if len(locationsBody.Locations) != 2 || locationsBody.Locations[0].ID != activeLocationFirst || locationsBody.Locations[1].ID != nullableLocationID {
		t.Fatalf("unexpected active location lookup rows: %+v", locationsBody.Locations)
	}
	if locationsBody.Locations[1].Code != nil {
		t.Fatalf("nullable location code was not preserved: %+v", locationsBody.Locations[1])
	}
}

func TestSPEC062DevelopmentSeedReturnsApprovedActiveDepartments(t *testing.T) {
	pool, ctx := openTicketsTestPool(t)
	if err := admin.SeedDevelopmentFixtures(ctx, pool); err != nil {
		t.Fatalf("seed development fixtures: %v", err)
	}
	if err := admin.SeedDevelopmentFixtures(ctx, pool); err != nil {
		t.Fatalf("seed development fixtures a second time: %v", err)
	}

	var departmentID int64
	if err := pool.QueryRow(ctx, "SELECT id FROM departments WHERE code = 'IT'").Scan(&departmentID); err != nil {
		t.Fatalf("read seeded IT department: %v", err)
	}
	userID := insertUser(t, pool, ctx, "spec062-lookup-user", "SPEC062-LOOKUP", "SPEC-06.2 Lookup User", departmentID)
	token := insertSession(t, pool, ctx, userID)
	server := newTicketsApp(pool)

	response := requestTickets(t, server, "/api/v1/departments", token)
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		t.Fatalf("departments: expected 200, got %d", response.StatusCode)
	}
	var body struct {
		Departments []spec06LookupDepartment `json:"departments"`
	}
	decodeTicketResponse(t, response, &body)

	expected := []spec06LookupDepartment{
		{Code: "CON", Name: "Concierge"},
		{Code: "DA", Name: "Damaged Asset"},
		{Code: "FB", Name: "F&B"},
		{Code: "FIN", Name: "Finance Request"},
		{Code: "FO", Name: "Front Office"},
		{Code: "HK", Name: "Housekeeping"},
		{Code: "HKPPM", Name: "Housekeeping PPM"},
		{Code: "IT", Name: "IT"},
		{Code: "KIT", Name: "Kitchen"},
		{Code: "LDRY", Name: "Laundry"},
		{Code: "LF", Name: "Lost & Found"},
		{Code: "MAINT", Name: "Maintenance"},
		{Code: "REC", Name: "REC"},
		{Code: "SEC", Name: "Security"},
	}
	if len(body.Departments) != len(expected) {
		t.Fatalf("unexpected approved department count: got=%d want=%d rows=%+v", len(body.Departments), len(expected), body.Departments)
	}
	for index, department := range body.Departments {
		if department.Code != expected[index].Code || department.Name != expected[index].Name {
			t.Fatalf("unexpected department at position %d: got=%+v want code=%q name=%q", index, department, expected[index].Code, expected[index].Name)
		}
		if department.ID <= 0 {
			t.Fatalf("department %q did not contain a real database id", department.Code)
		}
	}
}

func TestSPEC063LocationLookupReturnsAll96VillasRows(t *testing.T) {
	pool, ctx := openTicketsTestPool(t)
	if err := admin.SeedDevelopmentFixtures(ctx, pool); err != nil {
		t.Fatalf("seed development fixtures: %v", err)
	}

	var departmentID int64
	if err := pool.QueryRow(ctx, "SELECT id FROM departments WHERE code = 'IT'").Scan(&departmentID); err != nil {
		t.Fatalf("read seeded IT department: %v", err)
	}
	userID := insertUser(t, pool, ctx, "spec063-lookup-user", "SPEC063-LOOKUP", "SPEC-06.3 Lookup User", departmentID)
	token := insertSession(t, pool, ctx, userID)
	server := newTicketsApp(pool)

	response := requestTickets(t, server, "/api/v1/locations", token)
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		t.Fatalf("locations: expected 200, got %d", response.StatusCode)
	}
	var body struct {
		Locations []spec06LookupLocation `json:"locations"`
	}
	decodeTicketResponse(t, response, &body)

	expected := expectedSPEC063LookupLocations()
	actual := make(map[string]spec06LookupLocation, len(expected))
	for _, location := range body.Locations {
		if location.Code == nil {
			continue
		}
		if _, ok := expected[*location.Code]; !ok {
			continue
		}
		if location.ID <= 0 {
			t.Fatalf("96 Villas location %q did not contain a real database id", *location.Code)
		}
		actual[*location.Code] = location
	}
	if len(actual) != len(expected) {
		t.Fatalf("unexpected 96 Villas lookup count: got=%d want=%d", len(actual), len(expected))
	}
	for code, name := range expected {
		location, ok := actual[code]
		if !ok || location.Name != name {
			t.Fatalf("unexpected 96 Villas lookup row for %q: got=%+v want name=%q", code, location, name)
		}
	}
}

func expectedSPEC063LookupLocations() map[string]string {
	expected := map[string]string{
		"96V":            "96 Villas",
		"96BWV-AREA-001": "96-BWV - Asian kitchen",
		"96BWV-AREA-002": "96-BWV - Auxiliary Swimming Pool",
		"96BWV-AREA-003": "96-BWV - Bathroom",
		"96BWV-AREA-004": "96-BWV - Buffet counter",
		"96BWV-AREA-005": "96-BWV - Cold kitchen",
		"96BWV-AREA-006": "96-BWV - Eng Fire Pump Room",
		"96BWV-AREA-007": "96-BWV - Eng Mainpool MEP Room",
		"96BWV-AREA-008": "96-BWV - Eng MEP Room",
		"96BWV-AREA-009": "96-BWV - Eng Subpool MEP Room",
		"96BWV-AREA-010": "96-BWV - Eng Water Treatment Room",
		"96BWV-AREA-011": "96-BWV - Eng Well Water treatment Room",
		"96BWV-AREA-012": "96-BWV - Eng workshop",
		"96BWV-AREA-013": "96-BWV - European kitchen",
		"96BWV-AREA-014": "96-BWV - Extra Pool",
		"96BWV-AREA-015": "96-BWV - Female Locker",
		"96BWV-AREA-016": "96-BWV - FO Reception",
		"96BWV-AREA-017": "96-BWV - Generator Room",
		"96BWV-AREA-018": "96-BWV - Gym",
		"96BWV-AREA-019": "96-BWV - HK Store",
		"96BWV-AREA-020": "96-BWV - Kid club",
		"96BWV-AREA-021": "96-BWV - Kid's Club",
		"96BWV-AREA-022": "96-BWV - Kid's Playground",
		"96BWV-AREA-023": "96-BWV - Lobby",
		"96BWV-AREA-024": "96-BWV - Lobby Lounge",
		"96BWV-AREA-025": "96-BWV - Main kitchen",
		"96BWV-AREA-026": "96-BWV - Main Pool",
		"96BWV-AREA-027": "96-BWV - Male Locker",
		"96BWV-AREA-028": "96-BWV - Outside",
		"96BWV-AREA-029": "96-BWV - PA Store",
		"96BWV-AREA-030": "96-BWV - Pastry kitchen",
		"96BWV-AREA-031": "96-BWV - Spa",
		"96BWV-AREA-032": "96-BWV - Toilet Gym",
		"96BWV-AREA-033": "96-BWV - Toilet Hồ bơi phụ",
		"96BWV-AREA-034": "96-BWV - Toilet Lobby",
		"96BWV-AREA-035": "96-BWV - Toilet Spa",
		"96BWV-AREA-036": "96-BWV - Toilet Tropicana",
		"96BWV-AREA-037": "96-BWV - Tropicana Bar",
		"96BWV-AREA-038": "96-BWV - Tropicana Kitchen",
		"96BWV-AREA-039": "96-BWV - Tropicana Restaurant",
		"96BWV-AREA-040": "96-BWV-Kitchen Office",
		"96BWV-AREA-041": "96-BWV-Steward",
	}
	for room := 1001; room <= 1099; room++ {
		code := fmt.Sprintf("96BWV-ROOM-%d", room)
		expected[code] = fmt.Sprintf("%d", room)
	}
	return expected
}

func TestSPEC06CreateTicketUsesSessionRequesterAndWritesActivityAtomically(t *testing.T) {
	pool, ctx := openTicketsTestPool(t)
	departmentID := insertDepartment(t, pool, ctx, "SPEC06-CREATE", "SPEC06 Create Department")
	locationID := insertLocation(t, pool, ctx, "SPEC06-ROOM", "SPEC06 Room")
	userID := insertUser(t, pool, ctx, "spec06-create-user", "SPEC06-CREATE-USER", "SPEC-06 Create User", departmentID)
	token := insertSession(t, pool, ctx, userID)
	server := newTicketsApp(pool)

	unauthenticated := postTickets(t, server, `{"department_id":1,"title":"Unauthenticated"}`, "")
	unauthenticated.Body.Close()
	if unauthenticated.StatusCode != http.StatusUnauthorized {
		t.Fatalf("unauthenticated create: expected 401, got %d", unauthenticated.StatusCode)
	}

	requestBody := fmt.Sprintf(`{"department_id":%d,"location_id":%d,"title":"  Lobby TV is not displaying content  ","description":"  Optional details  ","priority":true,"due_at":"2026-09-12T09:30:00+07:00"}`, departmentID, locationID)
	response := postTickets(t, server, requestBody, token)
	defer response.Body.Close()
	if response.StatusCode != http.StatusCreated {
		t.Fatalf("valid create: expected 201, got %d", response.StatusCode)
	}
	var body struct {
		Ticket ticketResponse `json:"ticket"`
	}
	decodeTicketResponse(t, response, &body)
	if body.Ticket.ID <= 0 || body.Ticket.Title != "Lobby TV is not displaying content" || body.Ticket.Status != "pending" || !body.Ticket.Priority || body.Ticket.DueAt == nil {
		t.Fatalf("unexpected created ticket summary: %+v", body.Ticket)
	}
	if body.Ticket.Description == nil || *body.Ticket.Description != "Optional details" {
		t.Fatalf("created description was not returned: %+v", body.Ticket.Description)
	}
	if body.Ticket.Requester.ID != userID || body.Ticket.Requester.FullName != "SPEC-06 Create User" || body.Ticket.Requester.DepartmentCode != "SPEC06-CREATE" {
		t.Fatalf("requester was not taken from the authenticated principal: %+v", body.Ticket.Requester)
	}
	if body.Ticket.Department.ID != departmentID || body.Ticket.Location == nil || body.Ticket.Location.ID != locationID {
		t.Fatalf("created ticket references are incorrect: %+v", body.Ticket)
	}
	if body.Ticket.AcceptedBy != nil || body.Ticket.AcceptedAt != nil || body.Ticket.ClosedAt != nil || len(body.Ticket.AssignedDepartments) != 0 || len(body.Ticket.AssignedUsers) != 0 {
		t.Fatalf("new ticket invariants were not preserved: %+v", body.Ticket)
	}

	var (
		description string
		status      string
		priority    bool
		dueAt       time.Time
		requesterID int64
		storedDept  int64
		storedLoc   *int64
	)
	if err := pool.QueryRow(ctx, `
SELECT description, status::text, priority, due_at, requester_id, department_id, location_id
FROM tickets
WHERE id = $1`, body.Ticket.ID).Scan(&description, &status, &priority, &dueAt, &requesterID, &storedDept, &storedLoc); err != nil {
		t.Fatalf("read created ticket: %v", err)
	}
	if description != "Optional details" || status != "pending" || !priority || requesterID != userID || storedDept != departmentID || storedLoc == nil || *storedLoc != locationID || !dueAt.Equal(time.Date(2026, 9, 12, 2, 30, 0, 0, time.UTC)) {
		t.Fatalf("created ticket did not persist requested fields: description=%q status=%q priority=%t due=%s requester=%d department=%d location=%v", description, status, priority, dueAt, requesterID, storedDept, storedLoc)
	}

	var activityActor int64
	var activityAction string
	if err := pool.QueryRow(ctx, `
SELECT actor_user_id, action
FROM ticket_activity
WHERE ticket_id = $1`, body.Ticket.ID).Scan(&activityActor, &activityAction); err != nil {
		t.Fatalf("read created ticket activity: %v", err)
	}
	if activityActor != userID || activityAction != "created" {
		t.Fatalf("created activity is incorrect: actor=%d action=%q", activityActor, activityAction)
	}

	encoded, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal created response: %v", err)
	}
	for _, forbidden := range []string{"password", "token", "session"} {
		if strings.Contains(strings.ToLower(string(encoded)), forbidden) {
			t.Fatalf("created response exposed forbidden field %q", forbidden)
		}
	}

	defaultRequest := fmt.Sprintf(`{"department_id":%d,"title":"Request without optional fields"}`, departmentID)
	defaultResponse := postTickets(t, server, defaultRequest, token)
	defer defaultResponse.Body.Close()
	if defaultResponse.StatusCode != http.StatusCreated {
		t.Fatalf("default create: expected 201, got %d", defaultResponse.StatusCode)
	}
	var defaultBody struct {
		Ticket ticketResponse `json:"ticket"`
	}
	decodeTicketResponse(t, defaultResponse, &defaultBody)
	if defaultBody.Ticket.Location != nil || defaultBody.Ticket.Priority {
		t.Fatalf("optional location/priority defaults were not preserved: %+v", defaultBody.Ticket)
	}
	var optionalCount int
	if err := pool.QueryRow(ctx, `
SELECT COUNT(*)
FROM ticket_activity
WHERE ticket_id = $1 AND actor_user_id = $2 AND action = 'created'`, defaultBody.Ticket.ID, userID).Scan(&optionalCount); err != nil {
		t.Fatalf("count default ticket activity: %v", err)
	}
	if optionalCount != 1 {
		t.Fatalf("default ticket activity count = %d, want 1", optionalCount)
	}
}

func TestSPEC06CreateTicketRollsBackWhenActivityInsertFails(t *testing.T) {
	pool, ctx := openTicketsTestPool(t)
	departmentID := insertDepartment(t, pool, ctx, "SPEC06-ROLLBACK", "SPEC06 Rollback Department")
	userID := insertUser(t, pool, ctx, "spec06-rollback-user", "SPEC06-ROLLBACK-USER", "SPEC-06 Rollback User", departmentID)
	token := insertSession(t, pool, ctx, userID)
	server := newTicketsApp(pool)

	if _, err := pool.Exec(ctx, `
CREATE FUNCTION spec06_fail_ticket_activity() RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
    RAISE EXCEPTION 'test-only ticket activity failure';
END;
$$`); err != nil {
		t.Fatalf("create activity failure function: %v", err)
	}
	if _, err := pool.Exec(ctx, `
CREATE TRIGGER spec06_fail_ticket_activity_trigger
BEFORE INSERT ON ticket_activity
FOR EACH ROW EXECUTE FUNCTION spec06_fail_ticket_activity()`); err != nil {
		t.Fatalf("create activity failure trigger: %v", err)
	}
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), `DROP TRIGGER IF EXISTS spec06_fail_ticket_activity_trigger ON ticket_activity`)
		_, _ = pool.Exec(context.Background(), `DROP FUNCTION IF EXISTS spec06_fail_ticket_activity()`)
	})

	response := postTickets(t, server, fmt.Sprintf(`{"department_id":%d,"title":"Rollback request"}`, departmentID), token)
	defer response.Body.Close()
	if response.StatusCode != http.StatusInternalServerError {
		t.Fatalf("activity failure: expected 500, got %d", response.StatusCode)
	}
	var createdCount int
	if err := pool.QueryRow(ctx, `SELECT COUNT(*) FROM tickets WHERE title = 'Rollback request'`).Scan(&createdCount); err != nil {
		t.Fatalf("check rolled back ticket: %v", err)
	}
	if createdCount != 0 {
		t.Fatalf("ticket remained after activity failure: count=%d", createdCount)
	}
}

func TestSPEC06CreateTicketRejectsInvalidAndUnavailableInput(t *testing.T) {
	pool, ctx := openTicketsTestPool(t)
	departmentID := insertDepartment(t, pool, ctx, "SPEC06-VALIDATE", "SPEC06 Validate Department")
	locationID := insertLocation(t, pool, ctx, "SPEC06-VALIDATE-LOC", "SPEC06 Validate Location")
	userID := insertUser(t, pool, ctx, "spec06-validate-user", "SPEC06-VALIDATE-USER", "SPEC-06 Validate User", departmentID)
	token := insertSession(t, pool, ctx, userID)
	server := newTicketsApp(pool)

	cases := []struct {
		name string
		body string
	}{
		{name: "missing department", body: `{"title":"Request"}`},
		{name: "zero department", body: `{"department_id":0,"title":"Request"}`},
		{name: "zero location", body: fmt.Sprintf(`{"department_id":%d,"location_id":0,"title":"Request"}`, departmentID)},
		{name: "blank title", body: fmt.Sprintf(`{"department_id":%d,"title":"   "}`, departmentID)},
		{name: "oversized title", body: fmt.Sprintf(`{"department_id":%d,"title":"%s"}`, departmentID, strings.Repeat("x", 256))},
		{name: "oversized description", body: fmt.Sprintf(`{"department_id":%d,"title":"Request","description":"%s"}`, departmentID, strings.Repeat("x", 5001))},
		{name: "unknown field", body: fmt.Sprintf(`{"department_id":%d,"title":"Request","requester_id":999}`, departmentID)},
		{name: "invalid due time", body: fmt.Sprintf(`{"department_id":%d,"title":"Request","due_at":"2026-09-12T09:30:00"}`, departmentID)},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			response := postTickets(t, server, testCase.body, token)
			defer response.Body.Close()
			if response.StatusCode != http.StatusBadRequest {
				t.Fatalf("expected 400, got %d", response.StatusCode)
			}
			var errorBody struct {
				Error struct {
					Code string `json:"code"`
				} `json:"error"`
			}
			decodeTicketResponse(t, response, &errorBody)
			if errorBody.Error.Code != "invalid_request" {
				t.Fatalf("unexpected safe error: %+v", errorBody)
			}
		})
	}

	if _, err := pool.Exec(ctx, `UPDATE departments SET is_active = FALSE WHERE id = $1`, departmentID); err != nil {
		t.Fatalf("deactivate create department: %v", err)
	}
	departmentUnavailable := postTickets(t, server, fmt.Sprintf(`{"department_id":%d,"title":"Request"}`, departmentID), token)
	defer departmentUnavailable.Body.Close()
	if departmentUnavailable.StatusCode != http.StatusBadRequest {
		t.Fatalf("unavailable department: expected 400, got %d", departmentUnavailable.StatusCode)
	}
	var departmentError struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	decodeTicketResponse(t, departmentUnavailable, &departmentError)
	if departmentError.Error.Code != "department_unavailable" {
		t.Fatalf("unexpected department error: %+v", departmentError)
	}

	if _, err := pool.Exec(ctx, `UPDATE departments SET is_active = TRUE WHERE id = $1`, departmentID); err != nil {
		t.Fatalf("reactivate create department: %v", err)
	}
	if _, err := pool.Exec(ctx, `UPDATE locations SET is_active = FALSE WHERE id = $1`, locationID); err != nil {
		t.Fatalf("deactivate create location: %v", err)
	}
	locationUnavailable := postTickets(t, server, fmt.Sprintf(`{"department_id":%d,"location_id":%d,"title":"Request"}`, departmentID, locationID), token)
	defer locationUnavailable.Body.Close()
	if locationUnavailable.StatusCode != http.StatusBadRequest {
		t.Fatalf("unavailable location: expected 400, got %d", locationUnavailable.StatusCode)
	}
	var locationError struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	decodeTicketResponse(t, locationUnavailable, &locationError)
	if locationError.Error.Code != "location_unavailable" {
		t.Fatalf("unexpected location error: %+v", locationError)
	}
}

func postTickets(t *testing.T, server *fiber.App, body, token string) *http.Response {
	t.Helper()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/tickets", bytes.NewBufferString(body))
	request.Header.Set("Content-Type", "application/json")
	if token != "" {
		request.AddCookie(&http.Cookie{Name: "bwp_session", Value: token})
	}
	response, err := server.Test(request)
	if err != nil {
		t.Fatalf("POST /api/v1/tickets: %v", err)
	}
	return response
}
