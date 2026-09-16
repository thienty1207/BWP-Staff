package tickets_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ticketAcceptanceFixture struct {
	RequesterToken       string
	FirstAccepterToken   string
	SecondAccepterToken  string
	RequesterID          int64
	FirstAccepterID      int64
	SecondAccepterID     int64
	RequestDepartmentID  int64
	LocationID           int64
	AssignedDepartmentID int64
	AssignedUserID       int64
	PendingTicketID      int64
	ClosedTicketID       int64
}

type acceptHTTPResult struct {
	response *http.Response
	err      error
}

func TestAcceptTicketEndpointEnforcesLifecycleContract(t *testing.T) {
	pool, ctx := openTicketsTestPool(t)
	fixture := seedTicketAcceptanceFixture(t, pool, ctx)
	server := newTicketsApp(pool)

	t.Run("authentication precedes ID validation", func(t *testing.T) {
		for _, rawID := range []string{"not-an-id", "1.5", "0", "-1", "9223372036854775808"} {
			response := requestAcceptTicket(t, server, rawID, "", "")
			defer response.Body.Close()
			if response.StatusCode != http.StatusUnauthorized {
				t.Fatalf("%s: expected status 401, got %d", rawID, response.StatusCode)
			}
		}
	})

	t.Run("rejects malformed positive-ID requests safely", func(t *testing.T) {
		for _, rawID := range []string{"abc", "1.5", "0", "-1", "9223372036854775808"} {
			response := requestAcceptTicket(t, server, rawID, fixture.FirstAccepterToken, "")
			var body map[string]map[string]string
			decodeTicketResponse(t, response, &body)
			response.Body.Close()
			if response.StatusCode != http.StatusBadRequest {
				t.Fatalf("%s: expected status 400, got %d", rawID, response.StatusCode)
			}
			if body["error"]["code"] != "invalid_request" || body["error"]["message"] != "Invalid request" {
				t.Fatalf("%s: unexpected error: %+v", rawID, body)
			}
		}
	})

	t.Run("maps a missing positive ID to ticket_not_found", func(t *testing.T) {
		response := requestAcceptTicket(t, server, "9223372036854775807", fixture.FirstAccepterToken, "")
		var body map[string]map[string]string
		decodeTicketResponse(t, response, &body)
		response.Body.Close()
		if response.StatusCode != http.StatusNotFound {
			t.Fatalf("expected status 404, got %d", response.StatusCode)
		}
		if body["error"]["code"] != "ticket_not_found" || body["error"]["message"] != "Ticket not found" {
			t.Fatalf("unexpected error: %+v", body)
		}
	})

	t.Run("ignores actor fields in an untrusted request body", func(t *testing.T) {
		body := fmt.Sprintf(`{"accepted_by":%d,"accepted_at":"2035-01-01T00:00:00Z","status":"accepted","actor_user_id":%d}`, fixture.SecondAccepterID, fixture.SecondAccepterID)
		response := requestAcceptTicket(t, server, formatInt64(fixture.PendingTicketID), fixture.FirstAccepterToken, body)
		var result ticketDetailResponse
		decodeTicketResponse(t, response, &result)
		response.Body.Close()
		if response.StatusCode != http.StatusOK {
			t.Fatalf("expected status 200, got %d", response.StatusCode)
		}
		if result.Ticket.Status != "accepted" || result.Ticket.AcceptedBy == nil || result.Ticket.AcceptedBy.ID != fixture.FirstAccepterID {
			t.Fatalf("request body spoofed acceptance actor: %+v", result.Ticket)
		}
		if result.Ticket.AcceptedAt == nil || result.Ticket.AcceptedAt.Format(time.RFC3339) == "2035-01-01T00:00:00Z" {
			t.Fatalf("accepted_at was taken from the request body: %+v", result.Ticket.AcceptedAt)
		}
	})

	t.Run("first acceptance persists canonical state and one activity atomically", func(t *testing.T) {
		if fixture.PendingTicketID == 0 {
			t.Fatal("pending ticket was not seeded")
		}
		var body ticketDetailResponse
		response := requestAcceptTicket(t, server, formatInt64(fixture.PendingTicketID), fixture.FirstAccepterToken, "")
		decodeTicketResponse(t, response, &body)
		response.Body.Close()
		if response.StatusCode != http.StatusOK {
			t.Fatalf("expected status 200, got %d", response.StatusCode)
		}
		if body.Ticket.Status != "accepted" || body.Ticket.AcceptedBy == nil || body.Ticket.AcceptedBy.ID != fixture.FirstAccepterID || body.Ticket.AcceptedAt == nil {
			t.Fatalf("unexpected accepted ticket: %+v", body.Ticket)
		}
		if body.Ticket.Requester.ID != fixture.RequesterID || body.Ticket.Department.ID != fixture.RequestDepartmentID || body.Ticket.Location == nil || body.Ticket.Location.ID != fixture.LocationID {
			t.Fatalf("original ticket references changed: %+v", body.Ticket)
		}
		if len(body.Ticket.AssignedDepartments) != 1 || body.Ticket.AssignedDepartments[0].ID != fixture.AssignedDepartmentID || len(body.Ticket.AssignedUsers) != 1 || body.Ticket.AssignedUsers[0].ID != fixture.AssignedUserID {
			t.Fatalf("assignments changed during acceptance: %+v", body.Ticket)
		}
		if !body.Ticket.UpdatedAt.Equal(*body.Ticket.AcceptedAt) {
			t.Fatalf("accepted_at and updated_at are not the same event timestamp: accepted=%s updated=%s", body.Ticket.AcceptedAt, body.Ticket.UpdatedAt)
		}

		var activityActor int64
		var activityAt time.Time
		var activityCount int
		if err := pool.QueryRow(ctx, `
SELECT actor_user_id, created_at
FROM ticket_activity
WHERE ticket_id = $1 AND action = 'accepted'`, fixture.PendingTicketID).Scan(&activityActor, &activityAt); err != nil {
			t.Fatalf("read accepted activity: %v", err)
		}
		if err := pool.QueryRow(ctx, `
SELECT COUNT(*)
FROM ticket_activity
WHERE ticket_id = $1 AND action = 'accepted'`, fixture.PendingTicketID).Scan(&activityCount); err != nil {
			t.Fatalf("count accepted activities: %v", err)
		}
		if activityActor != fixture.FirstAccepterID || activityCount != 1 || !activityAt.Equal(*body.Ticket.AcceptedAt) {
			t.Fatalf("accepted activity is not the single persisted transition: actor=%d count=%d activity_at=%s ticket_at=%s", activityActor, activityCount, activityAt, body.Ticket.AcceptedAt)
		}
	})

	t.Run("same-user replay is idempotent", func(t *testing.T) {
		var beforeAcceptedAt, beforeUpdatedAt time.Time
		if err := pool.QueryRow(ctx, `SELECT accepted_at, updated_at FROM tickets WHERE id = $1`, fixture.PendingTicketID).Scan(&beforeAcceptedAt, &beforeUpdatedAt); err != nil {
			t.Fatalf("read accepted timestamps before replay: %v", err)
		}
		response := requestAcceptTicket(t, server, formatInt64(fixture.PendingTicketID), fixture.FirstAccepterToken, "")
		var body ticketDetailResponse
		decodeTicketResponse(t, response, &body)
		response.Body.Close()
		if response.StatusCode != http.StatusOK || body.Ticket.AcceptedAt == nil || !body.Ticket.AcceptedAt.Equal(beforeAcceptedAt) || !body.Ticket.UpdatedAt.Equal(beforeUpdatedAt) {
			t.Fatalf("same-user replay changed accepted state: status=%d ticket=%+v", response.StatusCode, body.Ticket)
		}
		var activityCount int
		if err := pool.QueryRow(ctx, `SELECT COUNT(*) FROM ticket_activity WHERE ticket_id = $1 AND action = 'accepted'`, fixture.PendingTicketID).Scan(&activityCount); err != nil {
			t.Fatalf("count accepted activities after replay: %v", err)
		}
		if activityCount != 1 {
			t.Fatalf("same-user replay duplicated accepted activity: %d", activityCount)
		}
	})

	t.Run("different-user replay preserves the first owner", func(t *testing.T) {
		response := requestAcceptTicket(t, server, formatInt64(fixture.PendingTicketID), fixture.SecondAccepterToken, "")
		var body map[string]map[string]string
		decodeTicketResponse(t, response, &body)
		response.Body.Close()
		if response.StatusCode != http.StatusConflict || body["error"]["code"] != "ticket_already_accepted" || body["error"]["message"] != "Ticket has already been accepted" {
			t.Fatalf("unexpected different-user conflict: status=%d body=%+v", response.StatusCode, body)
		}
		var ownerID int64
		if err := pool.QueryRow(ctx, `SELECT accepted_by FROM tickets WHERE id = $1`, fixture.PendingTicketID).Scan(&ownerID); err != nil {
			t.Fatalf("read accepted owner after conflict: %v", err)
		}
		if ownerID != fixture.FirstAccepterID {
			t.Fatalf("different user overwrote first owner: got %d", ownerID)
		}
	})

	t.Run("closed ticket returns a safe conflict without mutation", func(t *testing.T) {
		var beforeAcceptedAt, beforeUpdatedAt time.Time
		if err := pool.QueryRow(ctx, `SELECT accepted_at, updated_at FROM tickets WHERE id = $1`, fixture.ClosedTicketID).Scan(&beforeAcceptedAt, &beforeUpdatedAt); err != nil {
			t.Fatalf("read closed timestamps before accept: %v", err)
		}
		response := requestAcceptTicket(t, server, formatInt64(fixture.ClosedTicketID), fixture.FirstAccepterToken, "")
		var body map[string]map[string]string
		decodeTicketResponse(t, response, &body)
		response.Body.Close()
		if response.StatusCode != http.StatusConflict || body["error"]["code"] != "ticket_closed" || body["error"]["message"] != "Ticket is closed" {
			t.Fatalf("unexpected closed conflict: status=%d body=%+v", response.StatusCode, body)
		}
		var status string
		var acceptedAt, updatedAt time.Time
		if err := pool.QueryRow(ctx, `SELECT status::text, accepted_at, updated_at FROM tickets WHERE id = $1`, fixture.ClosedTicketID).Scan(&status, &acceptedAt, &updatedAt); err != nil {
			t.Fatalf("read closed state after conflict: %v", err)
		}
		if status != "closed" || !acceptedAt.Equal(beforeAcceptedAt) || !updatedAt.Equal(beforeUpdatedAt) {
			t.Fatalf("closed ticket was mutated: status=%s accepted_at=%s updated_at=%s", status, acceptedAt, updatedAt)
		}
	})
}

