package auth_test

import (
	"bytes"
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

const (
	testPassword       = "test-only-correct-password"
	activeUsername     = "spec03-active"
	inactiveUsername   = "spec03-inactive"
	activeEmployeeCode = "SPEC03-ACTIVE"
	inactiveEmployee   = "SPEC03-INACTIVE"
)

type authFixture struct {
	DepartmentID int64
	ActiveUserID int64
	InactiveID   int64
}

type errorResponse struct {
	Error struct {
		Code      string `json:"code"`
		Message   string `json:"message"`
		RequestID string `json:"request_id"`
	} `json:"error"`
}

type userResponse struct {
	User struct {
		ID           int64  `json:"id"`
		Username     string `json:"username"`
		EmployeeCode string `json:"employee_code"`
		FullName     string `json:"full_name"`
		Role         string `json:"role"`
		Department   struct {
			ID   int64  `json:"id"`
			Code string `json:"code"`
			Name string `json:"name"`
		} `json:"department"`
		AvatarURL *string `json:"avatar_url"`
	} `json:"user"`
}

func TestSPEC03LoginCreatesSecureServerSideSession(t *testing.T) {
	pool, ctx := openAuthTestPool(t)
	fixture := seedAuthFixture(t, pool, ctx)
	server := newAuthApp(pool, 2)

	request := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(`{"username":"  spec03-active  ","password":"test-only-correct-password"}`))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("User-Agent", strings.Repeat("agent/1 ", 100))
	request.Header.Set("X-Forwarded-For", "203.0.113.99")
	request.RemoteAddr = "198.51.100.24:54321"
	response, err := server.Test(request)
	if err != nil {
		t.Fatalf("login request: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		t.Fatalf("expected login status 200, got %d", response.StatusCode)
	}
	var body userResponse
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode login response: %v", err)
	}
	if body.User.Username != activeUsername || body.User.EmployeeCode != activeEmployeeCode || body.User.Role != "admin" {
		t.Fatalf("unexpected public user identity: %+v", body.User)
	}
	if body.User.Department.ID != fixture.DepartmentID || body.User.Department.Code != "IT" || body.User.Department.Name != "IT Department" {
		t.Fatalf("unexpected public department: %+v", body.User.Department)
	}
	encodedBody, _ := json.Marshal(body)
	for _, forbidden := range []string{"password_hash", "session_token_hash", testPassword} {
		if strings.Contains(string(encodedBody), forbidden) {
			t.Fatalf("login response exposed %q", forbidden)
		}
	}

	cookie := cookieNamed(response, "bwp_session")
	if cookie == nil || cookie.Value == "" {
		t.Fatal("expected non-empty bwp_session cookie")
	}
	if !cookie.HttpOnly || cookie.Path != "/" || cookie.SameSite != http.SameSiteLaxMode || cookie.Domain != "" {
		t.Fatalf("unexpected development session cookie policy: %+v", cookie)
	}
	if cookie.Secure || cookie.MaxAge <= 0 || cookie.MaxAge > 2*60*60 {
		t.Fatalf("unexpected development session cookie lifetime/security: %+v", cookie)
	}

	var storedHash, storedUserAgent, storedIP string
	var sessionUserID int64
	var createdAt, expiresAt time.Time
	var lastUsedAt *time.Time
	if err := pool.QueryRow(ctx, `
SELECT session_token_hash, user_id, created_at, expires_at, COALESCE(user_agent, ''), COALESCE(host(ip_address), ''), last_used_at
FROM auth_sessions
WHERE user_id = $1`, fixture.ActiveUserID).Scan(&storedHash, &sessionUserID, &createdAt, &expiresAt, &storedUserAgent, &storedIP, &lastUsedAt); err != nil {
		t.Fatalf("read created session: %v", err)
	}
	if storedHash != security.HashSessionToken(cookie.Value) || storedHash == cookie.Value || len(storedHash) != 64 {
		t.Fatalf("database did not store only the expected token hash: %q", storedHash)
	}
	if sessionUserID != fixture.ActiveUserID {
		t.Fatalf("session belongs to user %d, want %d", sessionUserID, fixture.ActiveUserID)
	}
	if lifetime := expiresAt.Sub(createdAt); lifetime < 2*time.Hour-time.Minute || lifetime > 2*time.Hour {
		t.Fatalf("session lifetime %s does not reflect configured TTL", lifetime)
	}
	if len(storedUserAgent) > 512 || storedUserAgent == "" {
		t.Fatalf("unexpected capped user-agent metadata length: %d", len(storedUserAgent))
	}
	if net.ParseIP(storedIP) == nil || storedIP == "203.0.113.99" {
		t.Fatalf("unexpected direct client IP metadata: %q", storedIP)
	}
	if lastUsedAt != nil {
		t.Fatalf("new session unexpectedly populated last_used_at: %v", lastUsedAt)
	}

	var lastLoginAt *time.Time
	if err := pool.QueryRow(ctx, "SELECT last_login_at FROM users WHERE id = $1", fixture.ActiveUserID).Scan(&lastLoginAt); err != nil {
		t.Fatalf("read last_login_at: %v", err)
	}
	if lastLoginAt == nil {
		t.Fatal("successful login did not update last_login_at")
	}
}

