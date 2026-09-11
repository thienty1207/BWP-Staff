package tickets_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/thienty1207/BWP-Staff/backend/app"
	"github.com/thienty1207/BWP-Staff/backend/config"
	"github.com/thienty1207/BWP-Staff/backend/shared"
	"github.com/thienty1207/BWP-Staff/backend/shared/security"
)

type ticketListResponse struct {
	Tickets []ticketResponse `json:"tickets"`
	Page    ticketPage       `json:"page"`
}

type ticketPage struct {
	HasMore             bool       `json:"has_more"`
	NextBeforeCreatedAt *time.Time `json:"next_before_created_at"`
	NextBeforeID        *int64     `json:"next_before_id"`
}

type ticketResponse struct {
	ID                  int64               `json:"id"`
	Title               string              `json:"title"`
	Status              string              `json:"status"`
	Priority            bool                `json:"priority"`
	DueAt               *time.Time          `json:"due_at"`
	CreatedAt           time.Time           `json:"created_at"`
	UpdatedAt           time.Time           `json:"updated_at"`
	Requester           identitySummary     `json:"requester"`
	Department          departmentSummary   `json:"department"`
	Location            *locationSummary    `json:"location"`
	AcceptedBy          *identitySummary    `json:"accepted_by"`
	AcceptedAt          *time.Time          `json:"accepted_at"`
	AssignedDepartments []departmentSummary `json:"assigned_departments"`
	AssignedUsers       []identitySummary   `json:"assigned_users"`
	ClosedAt            *time.Time          `json:"closed_at"`
}

type identitySummary struct {
	ID       int64  `json:"id"`
	FullName string `json:"full_name"`
}

type departmentSummary struct {
	ID   int64  `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
}

type locationSummary struct {
	ID   int64  `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
}

type ticketFixture struct {
	Token                string
	PrimaryUserID        int64
	PrimaryDepartmentID  int64
	AssignedDepartmentID int64
	SecondAssignedDeptID int64
	PrimaryLocationID    int64
	AssignedUserID       int64
	SecondAssignedUserID int64
	FirstOpenTicketID    int64
	SecondOpenTicketID   int64
	ClosedTicketID       int64
	OpenCreatedAt        time.Time
	ClosedCreatedAt      time.Time
}