func TestAcceptTicketConcurrencyAllowsOnlyOneFirstAccepter(t *testing.T) {
	pool, ctx := openTicketsTestPool(t)
	fixture := seedTicketAcceptanceFixture(t, pool, ctx)
	server := newTicketsApp(pool)

	results := make(chan acceptHTTPResult, 2)
	var waitGroup sync.WaitGroup
	for _, token := range []string{fixture.FirstAccepterToken, fixture.SecondAccepterToken} {
		waitGroup.Add(1)
		go func(token string) {
			defer waitGroup.Done()
			response, err := doAcceptTicket(server, fixture.PendingTicketID, token, "")
			results <- acceptHTTPResult{response: response, err: err}
		}(token)
	}
	waitGroup.Wait()
	close(results)

	statusCounts := map[int]int{}
	for result := range results {
		if result.err != nil {
			t.Fatalf("concurrent accept request failed: %v", result.err)
		}
		statusCounts[result.response.StatusCode]++
		result.response.Body.Close()
	}
	if statusCounts[http.StatusOK] != 1 || statusCounts[http.StatusConflict] != 1 {
		t.Fatalf("expected one success and one conflict, got statuses %+v", statusCounts)
	}

	var ownerID int64
	var acceptedAt time.Time
	if err := pool.QueryRow(ctx, `SELECT accepted_by, accepted_at FROM tickets WHERE id = $1`, fixture.PendingTicketID).Scan(&ownerID, &acceptedAt); err != nil {
		t.Fatalf("read concurrent accepted ticket: %v", err)
	}
	if ownerID != fixture.FirstAccepterID && ownerID != fixture.SecondAccepterID {
		t.Fatalf("unexpected concurrent owner %d", ownerID)
	}
	var activityCount int
	if err := pool.QueryRow(ctx, `SELECT COUNT(*) FROM ticket_activity WHERE ticket_id = $1 AND action = 'accepted'`, fixture.PendingTicketID).Scan(&activityCount); err != nil {
		t.Fatalf("count concurrent accepted activities: %v", err)
	}
	if activityCount != 1 || acceptedAt.IsZero() {
		t.Fatalf("concurrent acceptance produced invalid history: count=%d accepted_at=%s", activityCount, acceptedAt)
	}
}