func TestSPEC03LoginRejectsMalformedAndInvalidInput(t *testing.T) {
	server := newAuthApp(nil, 2)
	cases := []struct {
		name string
		body string
	}{
		{name: "malformed json", body: "{"},
		{name: "missing username", body: `{"password":"secret"}`},
		{name: "missing password", body: `{"username":"user"}`},
		{name: "empty username", body: `{"username":"   ","password":"secret"}`},
		{name: "empty password", body: `{"username":"user","password":""}`},
		{name: "username too long", body: fmt.Sprintf(`{"username":"%s","password":"secret"}`, strings.Repeat("u", 51))},
		{name: "password too long", body: fmt.Sprintf(`{"username":"user","password":"%s"}`, strings.Repeat("p", 1025))},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(testCase.body))
			request.Header.Set("Content-Type", "application/json")
			response, err := server.Test(request)
			if err != nil {
				t.Fatalf("invalid login request: %v", err)
			}
			defer response.Body.Close()

			if response.StatusCode != http.StatusBadRequest {
				t.Fatalf("expected status 400, got %d", response.StatusCode)
			}
			var body errorResponse
			if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
				t.Fatalf("decode invalid request response: %v", err)
			}
			if body.Error.Code != "invalid_request" || body.Error.Message != "Invalid request" {
				t.Fatalf("unexpected invalid request error: %+v", body.Error)
			}
			if body.Error.RequestID == "" || body.Error.RequestID != response.Header.Get("X-Request-ID") {
				t.Fatalf("request ID mismatch in invalid request response: header=%q body=%q", response.Header.Get("X-Request-ID"), body.Error.RequestID)
			}
			if cookieNamed(response, "bwp_session") != nil {
				t.Fatal("invalid login unexpectedly set a session cookie")
			}
		})
	}
}

func TestSPEC03CredentialFailuresAreIndistinguishableAndDoNotWriteLoginState(t *testing.T) {
	pool, ctx := openAuthTestPool(t)
	fixture := seedAuthFixture(t, pool, ctx)
	server := newAuthApp(pool, 2)

	cases := []struct {
		name     string
		username string
		password string
	}{
		{name: "wrong password", username: activeUsername, password: "wrong-password"},
		{name: "unknown username", username: "spec03-unknown", password: testPassword},
		{name: "inactive user", username: inactiveUsername, password: testPassword},
		{name: "email is not an identifier", username: "spec03-active@example.invalid", password: testPassword},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			requestBody, _ := json.Marshal(loginRequestForTest{Username: testCase.username, Password: testCase.password})
			request := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(requestBody))
			request.Header.Set("Content-Type", "application/json")
			response, err := server.Test(request)
			if err != nil {
				t.Fatalf("credential failure request: %v", err)
			}
			defer response.Body.Close()

			if response.StatusCode != http.StatusUnauthorized {
				t.Fatalf("expected status 401, got %d", response.StatusCode)
			}
			var body errorResponse
			if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
				t.Fatalf("decode credential failure: %v", err)
			}
			if body.Error.Code != "invalid_credentials" || body.Error.Message != "Invalid username or password" {
				t.Fatalf("unexpected credential failure response: %+v", body.Error)
			}
			if body.Error.RequestID == "" || body.Error.RequestID != response.Header.Get("X-Request-ID") {
				t.Fatalf("request ID mismatch in credential failure: header=%q body=%q", response.Header.Get("X-Request-ID"), body.Error.RequestID)
			}
			if cookieNamed(response, "bwp_session") != nil {
				t.Fatal("credential failure unexpectedly set a session cookie")
			}
		})
	}

	var sessionCount int
	if err := pool.QueryRow(ctx, "SELECT COUNT(*) FROM auth_sessions").Scan(&sessionCount); err != nil {
		t.Fatalf("count sessions after credential failures: %v", err)
	}
	if sessionCount != 0 {
		t.Fatalf("credential failures created %d sessions", sessionCount)
	}
	var lastLoginAt *time.Time
	if err := pool.QueryRow(ctx, "SELECT last_login_at FROM users WHERE id = $1", fixture.ActiveUserID).Scan(&lastLoginAt); err != nil {
		t.Fatalf("read failed-login timestamp: %v", err)
	}
	if lastLoginAt != nil {
		t.Fatal("failed login updated last_login_at")
	}
}