func TestSPEC05TicketsListRequiresAuthAndReturnsRealRelationalData(t *testing.T) {
	pool, ctx := openTicketsTestPool(t)
	fixture := seedTicketsFixture(t, pool, ctx)
	server := newTicketsApp(pool)

	t.Run("requires existing session authentication", func(t *testing.T) {
		response := requestTickets(t, server, "/api/v1/tickets?view=open", "")
		defer response.Body.Close()
		if response.StatusCode != http.StatusUnauthorized {
			t.Fatalf("expected unauthenticated status 401, got %d", response.StatusCode)
		}
	})

	t.Run("open view maps fields and assignments without duplication", func(t *testing.T) {
		response := requestTickets(t, server, "/api/v1/tickets?view=open", fixture.Token)
		defer response.Body.Close()
		if response.StatusCode != http.StatusOK {
			t.Fatalf("expected open list status 200, got %d", response.StatusCode)
		}

		var body ticketListResponse
		decodeTicketResponse(t, response, &body)
		if len(body.Tickets) != 2 || body.Page.HasMore || body.Page.NextBeforeID != nil || body.Page.NextBeforeCreatedAt != nil {
			t.Fatalf("unexpected open page: %+v", body)
		}
		newest := body.Tickets[0]
		if newest.ID != fixture.SecondOpenTicketID || newest.Title != "Accepted open request" || newest.Status != "accepted" {
			t.Fatalf("unexpected newest open ticket: %+v", newest)
		}
		if !newest.Priority || newest.DueAt == nil {
			t.Fatalf("priority/due_at were not returned: %+v", newest)
		}
		if newest.Location == nil || newest.Location.ID != fixture.PrimaryLocationID || newest.Location.Code != "LOBBY" {
			t.Fatalf("unexpected joined location: %+v", newest.Location)
		}
		if newest.Requester.ID != fixture.PrimaryUserID || newest.Requester.FullName != "SPEC-05 Requester" {
			t.Fatalf("unexpected requester identity: %+v", newest.Requester)
		}
		if newest.Department.ID != fixture.PrimaryDepartmentID || newest.Department.Code != "SPEC05" {
			t.Fatalf("unexpected original department: %+v", newest.Department)
		}
		if newest.AcceptedBy == nil || newest.AcceptedBy.ID != fixture.AssignedUserID || newest.AcceptedAt == nil {
			t.Fatalf("unexpected accepted identity/time: %+v", newest)
		}
		if len(newest.AssignedDepartments) != 2 || newest.AssignedDepartments[0].ID != fixture.AssignedDepartmentID || newest.AssignedDepartments[1].ID != fixture.SecondAssignedDeptID {
			t.Fatalf("unexpected department assignments: %+v", newest.AssignedDepartments)
		}
		if len(newest.AssignedUsers) != 2 || newest.AssignedUsers[0].ID != fixture.AssignedUserID || newest.AssignedUsers[1].ID != fixture.SecondAssignedUserID {
			t.Fatalf("unexpected user assignments: %+v", newest.AssignedUsers)
		}

		oldest := body.Tickets[1]
		if oldest.ID != fixture.FirstOpenTicketID || oldest.Status != "pending" || oldest.Priority || oldest.DueAt != nil {
			t.Fatalf("unexpected oldest open ticket: %+v", oldest)
		}
		if oldest.Location != nil || oldest.AcceptedBy != nil || oldest.AcceptedAt != nil || oldest.ClosedAt != nil {
			t.Fatalf("nullable ticket fields were not preserved: %+v", oldest)
		}
		encoded, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal compact ticket response: %v", err)
		}
		for _, forbidden := range []string{"description", "password_hash", "session_token_hash", "messages", "attachments"} {
			if strings.Contains(string(encoded), forbidden) {
				t.Fatalf("ticket list response exposed heavy or secret field %q", forbidden)
			}
		}
	})

	t.Run("closed view filters only closed tickets", func(t *testing.T) {
		response := requestTickets(t, server, "/api/v1/tickets?view=closed", fixture.Token)
		defer response.Body.Close()
		if response.StatusCode != http.StatusOK {
			t.Fatalf("expected closed list status 200, got %d", response.StatusCode)
		}
		var body ticketListResponse
		decodeTicketResponse(t, response, &body)
		if len(body.Tickets) != 1 || body.Tickets[0].ID != fixture.ClosedTicketID || body.Tickets[0].Status != "closed" {
			t.Fatalf("unexpected closed list: %+v", body.Tickets)
		}
		if body.Tickets[0].ClosedAt == nil {
			t.Fatal("closed ticket did not return closed_at")
		}
	})

	t.Run("keyset page has no duplicate across pages", func(t *testing.T) {
		firstResponse := requestTickets(t, server, "/api/v1/tickets?view=open&limit=1", fixture.Token)
		defer firstResponse.Body.Close()
		if firstResponse.StatusCode != http.StatusOK {
			t.Fatalf("expected first page status 200, got %d", firstResponse.StatusCode)
		}
		var first ticketListResponse
		decodeTicketResponse(t, firstResponse, &first)
		if len(first.Tickets) != 1 || !first.Page.HasMore || first.Page.NextBeforeCreatedAt == nil || first.Page.NextBeforeID == nil {
			t.Fatalf("unexpected first keyset page: %+v", first)
		}

		secondPath := "/api/v1/tickets?view=open&limit=1&before_created_at=" + url.QueryEscape(first.Page.NextBeforeCreatedAt.Format(time.RFC3339Nano)) + "&before_id=" + fmt.Sprint(*first.Page.NextBeforeID)
		secondResponse := requestTickets(t, server, secondPath, fixture.Token)
		defer secondResponse.Body.Close()
		if secondResponse.StatusCode != http.StatusOK {
			t.Fatalf("expected second page status 200, got %d", secondResponse.StatusCode)
		}
		var second ticketListResponse
		decodeTicketResponse(t, secondResponse, &second)
		if len(second.Tickets) != 1 || second.Page.HasMore || second.Page.NextBeforeID != nil || second.Page.NextBeforeCreatedAt != nil {
			t.Fatalf("unexpected second keyset page: %+v", second)
		}
		if first.Tickets[0].ID == second.Tickets[0].ID {
			t.Fatalf("keyset pages duplicated ticket %d", first.Tickets[0].ID)
		}
		if first.Tickets[0].CreatedAt.Before(second.Tickets[0].CreatedAt) || (first.Tickets[0].CreatedAt.Equal(second.Tickets[0].CreatedAt) && first.Tickets[0].ID <= second.Tickets[0].ID) {
			t.Fatalf("tickets are not ordered by created_at DESC, id DESC: first=%+v second=%+v", first.Tickets[0], second.Tickets[0])
		}
	})

	t.Run("invalid query is safe 400", func(t *testing.T) {
		for _, path := range []string{
			"/api/v1/tickets?view=all",
			"/api/v1/tickets?limit=0",
			"/api/v1/tickets?limit=101",
			"/api/v1/tickets?before_id=1",
			"/api/v1/tickets?before_created_at=2026-09-11T10:00:00Z&before_id=0",
		} {
			response := requestTickets(t, server, path, fixture.Token)
			var body map[string]map[string]string
			decodeTicketResponse(t, response, &body)
			response.Body.Close()
			if response.StatusCode != http.StatusBadRequest {
				t.Fatalf("%s: expected status 400, got %d", path, response.StatusCode)
			}
			if body["error"]["code"] != "invalid_request" || body["error"]["message"] != "Invalid request" {
				t.Fatalf("%s: unexpected safe error: %+v", path, body)
			}
		}
	})
}