func TestAcceptTicketRollsBackTicketAndActivityTogether(t *testing.T) {
	pool, ctx := openTicketsTestPool(t)
	fixture := seedTicketAcceptanceFixture(t, pool, ctx)
	server := newTicketsApp(pool)

	if _, err := pool.Exec(ctx, `
CREATE FUNCTION spec08_fail_ticket_activity() RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
    RAISE EXCEPTION 'test-only ticket activity failure';
END;
$$`); err != nil {
		t.Fatalf("create activity failure function: %v", err)
	}
	if _, err := pool.Exec(ctx, `
CREATE TRIGGER spec08_fail_ticket_activity_trigger
BEFORE INSERT ON ticket_activity
FOR EACH ROW EXECUTE FUNCTION spec08_fail_ticket_activity()`); err != nil {
		t.Fatalf("create activity failure trigger: %v", err)
	}
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), `DROP TRIGGER IF EXISTS spec08_fail_ticket_activity_trigger ON ticket_activity`)
		_, _ = pool.Exec(context.Background(), `DROP FUNCTION IF EXISTS spec08_fail_ticket_activity()`)
	})

	response := requestAcceptTicket(t, server, formatInt64(fixture.PendingTicketID), fixture.FirstAccepterToken, "")
	var body map[string]map[string]string
	decodeTicketResponse(t, response, &body)
	response.Body.Close()
	if response.StatusCode != http.StatusInternalServerError || body["error"]["code"] != "internal_server_error" || body["error"]["message"] != "Internal server error" {
		t.Fatalf("unexpected atomicity failure response: status=%d body=%+v", response.StatusCode, body)
	}

	var status string
	var acceptedBy *int64
	var acceptedAt *time.Time
	if err := pool.QueryRow(ctx, `SELECT status::text, accepted_by, accepted_at FROM tickets WHERE id = $1`, fixture.PendingTicketID).Scan(&status, &acceptedBy, &acceptedAt); err != nil {
		t.Fatalf("read ticket after rollback: %v", err)
	}
	if status != "pending" || acceptedBy != nil || acceptedAt != nil {
		t.Fatalf("ticket update was not rolled back: status=%s accepted_by=%v accepted_at=%v", status, acceptedBy, acceptedAt)
	}
	var activityCount int
	if err := pool.QueryRow(ctx, `SELECT COUNT(*) FROM ticket_activity WHERE ticket_id = $1 AND action = 'accepted'`, fixture.PendingTicketID).Scan(&activityCount); err != nil {
		t.Fatalf("count activities after rollback: %v", err)
	}
	if activityCount != 0 {
		t.Fatalf("accepted activity survived ticket rollback: %d", activityCount)
	}
}