func TestSPEC03MultipleSessionsRemainIndependent(t *testing.T) {
	pool, ctx := openAuthTestPool(t)
	fixture := seedAuthFixture(t, pool, ctx)
	server := newAuthApp(pool, 2)

	first := performLogin(t, server, activeUsername, testPassword)
	second := performLogin(t, server, activeUsername, testPassword)
	if first.Value == second.Value {
		t.Fatal("independent logins returned the same session token")
	}

	for name, cookie := range map[string]*http.Cookie{"first": first, "second": second} {
		t.Run(name, func(t *testing.T) {
			response := performAuthenticatedRequest(t, server, http.MethodGet, "/api/v1/auth/me", cookie.Value)
			defer response.Body.Close()
			if response.StatusCode != http.StatusOK {
				t.Fatalf("expected active session to authenticate, got %d", response.StatusCode)
			}
		})
	}

	var sessionCount int
	if err := pool.QueryRow(ctx, "SELECT COUNT(*) FROM auth_sessions WHERE user_id = $1", fixture.ActiveUserID).Scan(&sessionCount); err != nil {
		t.Fatalf("count independent sessions: %v", err)
	}
	if sessionCount != 2 {
		t.Fatalf("expected two independent sessions, got %d", sessionCount)
	}
	var revokedCount int
	if err := pool.QueryRow(ctx, "SELECT COUNT(*) FROM auth_sessions WHERE user_id = $1 AND revoked_at IS NOT NULL", fixture.ActiveUserID).Scan(&revokedCount); err != nil {
		t.Fatalf("count revoked sessions: %v", err)
	}
	if revokedCount != 0 {
		t.Fatalf("login unexpectedly revoked %d existing sessions", revokedCount)
	}
}

func TestSPEC03MeRejectsInvalidSessionsAndRechecksActiveUser(t *testing.T) {
	pool, ctx := openAuthTestPool(t)
	fixture := seedAuthFixture(t, pool, ctx)
	server := newAuthApp(pool, 2)

	validCookie := performLogin(t, server, activeUsername, testPassword)
	expiredToken := insertManualSession(t, pool, ctx, fixture.ActiveUserID, time.Now().Add(-2*time.Hour), time.Now().Add(-time.Hour), nil)
	revokedAt := time.Now().UTC()
	revokedToken := insertManualSession(t, pool, ctx, fixture.ActiveUserID, time.Now().Add(-2*time.Hour), time.Now().Add(time.Hour), &revokedAt)
	unknownToken, _, err := security.GenerateSessionToken()
	if err != nil {
		t.Fatalf("generate unknown session token: %v", err)
	}

	cases := []struct {
		name  string
		token string
	}{
		{name: "missing cookie", token: ""},
		{name: "malformed cookie", token: "not-a-session-token"},
		{name: "unknown token", token: unknownToken},
		{name: "expired session", token: expiredToken},
		{name: "revoked session", token: revokedToken},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			response := performAuthenticatedRequest(t, server, http.MethodGet, "/api/v1/auth/me", testCase.token)
			defer response.Body.Close()
			assertUnauthenticated(t, response)
		})
	}

	activeResponse := performAuthenticatedRequest(t, server, http.MethodGet, "/api/v1/auth/me", validCookie.Value)
	if activeResponse.StatusCode != http.StatusOK {
		activeResponse.Body.Close()
		t.Fatalf("valid session was not accepted before deactivation: %d", activeResponse.StatusCode)
	}
	activeResponse.Body.Close()
	if _, err := pool.Exec(ctx, "UPDATE users SET is_active = FALSE WHERE id = $1", fixture.ActiveUserID); err != nil {
		t.Fatalf("deactivate test user: %v", err)
	}
	deactivatedResponse := performAuthenticatedRequest(t, server, http.MethodGet, "/api/v1/auth/me", validCookie.Value)
	defer deactivatedResponse.Body.Close()
	assertUnauthenticated(t, deactivatedResponse)
}