func newTicketsApp(pool *pgxpool.Pool) *fiber.App {
	return app.New(app.AppState{DB: pool}, config.Config{
		AppEnv:                        "development",
		FrontendOrigin:                "http://localhost:5173",
		DatabaseAcquireTimeoutSeconds: 1,
		AuthSessionTTLHours:           12,
	})
}

func requestTickets(t *testing.T, server *fiber.App, path, token string) *http.Response {
	t.Helper()
	request := httptest.NewRequest(http.MethodGet, path, nil)
	if token != "" {
		request.AddCookie(&http.Cookie{Name: "bwp_session", Value: token})
	}
	response, err := server.Test(request)
	if err != nil {
		t.Fatalf("GET %s: %v", path, err)
	}
	return response
}

func decodeTicketResponse(t *testing.T, response *http.Response, target any) {
	t.Helper()
	if err := json.NewDecoder(response.Body).Decode(target); err != nil {
		t.Fatalf("decode ticket response: %v", err)
	}
}

func seedTicketsFixture(t *testing.T, pool *pgxpool.Pool, ctx context.Context) ticketFixture {
	t.Helper()
	var fixture ticketFixture
	fixture.OpenCreatedAt = time.Date(2026, 9, 11, 10, 0, 0, 0, time.UTC)
	fixture.ClosedCreatedAt = time.Date(2026, 9, 10, 10, 0, 0, 0, time.UTC)
	fixture.PrimaryDepartmentID = insertDepartment(t, pool, ctx, "SPEC05", "SPEC-05 Primary")
	fixture.AssignedDepartmentID = insertDepartment(t, pool, ctx, "SPEC05-HK", "SPEC-05 Housekeeping")
	fixture.SecondAssignedDeptID = insertDepartment(t, pool, ctx, "SPEC05-FO", "SPEC-05 Front Office")
	fixture.PrimaryLocationID = insertLocation(t, pool, ctx, "LOBBY", "SPEC-05 Lobby")
	fixture.PrimaryUserID = insertUser(t, pool, ctx, "spec05-requester", "SPEC05-REQUESTER", "SPEC-05 Requester", fixture.PrimaryDepartmentID)
	fixture.AssignedUserID = insertUser(t, pool, ctx, "spec05-assignee-1", "SPEC05-ASSIGNEE-1", "SPEC-05 Assignee One", fixture.AssignedDepartmentID)
	fixture.SecondAssignedUserID = insertUser(t, pool, ctx, "spec05-assignee-2", "SPEC05-ASSIGNEE-2", "SPEC-05 Assignee Two", fixture.AssignedDepartmentID)

	fixture.Token = insertSession(t, pool, ctx, fixture.PrimaryUserID)

	var err error
	if err = pool.QueryRow(ctx, `
INSERT INTO tickets (
    requester_id, department_id, title, status, priority, due_at,
    created_at, updated_at
)
VALUES ($1, $2, 'Pending open request', 'pending'::ticket_status, FALSE, NULL, $3, $3)
RETURNING id`, fixture.PrimaryUserID, fixture.PrimaryDepartmentID, fixture.OpenCreatedAt).Scan(&fixture.FirstOpenTicketID); err != nil {
		t.Fatalf("insert pending ticket: %v", err)
	}

	acceptedAt := fixture.OpenCreatedAt.Add(10 * time.Minute)
	dueAt := fixture.OpenCreatedAt.Add(2 * time.Hour)
	if err = pool.QueryRow(ctx, `
INSERT INTO tickets (
    requester_id, department_id, location_id, title, status, accepted_by,
    accepted_at, priority, due_at, created_at, updated_at
)
VALUES ($1, $2, $3, 'Accepted open request', 'accepted'::ticket_status, $4, $5, TRUE, $6, $7, $7)
RETURNING id`, fixture.PrimaryUserID, fixture.PrimaryDepartmentID, fixture.PrimaryLocationID, fixture.AssignedUserID, acceptedAt, dueAt, fixture.OpenCreatedAt).Scan(&fixture.SecondOpenTicketID); err != nil {
		t.Fatalf("insert accepted ticket: %v", err)
	}

	closedAt := fixture.ClosedCreatedAt.Add(30 * time.Minute)
	if err = pool.QueryRow(ctx, `
INSERT INTO tickets (
    requester_id, department_id, title, status, accepted_by, accepted_at,
    closed_by, closed_at, priority, created_at, updated_at
)
VALUES ($1, $2, 'Closed request', 'closed'::ticket_status, $3, $4, $1, $5, FALSE, $6, $6)
RETURNING id`, fixture.PrimaryUserID, fixture.PrimaryDepartmentID, fixture.AssignedUserID, closedAt.Add(-20*time.Minute), closedAt, fixture.ClosedCreatedAt).Scan(&fixture.ClosedTicketID); err != nil {
		t.Fatalf("insert closed ticket: %v", err)
	}

	for _, departmentID := range []int64{fixture.AssignedDepartmentID, fixture.SecondAssignedDeptID} {
		if _, err := pool.Exec(ctx, `
INSERT INTO ticket_assigned_departments (ticket_id, department_id)
VALUES ($1, $2)`, fixture.SecondOpenTicketID, departmentID); err != nil {
			t.Fatalf("insert ticket department assignment: %v", err)
		}
	}
	for _, userID := range []int64{fixture.AssignedUserID, fixture.SecondAssignedUserID} {
		if _, err := pool.Exec(ctx, `
INSERT INTO ticket_assigned_users (ticket_id, user_id)
VALUES ($1, $2)`, fixture.SecondOpenTicketID, userID); err != nil {
			t.Fatalf("insert ticket user assignment: %v", err)
		}
	}
	return fixture
}