func seedTicketAcceptanceFixture(t *testing.T, pool *pgxpool.Pool, ctx context.Context) ticketAcceptanceFixture {
	t.Helper()
	var fixture ticketAcceptanceFixture
	requesterDepartmentID := insertDepartment(t, pool, ctx, "SPEC08-REQ", "SPEC-08 Requester")
	fixture.RequestDepartmentID = insertDepartment(t, pool, ctx, "SPEC08-DEST", "SPEC-08 Destination")
	fixture.AssignedDepartmentID = insertDepartment(t, pool, ctx, "SPEC08-ASSIGN", "SPEC-08 Assigned")
	fixture.LocationID = insertLocation(t, pool, ctx, "SPEC08-LOC", "SPEC-08 Location")
	fixture.RequesterID = insertUser(t, pool, ctx, "spec08-requester", "SPEC08-REQUESTER", "SPEC-08 Requester", requesterDepartmentID)
	fixture.FirstAccepterID = insertUser(t, pool, ctx, "spec08-first", "SPEC08-FIRST", "SPEC-08 First Accepter", fixture.AssignedDepartmentID)
	fixture.SecondAccepterID = insertUser(t, pool, ctx, "spec08-second", "SPEC08-SECOND", "SPEC-08 Second Accepter", fixture.AssignedDepartmentID)
	fixture.AssignedUserID = insertUser(t, pool, ctx, "spec08-assignee", "SPEC08-ASSIGNEE", "SPEC-08 Assignee", fixture.AssignedDepartmentID)
	fixture.RequesterToken = insertSession(t, pool, ctx, fixture.RequesterID)
	fixture.FirstAccepterToken = insertSession(t, pool, ctx, fixture.FirstAccepterID)
	fixture.SecondAccepterToken = insertSession(t, pool, ctx, fixture.SecondAccepterID)

	createdAt := time.Date(2026, 9, 16, 8, 0, 0, 0, time.UTC)
	if err := pool.QueryRow(ctx, `
INSERT INTO tickets (requester_id, department_id, location_id, title, description, status, priority, created_at, updated_at)
VALUES ($1, $2, $3, 'SPEC-08 pending request', 'Acceptance test description', 'pending'::ticket_status, TRUE, $4, $4)
RETURNING id`, fixture.RequesterID, fixture.RequestDepartmentID, fixture.LocationID, createdAt).Scan(&fixture.PendingTicketID); err != nil {
		t.Fatalf("insert pending acceptance ticket: %v", err)
	}
	if _, err := pool.Exec(ctx, `INSERT INTO ticket_assigned_departments (ticket_id, department_id) VALUES ($1, $2)`, fixture.PendingTicketID, fixture.AssignedDepartmentID); err != nil {
		t.Fatalf("insert acceptance department assignment: %v", err)
	}
	if _, err := pool.Exec(ctx, `INSERT INTO ticket_assigned_users (ticket_id, user_id) VALUES ($1, $2)`, fixture.PendingTicketID, fixture.AssignedUserID); err != nil {
		t.Fatalf("insert acceptance user assignment: %v", err)
	}

	closedAcceptedAt := createdAt.Add(-time.Hour)
	closedAt := createdAt.Add(-30 * time.Minute)
	if err := pool.QueryRow(ctx, `
INSERT INTO tickets (requester_id, department_id, title, status, accepted_by, accepted_at, closed_by, closed_at, created_at, updated_at)
VALUES ($1, $2, 'SPEC-08 closed request', 'closed'::ticket_status, $3, $4, $1, $5, $6, $5)
RETURNING id`, fixture.RequesterID, fixture.RequestDepartmentID, fixture.FirstAccepterID, closedAcceptedAt, closedAt, createdAt.Add(-2*time.Hour)).Scan(&fixture.ClosedTicketID); err != nil {
		t.Fatalf("insert closed acceptance ticket: %v", err)
	}
	return fixture
}

func requestAcceptTicket(t *testing.T, server *fiber.App, rawID, token, body string) *http.Response {
	t.Helper()
	response, err := doAcceptTicket(server, rawID, token, body)
	if err != nil {
		t.Fatalf("POST /api/v1/tickets/%s/accept: %v", rawID, err)
	}
	return response
}

func doAcceptTicket(server *fiber.App, ticketID any, token, body string) (*http.Response, error) {
	var rawID string
	switch value := ticketID.(type) {
	case int64:
		rawID = formatInt64(value)
	case string:
		rawID = value
	default:
		rawID = fmt.Sprint(value)
	}
	request := httptest.NewRequest(http.MethodPost, "/api/v1/tickets/"+rawID+"/accept", strings.NewReader(body))
	if token != "" {
		request.AddCookie(&http.Cookie{Name: "bwp_session", Value: token})
	}
	if body != "" {
		request.Header.Set("Content-Type", "application/json")
	}
	return server.Test(request)
}