func TestSPEC03LogoutRevokesSessionAndIsIdempotent(t *testing.T) {
	pool, ctx := openAuthTestPool(t)
	fixture := seedAuthFixture(t, pool, ctx)
	server := newAuthApp(pool, 2)

	cookie := performLogin(t, server, activeUsername, testPassword)
	logoutResponse := performAuthenticatedRequest(t, server, http.MethodPost, "/api/v1/auth/logout", cookie.Value)
	if logoutResponse.StatusCode != http.StatusNoContent {
		logoutResponse.Body.Close()
		t.Fatalf("expected logout status 204, got %d", logoutResponse.StatusCode)
	}
	logoutResponse.Body.Close()
	deletionCookie := cookieNamed(logoutResponse, "bwp_session")
	if deletionCookie == nil || deletionCookie.Value != "" || deletionCookie.MaxAge >= 0 || deletionCookie.Path != "/" || !deletionCookie.HttpOnly || deletionCookie.SameSite != http.SameSiteLaxMode || deletionCookie.Secure {
		t.Fatalf("unexpected logout deletion cookie: %+v", deletionCookie)
	}

	var revokedAt *time.Time
	var sessionCount int
	if err := pool.QueryRow(ctx, "SELECT COUNT(*), MAX(revoked_at) FROM auth_sessions WHERE user_id = $1", fixture.ActiveUserID).Scan(&sessionCount, &revokedAt); err != nil {
		t.Fatalf("read revoked session: %v", err)
	}
	if sessionCount != 1 || revokedAt == nil {
		t.Fatalf("logout did not persist revocation without deleting row: count=%d revoked_at=%v", sessionCount, revokedAt)
	}

	oldSessionResponse := performAuthenticatedRequest(t, server, http.MethodGet, "/api/v1/auth/me", cookie.Value)
	defer oldSessionResponse.Body.Close()
	assertUnauthenticated(t, oldSessionResponse)

	cases := []struct {
		name  string
		token string
	}{
		{name: "missing cookie", token: ""},
		{name: "malformed cookie", token: "malformed"},
		{name: "unknown cookie", token: mustGenerateToken(t)},
		{name: "already revoked cookie", token: cookie.Value},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			response := performAuthenticatedRequest(t, server, http.MethodPost, "/api/v1/auth/logout", testCase.token)
			defer response.Body.Close()
			if response.StatusCode != http.StatusNoContent {
				t.Fatalf("expected idempotent logout status 204, got %d", response.StatusCode)
			}
			if cookieNamed(response, "bwp_session") == nil {
				t.Fatal("idempotent logout did not clear the browser cookie")
			}
		})
	}
}

func TestSPEC03InternalDatabaseFailureReturnsSafeErrorWithoutCookie(t *testing.T) {
	pool, _ := openAuthTestPool(t)
	pool.Close()
	server := newAuthApp(pool, 2)

	request := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(`{"username":"spec03-active","password":"test-only-correct-password"}`))
	request.Header.Set("Content-Type", "application/json")
	response, err := server.Test(request)
	if err != nil {
		t.Fatalf("request with closed database pool: %v", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusInternalServerError {
		t.Fatalf("expected safe status 500, got %d", response.StatusCode)
	}
	var body errorResponse
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode internal error response: %v", err)
	}
	if body.Error.Code != "internal_server_error" || body.Error.Message != "Internal server error" {
		t.Fatalf("unexpected internal error response: %+v", body.Error)
	}
	if cookieNamed(response, "bwp_session") != nil {
		t.Fatal("internal login failure returned a valid session cookie")
	}
}

