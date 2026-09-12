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
	if body.Ticket.Requester.ID != userID || body.Ticket.Requester.FullName != "SPEC-06 Create User" {
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
	for _, forbidden := range []string{"password", "token", "session", "description"} {
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
