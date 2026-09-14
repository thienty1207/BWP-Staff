package tickets_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type ticketDetailResponse struct {
	Ticket ticketResponse `json:"ticket"`
}

type ticketDetailFixture struct {
	RequesterToken       string
	CrossDepartmentToken string
	RequesterID          int64
	RequestDepartmentID  int64
	LocationID           int64
	OwnerID              int64
	AssignedDepartmentID int64
	SecondAssignedDeptID int64
	AssignedUserID       int64
	SecondAssignedUserID int64
	PendingTicketID      int64
	AcceptedTicketID     int64
	ClosedTicketID       int64
	HistoricalTicketID   int64
}

func TestTicketDetailReadUsesAuthenticatedRealRelationalData(t *testing.T) {
	pool, ctx := openTicketsTestPool(t)
	fixture := seedTicketDetailFixture(t, pool, ctx)
	server := newTicketsApp(pool)

	t.Run("authentication runs before path validation", func(t *testing.T) {
		response := requestTickets(t, server, "/api/v1/tickets/not-an-id", "")
		defer response.Body.Close()
		if response.StatusCode != http.StatusUnauthorized {
			t.Fatalf("expected unauthenticated status 401, got %d", response.StatusCode)
		}
	})

	t.Run("requires authentication for a valid positive detail ID", func(t *testing.T) {
		response := requestTickets(t, server, "/api/v1/tickets/1", "")
		defer response.Body.Close()
		if response.StatusCode != http.StatusUnauthorized {
			t.Fatalf("expected unauthenticated status 401, got %d", response.StatusCode)
		}
	})

	t.Run("rejects malformed non-positive and overflowing IDs", func(t *testing.T) {
		for _, rawID := range []string{"abc", "1.5", "0", "-1", "9223372036854775808"} {
			response := requestTickets(t, server, "/api/v1/tickets/"+rawID, fixture.RequesterToken)
			var body map[string]map[string]string
			decodeTicketResponse(t, response, &body)
			response.Body.Close()
			if response.StatusCode != http.StatusBadRequest {
				t.Fatalf("%s: expected status 400, got %d", rawID, response.StatusCode)
			}
			if body["error"]["code"] != "invalid_request" || body["error"]["message"] != "Invalid request" {
				t.Fatalf("%s: unexpected safe error: %+v", rawID, body)
			}
		}
	})

	t.Run("maps a missing positive ID to ticket_not_found", func(t *testing.T) {
		response := requestTickets(t, server, "/api/v1/tickets/9223372036854775807", fixture.RequesterToken)
		var body map[string]map[string]string
		decodeTicketResponse(t, response, &body)
		response.Body.Close()
		if response.StatusCode != http.StatusNotFound {
			t.Fatalf("expected status 404, got %d", response.StatusCode)
		}
		if body["error"]["code"] != "ticket_not_found" || body["error"]["message"] != "Ticket not found" {
			t.Fatalf("unexpected not-found error: %+v", body)
		}
	})

	t.Run("returns pending core fields and explicit empty nullable assignments", func(t *testing.T) {
		response := requestTickets(t, server, "/api/v1/tickets/"+formatInt64(fixture.PendingTicketID), fixture.RequesterToken)
		defer response.Body.Close()
		if response.StatusCode != http.StatusOK {
			t.Fatalf("expected pending detail status 200, got %d", response.StatusCode)
		}

		var body ticketDetailResponse
		decodeTicketResponse(t, response, &body)
		if body.Ticket.ID != fixture.PendingTicketID || body.Ticket.Title != "SPEC-07 pending request" || body.Ticket.Status != "pending" {
			t.Fatalf("unexpected pending ticket: %+v", body.Ticket)
		}
		if body.Ticket.Description != nil || body.Ticket.Location != nil || body.Ticket.AcceptedBy != nil || body.Ticket.AcceptedAt != nil || body.Ticket.ClosedAt != nil {
			t.Fatalf("pending nullable fields were not preserved: %+v", body.Ticket)
		}
		if body.Ticket.AssignedDepartments == nil || len(body.Ticket.AssignedDepartments) != 0 || body.Ticket.AssignedUsers == nil || len(body.Ticket.AssignedUsers) != 0 {
			t.Fatalf("pending assignments must be explicit empty arrays: %+v", body.Ticket)
		}
		if body.Ticket.Requester.ID != fixture.RequesterID || body.Ticket.Department.ID != fixture.RequestDepartmentID {
			t.Fatalf("pending requester/department semantics were not preserved: %+v", body.Ticket)
		}
	})

	t.Run("returns accepted owner and assignments in stable insertion order", func(t *testing.T) {
		response := requestTickets(t, server, "/api/v1/tickets/"+formatInt64(fixture.AcceptedTicketID), fixture.RequesterToken)
		defer response.Body.Close()
		if response.StatusCode != http.StatusOK {
			t.Fatalf("expected accepted detail status 200, got %d", response.StatusCode)
		}

		var body ticketDetailResponse
		decodeTicketResponse(t, response, &body)
		if body.Ticket.ID != fixture.AcceptedTicketID || body.Ticket.Status != "accepted" || body.Ticket.Description == nil || *body.Ticket.Description != "The complete SPEC-07 description." {
			t.Fatalf("unexpected accepted ticket: %+v", body.Ticket)
		}
		if body.Ticket.Department.ID != fixture.RequestDepartmentID || body.Ticket.Location == nil || body.Ticket.Location.ID != fixture.LocationID {
			t.Fatalf("original request references were not preserved: %+v", body.Ticket)
		}
		if body.Ticket.AcceptedBy == nil || body.Ticket.AcceptedBy.ID != fixture.OwnerID || body.Ticket.AcceptedAt == nil {
			t.Fatalf("owner semantics were not taken from accepted_by: %+v", body.Ticket)
		}
		if len(body.Ticket.AssignedDepartments) != 2 || body.Ticket.AssignedDepartments[0].ID != fixture.AssignedDepartmentID || body.Ticket.AssignedDepartments[1].ID != fixture.SecondAssignedDeptID {
			t.Fatalf("department assignment order was not stable: %+v", body.Ticket.AssignedDepartments)
		}
		if len(body.Ticket.AssignedUsers) != 2 || body.Ticket.AssignedUsers[0].ID != fixture.AssignedUserID || body.Ticket.AssignedUsers[1].ID != fixture.SecondAssignedUserID {
			t.Fatalf("user assignment order was not stable: %+v", body.Ticket.AssignedUsers)
		}
		if body.Ticket.AssignedUsers[0].ID == body.Ticket.AcceptedBy.ID {
			t.Fatal("assigned user incorrectly redefined owner")
		}
	})

	t.Run("keeps accepted owner and closure timestamp on closed tickets", func(t *testing.T) {
		response := requestTickets(t, server, "/api/v1/tickets/"+formatInt64(fixture.ClosedTicketID), fixture.RequesterToken)
		defer response.Body.Close()
		if response.StatusCode != http.StatusOK {
			t.Fatalf("expected closed detail status 200, got %d", response.StatusCode)
		}

		var body ticketDetailResponse
		decodeTicketResponse(t, response, &body)
		if body.Ticket.Status != "closed" || body.Ticket.AcceptedBy == nil || body.Ticket.AcceptedBy.ID != fixture.OwnerID || body.Ticket.ClosedAt == nil {
			t.Fatalf("closed lifecycle fields were not preserved: %+v", body.Ticket)
		}
	})

	t.Run("reads historical inactive references", func(t *testing.T) {
		response := requestTickets(t, server, "/api/v1/tickets/"+formatInt64(fixture.HistoricalTicketID), fixture.RequesterToken)
		defer response.Body.Close()
		if response.StatusCode != http.StatusOK {
			t.Fatalf("expected historical detail status 200, got %d", response.StatusCode)
		}

		var body ticketDetailResponse
		decodeTicketResponse(t, response, &body)
		if body.Ticket.Department.ID != fixture.RequestDepartmentID || body.Ticket.Location == nil || body.Ticket.Location.ID != fixture.LocationID {
			t.Fatalf("inactive historical references caused data loss: %+v", body.Ticket)
		}
		if body.Ticket.AcceptedBy == nil || body.Ticket.AcceptedBy.ID != fixture.OwnerID || len(body.Ticket.AssignedDepartments) != 1 || len(body.Ticket.AssignedUsers) != 1 {
			t.Fatalf("inactive historical assignment references were not readable: %+v", body.Ticket)
		}
	})

	t.Run("allows another active staff user under the shared list access model", func(t *testing.T) {
		response := requestTickets(t, server, "/api/v1/tickets/"+formatInt64(fixture.AcceptedTicketID), fixture.CrossDepartmentToken)
		defer response.Body.Close()
		if response.StatusCode != http.StatusOK {
			t.Fatalf("expected cross-department detail status 200, got %d", response.StatusCode)
		}
	})

	t.Run("uses a compact ticket envelope without secret or heavy fields", func(t *testing.T) {
		response := requestTickets(t, server, "/api/v1/tickets/"+formatInt64(fixture.AcceptedTicketID), fixture.RequesterToken)
		payload, err := io.ReadAll(response.Body)
		response.Body.Close()
		if err != nil {
			t.Fatalf("read detail response: %v", err)
		}
		if response.StatusCode != http.StatusOK {
			t.Fatalf("expected detail status 200, got %d", response.StatusCode)
		}
		var body ticketDetailResponse
		if err := json.Unmarshal(payload, &body); err != nil {
			t.Fatalf("decode detail envelope: %v", err)
		}
		if body.Ticket.ID != fixture.AcceptedTicketID {
			t.Fatalf("unexpected detail envelope: %+v", body)
		}
		for _, forbidden := range []string{"password_hash", "session_token_hash", "messages", "attachments", "audit_logs"} {
			if string(payload) != "" && containsString(string(payload), forbidden) {
				t.Fatalf("detail response exposed forbidden field %q", forbidden)
			}
		}
	})
}