func insertDepartment(t *testing.T, pool *pgxpool.Pool, ctx context.Context, code, name string) int64 {
	t.Helper()
	var id int64
	if err := pool.QueryRow(ctx, `
INSERT INTO departments (code, name)
VALUES ($1, $2)
RETURNING id`, code, name).Scan(&id); err != nil {
		t.Fatalf("insert department %s: %v", code, err)
	}
	return id
}

func insertLocation(t *testing.T, pool *pgxpool.Pool, ctx context.Context, code, name string) int64 {
	t.Helper()
	var id int64
	if err := pool.QueryRow(ctx, `
INSERT INTO locations (code, name)
VALUES ($1, $2)
RETURNING id`, code, name).Scan(&id); err != nil {
		t.Fatalf("insert location %s: %v", code, err)
	}
	return id
}

func insertUser(t *testing.T, pool *pgxpool.Pool, ctx context.Context, username, employeeCode, fullName string, departmentID int64) int64 {
	t.Helper()
	hash, err := security.HashPassword("spec-05-test-password")
	if err != nil {
		t.Fatalf("hash test password: %v", err)
	}
	var id int64
	if err := pool.QueryRow(ctx, `
INSERT INTO users (username, employee_code, password_hash, full_name, department_id, role)
VALUES ($1, $2, $3, $4, $5, 'staff'::user_role)
RETURNING id`, username, employeeCode, hash, fullName, departmentID).Scan(&id); err != nil {
		t.Fatalf("insert user %s: %v", username, err)
	}
	return id
}

func insertSession(t *testing.T, pool *pgxpool.Pool, ctx context.Context, userID int64) string {
	t.Helper()
	rawToken, tokenHash, err := security.GenerateSessionToken()
	if err != nil {
		t.Fatalf("generate test session token: %v", err)
	}
	if _, err := pool.Exec(ctx, `
INSERT INTO auth_sessions (session_token_hash, user_id, expires_at)
VALUES ($1, $2, NOW() + INTERVAL '1 hour')`, tokenHash, userID); err != nil {
		t.Fatalf("insert test session: %v", err)
	}
	return rawToken
}