type loginRequestForTest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func newAuthApp(pool *pgxpool.Pool, ttlHours int) *fiber.App {
	return app.New(app.AppState{DB: pool}, config.Config{
		AppEnv:                        "development",
		FrontendOrigin:                "http://localhost:5173",
		DatabaseAcquireTimeoutSeconds: 1,
		AuthSessionTTLHours:           ttlHours,
	})
}

func performLogin(t *testing.T, server *fiber.App, username, password string) *http.Cookie {
	t.Helper()
	body, err := json.Marshal(loginRequestForTest{Username: username, Password: password})
	if err != nil {
		t.Fatalf("marshal login request: %v", err)
	}
	request := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	response, err := server.Test(request)
	if err != nil {
		t.Fatalf("login request: %v", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		t.Fatalf("expected successful login, got %d", response.StatusCode)
	}
	cookie := cookieNamed(response, "bwp_session")
	if cookie == nil || cookie.Value == "" {
		t.Fatal("successful login returned no session cookie")
	}
	return cookie
}

func performAuthenticatedRequest(t *testing.T, server *fiber.App, method, path, token string) *http.Response {
	t.Helper()
	request := httptest.NewRequest(method, path, nil)
	if token != "" {
		request.AddCookie(&http.Cookie{Name: "bwp_session", Value: token})
	}
	response, err := server.Test(request)
	if err != nil {
		t.Fatalf("authenticated request %s %s: %v", method, path, err)
	}
	return response
}

func assertUnauthenticated(t *testing.T, response *http.Response) {
	t.Helper()
	if response.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", response.StatusCode)
	}
	var body errorResponse
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode unauthenticated response: %v", err)
	}
	if body.Error.Code != "unauthenticated" || body.Error.Message != "Authentication required" {
		t.Fatalf("unexpected unauthenticated response: %+v", body.Error)
	}
	if body.Error.RequestID == "" || body.Error.RequestID != response.Header.Get("X-Request-ID") {
		t.Fatalf("request ID mismatch in unauthenticated response: header=%q body=%q", response.Header.Get("X-Request-ID"), body.Error.RequestID)
	}
}

func cookieNamed(response *http.Response, name string) *http.Cookie {
	for _, cookie := range response.Cookies() {
		if cookie.Name == name {
			return cookie
		}
	}
	return nil
}

func insertManualSession(t *testing.T, pool *pgxpool.Pool, ctx context.Context, userID int64, createdAt, expiresAt time.Time, revokedAt *time.Time) string {
	t.Helper()
	rawToken, tokenHash, err := security.GenerateSessionToken()
	if err != nil {
		t.Fatalf("generate manual session token: %v", err)
	}
	if _, err := pool.Exec(ctx, `
INSERT INTO auth_sessions (session_token_hash, user_id, expires_at, created_at, revoked_at)
VALUES ($1, $2, $3, $4, $5)`, tokenHash, userID, expiresAt, createdAt, revokedAt); err != nil {
		t.Fatalf("insert manual session: %v", err)
	}
	return rawToken
}

func mustGenerateToken(t *testing.T) string {
	t.Helper()
	rawToken, _, err := security.GenerateSessionToken()
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}
	return rawToken
}