func seedTicketDetailFixture(t *testing.T, pool *pgxpool.Pool, ctx context.Context) ticketDetailFixture {
	t.Helper()
	var fixture ticketDetailFixture
	requesterDepartmentID := insertDepartment(t, pool, ctx, "SPEC07-REQ", "SPEC-07 Request Department")
	fixture.RequestDepartmentID = insertDepartment(t, pool, ctx, "SPEC07-DEST", "SPEC-07 Destination Department")
	fixture.AssignedDepartmentID = insertDepartment(t, pool, ctx, "SPEC07-HK", "SPEC-07 Housekeeping")
	fixture.SecondAssignedDeptID = insertDepartment(t, pool, ctx, "SPEC07-FO", "SPEC-07 Front Office")
	crossDepartmentID := insertDepartment(t, pool, ctx, "SPEC07-CROSS", "SPEC-07 Cross Department")
	fixture.LocationID = insertLocation(t, pool, ctx, "SPEC07-LOC", "SPEC-07 Historical Location")
	fixture.RequesterID = insertUser(t, pool, ctx, "spec07-requester", "SPEC07-REQUESTER", "SPEC-07 Requester", requesterDepartmentID)
	fixture.OwnerID = insertUser(t, pool, ctx, "spec07-owner", "SPEC07-OWNER", "SPEC-07 Owner", fixture.AssignedDepartmentID)
	fixture.AssignedUserID = insertUser(t, pool, ctx, "spec07-assignee-1", "SPEC07-ASSIGNEE-1", "SPEC-07 Assignee One", fixture.AssignedDepartmentID)
	fixture.SecondAssignedUserID = insertUser(t, pool, ctx, "spec07-assignee-2", "SPEC07-ASSIGNEE-2", "SPEC-07 Assignee Two", fixture.SecondAssignedDeptID)
	crossUserID := insertUser(t, pool, ctx, "spec07-cross-user", "SPEC07-CROSS-USER", "SPEC-07 Cross User", crossDepartmentID)

	fixture.RequesterToken = insertSession(t, pool, ctx, fixture.RequesterID)
	fixture.CrossDepartmentToken = insertSession(t, pool, ctx, crossUserID)
	pendingCreatedAt := time.Date(2026, 9, 14, 8, 0, 0, 0, time.UTC)
	acceptedCreatedAt := pendingCreatedAt.Add(-time.Hour)
	closedCreatedAt := pendingCreatedAt.Add(-2 * time.Hour)
	historicalCreatedAt := pendingCreatedAt.Add(-3 * time.Hour)
	acceptedAt := acceptedCreatedAt.Add(10 * time.Minute)
	closedAt := closedCreatedAt.Add(45 * time.Minute)

	if err := pool.QueryRow(ctx, `
INSERT INTO tickets (requester_id, department_id, title, status, priority, created_at, updated_at)
VALUES ($1, $2, 'SPEC-07 pending request', 'pending'::ticket_status, FALSE, $3, $3)
RETURNING id`, fixture.RequesterID, fixture.RequestDepartmentID, pendingCreatedAt).Scan(&fixture.PendingTicketID); err != nil {
		t.Fatalf("insert pending detail ticket: %v", err)
	}
	if err := pool.QueryRow(ctx, `
INSERT INTO tickets (requester_id, department_id, location_id, title, description, status, accepted_by, accepted_at, priority, due_at, created_at, updated_at)
VALUES ($1, $2, $3, 'SPEC-07 accepted request', 'The complete SPEC-07 description.', 'accepted'::ticket_status, $4, $5, TRUE, $6, $7, $7)
RETURNING id`, fixture.RequesterID, fixture.RequestDepartmentID, fixture.LocationID, fixture.OwnerID, acceptedAt, acceptedAt.Add(2*time.Hour), acceptedCreatedAt).Scan(&fixture.AcceptedTicketID); err != nil {
		t.Fatalf("insert accepted detail ticket: %v", err)
	}
	if err := pool.QueryRow(ctx, `
INSERT INTO tickets (requester_id, department_id, title, status, accepted_by, accepted_at, closed_by, closed_at, priority, created_at, updated_at)
VALUES ($1, $2, 'SPEC-07 closed request', 'closed'::ticket_status, $3, $4, $1, $5, FALSE, $6, $6)
RETURNING id`, fixture.RequesterID, fixture.RequestDepartmentID, fixture.OwnerID, closedAt.Add(-20*time.Minute), closedAt, closedCreatedAt).Scan(&fixture.ClosedTicketID); err != nil {
		t.Fatalf("insert closed detail ticket: %v", err)
	}
	if err := pool.QueryRow(ctx, `
INSERT INTO tickets (requester_id, department_id, location_id, title, description, status, accepted_by, accepted_at, priority, created_at, updated_at)
VALUES ($1, $2, $3, 'SPEC-07 historical request', 'Historical detail remains readable.', 'accepted'::ticket_status, $4, $5, FALSE, $6, $6)
RETURNING id`, fixture.RequesterID, fixture.RequestDepartmentID, fixture.LocationID, fixture.OwnerID, acceptedAt, historicalCreatedAt).Scan(&fixture.HistoricalTicketID); err != nil {
		t.Fatalf("insert historical detail ticket: %v", err)
	}

	for _, departmentID := range []int64{fixture.AssignedDepartmentID, fixture.SecondAssignedDeptID} {
		if _, err := pool.Exec(ctx, `INSERT INTO ticket_assigned_departments (ticket_id, department_id) VALUES ($1, $2)`, fixture.AcceptedTicketID, departmentID); err != nil {
			t.Fatalf("insert accepted department assignment: %v", err)
		}
	}
	for _, userID := range []int64{fixture.AssignedUserID, fixture.SecondAssignedUserID} {
		if _, err := pool.Exec(ctx, `INSERT INTO ticket_assigned_users (ticket_id, user_id) VALUES ($1, $2)`, fixture.AcceptedTicketID, userID); err != nil {
			t.Fatalf("insert accepted user assignment: %v", err)
		}
	}
	if _, err := pool.Exec(ctx, `INSERT INTO ticket_assigned_departments (ticket_id, department_id) VALUES ($1, $2)`, fixture.HistoricalTicketID, fixture.AssignedDepartmentID); err != nil {
		t.Fatalf("insert historical department assignment: %v", err)
	}
	if _, err := pool.Exec(ctx, `INSERT INTO ticket_assigned_users (ticket_id, user_id) VALUES ($1, $2)`, fixture.HistoricalTicketID, fixture.AssignedUserID); err != nil {
		t.Fatalf("insert historical user assignment: %v", err)
	}

	for _, departmentID := range []int64{fixture.RequestDepartmentID, fixture.AssignedDepartmentID, fixture.SecondAssignedDeptID} {
		if _, err := pool.Exec(ctx, `UPDATE departments SET is_active = FALSE WHERE id = $1`, departmentID); err != nil {
			t.Fatalf("deactivate historical department %d: %v", departmentID, err)
		}
	}
	if _, err := pool.Exec(ctx, `UPDATE locations SET is_active = FALSE WHERE id = $1`, fixture.LocationID); err != nil {
		t.Fatalf("deactivate historical location: %v", err)
	}
	if _, err := pool.Exec(ctx, `UPDATE users SET is_active = FALSE WHERE id = ANY($1::bigint[])`, []int64{fixture.OwnerID, fixture.AssignedUserID}); err != nil {
		t.Fatalf("deactivate historical users: %v", err)
	}
	return fixture
}

func formatInt64(value int64) string {
	return fmt.Sprintf("%d", value)
}

func containsString(value, needle string) bool {
	return strings.Contains(value, needle)
}