func openTicketsTestPool(t *testing.T) (*pgxpool.Pool, context.Context) {
	t.Helper()
	loadLocalTicketsTestEnv(t)
	databaseURL := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if databaseURL == "" {
		t.Skip("set DATABASE_URL to run PostgreSQL-backed SPEC-05 tests")
	}
	if appEnv := strings.TrimSpace(os.Getenv("APP_ENV")); appEnv != "" && !strings.EqualFold(appEnv, "development") {
		t.Skipf("refusing PostgreSQL-backed SPEC-05 test with APP_ENV=%q", appEnv)
	}
	parsedURL, err := url.Parse(databaseURL)
	if err != nil || !isLoopbackDatabaseURL(parsedURL) {
		t.Skip("refusing SPEC-05 test unless DATABASE_URL points to a loopback PostgreSQL instance")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	schema := fmt.Sprintf("spec05_%d", time.Now().UnixNano())
	quotedSchema := pgx.Identifier{schema}.Sanitize()
	adminConnection, err := pgx.Connect(ctx, databaseURL)
	if err != nil {
		t.Skipf("PostgreSQL is unavailable for SPEC-05 tests: %v", err)
	}
	if _, err := adminConnection.Exec(ctx, "CREATE SCHEMA "+quotedSchema); err != nil {
		_ = adminConnection.Close(ctx)
		t.Skipf("cannot create isolated SPEC-05 schema: %v", err)
	}
	if err := adminConnection.Close(ctx); err != nil {
		t.Fatalf("close PostgreSQL setup connection: %v", err)
	}

	query := parsedURL.Query()
	query.Del("options")
	parsedURL.RawQuery = query.Encode()
	if parsedURL.RawQuery != "" {
		parsedURL.RawQuery += "&"
	}
	parsedURL.RawQuery += "options=" + strings.ReplaceAll(url.QueryEscape("-c search_path="+schema+",public"), "+", "%20")

	poolConfig, err := pgxpool.ParseConfig(parsedURL.String())
	if err != nil {
		dropTicketsTestSchema(t, ctx, databaseURL, quotedSchema)
		t.Fatalf("parse isolated SPEC-05 pool config: %v", err)
	}
	poolConfig.MaxConns = 4
	poolConfig.MinConns = 0
	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		dropTicketsTestSchema(t, ctx, databaseURL, quotedSchema)
		t.Skipf("cannot create SPEC-05 PostgreSQL pool: %v", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		dropTicketsTestSchema(t, ctx, databaseURL, quotedSchema)
		t.Skipf("PostgreSQL is unavailable for SPEC-05 tests: %v", err)
	}
	if err := shared.RunMigrations(ctx, pool, filepath.Join("..", "..", "migrations"), 5*time.Second); err != nil {
		pool.Close()
		dropTicketsTestSchema(t, ctx, databaseURL, quotedSchema)
		t.Fatalf("run SPEC-05 test migrations: %v", err)
	}
	t.Cleanup(func() {
		pool.Close()
		dropTicketsTestSchema(t, context.Background(), databaseURL, quotedSchema)
		cancel()
	})
	return pool, ctx
}

func loadLocalTicketsTestEnv(t *testing.T) {
	t.Helper()
	if strings.TrimSpace(os.Getenv("DATABASE_URL")) != "" {
		return
	}
	if err := godotenv.Load(filepath.Join("..", "..", ".env")); err != nil && !os.IsNotExist(err) {
		t.Fatalf("load backend/.env for SPEC-05 test: %v", err)
	}
}

func dropTicketsTestSchema(t *testing.T, ctx context.Context, databaseURL, quotedSchema string) {
	t.Helper()
	cleanupContext, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	connection, err := pgx.Connect(cleanupContext, databaseURL)
	if err != nil {
		t.Errorf("connect PostgreSQL cleanup connection: %v", err)
		return
	}
	defer connection.Close(cleanupContext)
	if _, err := connection.Exec(cleanupContext, "DROP SCHEMA "+quotedSchema+" CASCADE"); err != nil {
		t.Errorf("drop isolated SPEC-05 schema: %v", err)
	}
}

func isLoopbackDatabaseURL(databaseURL *url.URL) bool {
	if databaseURL == nil || (databaseURL.Scheme != "postgres" && databaseURL.Scheme != "postgresql") {
		return false
	}
	host := strings.TrimSuffix(strings.ToLower(databaseURL.Hostname()), ".")
	if host == "localhost" || host == "127.0.0.1" || host == "::1" {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}