func seedAuthFixture(t *testing.T, pool *pgxpool.Pool, ctx context.Context) authFixture {
	t.Helper()
	var fixture authFixture
	if err := pool.QueryRow(ctx, `
INSERT INTO departments (code, name)
VALUES ('IT', 'IT Department')
RETURNING id`).Scan(&fixture.DepartmentID); err != nil {
		t.Fatalf("insert auth test department: %v", err)
	}
	passwordHash, err := security.HashPassword(testPassword)
	if err != nil {
		t.Fatalf("hash auth test password: %v", err)
	}
	if err := pool.QueryRow(ctx, `
INSERT INTO users (username, employee_code, email, password_hash, full_name, department_id, role, is_active)
VALUES ($1, $2, $3, $4, 'Active Test User', $5, 'admin'::user_role, TRUE)
RETURNING id`, activeUsername, activeEmployeeCode, "spec03-active@example.invalid", passwordHash, fixture.DepartmentID).Scan(&fixture.ActiveUserID); err != nil {
		t.Fatalf("insert active auth test user: %v", err)
	}
	if err := pool.QueryRow(ctx, `
INSERT INTO users (username, employee_code, email, password_hash, full_name, department_id, role, is_active)
VALUES ($1, $2, $3, $4, 'Inactive Test User', $5, 'staff'::user_role, FALSE)
RETURNING id`, inactiveUsername, inactiveEmployee, "spec03-inactive@example.invalid", passwordHash, fixture.DepartmentID).Scan(&fixture.InactiveID); err != nil {
		t.Fatalf("insert inactive auth test user: %v", err)
	}
	return fixture
}

func openAuthTestPool(t *testing.T) (*pgxpool.Pool, context.Context) {
	t.Helper()
	loadLocalAuthTestEnv(t)
	databaseURL := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if databaseURL == "" {
		t.Skip("set DATABASE_URL to run PostgreSQL-backed SPEC-03 tests")
	}
	if appEnv := strings.TrimSpace(os.Getenv("APP_ENV")); appEnv != "" && !strings.EqualFold(appEnv, "development") {
		t.Skipf("refusing PostgreSQL-backed SPEC-03 test with APP_ENV=%q", appEnv)
	}
	parsedURL, err := url.Parse(databaseURL)
	if err != nil || !isLoopbackDatabaseURL(parsedURL) {
		t.Skip("refusing SPEC-03 test unless DATABASE_URL points to a loopback PostgreSQL instance")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	t.Cleanup(cancel)
	schema := fmt.Sprintf("spec03_%d", time.Now().UnixNano())
	quotedSchema := pgx.Identifier{schema}.Sanitize()
	adminConnection, err := pgx.Connect(ctx, databaseURL)
	if err != nil {
		t.Skipf("PostgreSQL is unavailable for SPEC-03 tests: %v", err)
	}
	if _, err := adminConnection.Exec(ctx, "CREATE SCHEMA "+quotedSchema); err != nil {
		_ = adminConnection.Close(ctx)
		t.Skipf("cannot create isolated SPEC-03 schema: %v", err)
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

	p, err := pgxpool.ParseConfig(parsedURL.String())
	if err != nil {
		dropAuthTestSchema(t, ctx, databaseURL, quotedSchema)
		t.Fatalf("parse isolated SPEC-03 pool config: %v", err)
	}
	p.MaxConns = 2
	p.MinConns = 0
	pool, err := pgxpool.NewWithConfig(ctx, p)
	if err != nil {
		dropAuthTestSchema(t, ctx, databaseURL, quotedSchema)
		t.Skipf("cannot create SPEC-03 PostgreSQL pool: %v", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		dropAuthTestSchema(t, ctx, databaseURL, quotedSchema)
		t.Skipf("PostgreSQL is unavailable for SPEC-03 tests: %v", err)
	}
	if err := shared.RunMigrations(ctx, pool, filepath.Join("..", "..", "migrations"), 5*time.Second); err != nil {
		pool.Close()
		dropAuthTestSchema(t, ctx, databaseURL, quotedSchema)
		t.Fatalf("run SPEC-03 test migrations: %v", err)
	}
	t.Cleanup(func() {
		pool.Close()
		dropAuthTestSchema(t, context.Background(), databaseURL, quotedSchema)
	})
	return pool, ctx
}

func loadLocalAuthTestEnv(t *testing.T) {
	t.Helper()
	if strings.TrimSpace(os.Getenv("DATABASE_URL")) != "" {
		return
	}
	if err := godotenv.Load(filepath.Join("..", "..", ".env")); err != nil && !os.IsNotExist(err) {
		t.Fatalf("load backend/.env for SPEC-03 test: %v", err)
	}
}

func dropAuthTestSchema(t *testing.T, ctx context.Context, databaseURL, quotedSchema string) {
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
		t.Errorf("drop isolated SPEC-03 schema: %v", err)
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
