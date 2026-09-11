# SPEC-02 — Backend Foundation

> **Project:** BWP SonaSea  
> **Phase:** Backend foundation after SPEC-01 Database Foundation  
> **Authoritative context:** `Context-Spec-BWP-SonaSea/PROJECT_CONTEXT.md`
>
> Every coding agent must read `PROJECT_CONTEXT.md` first, then this SPEC, then inspect the current repository before editing.
>
> This SPEC builds the reusable Go/Fiber HTTP foundation only. It does not implement authentication, tickets, chat, uploads, frontend features, or any other product feature.

---

# 1. Goal

Build a small, production-oriented backend foundation around the PostgreSQL foundation completed in SPEC-01.

After SPEC-02, the project should have a stable base for later feature specs:

```text
Process startup
    ↓
Configuration
    ↓
PostgreSQL / pgxpool
    ↓
Migrations
    ↓
AppState
    ↓
Fiber application
    ↓
Foundation middleware
    ↓
Routes
    ↓
Central error handling
    ↓
Graceful shutdown
```

The foundation must remain:

```text
boring
explicit
small
testable
production-oriented
easy to trace
```

Do not build feature architecture prematurely.

---

# 2. Current Repository Baseline

At the time this SPEC is written, `main` has already completed the revised SPEC-01 database foundation.

The current backend already contains:

```text
backend/
├── app/
│   ├── app.go
│   └── app_test.go
├── cmd/
│   ├── server/main.go
│   └── seed_development/main.go
├── config/
│   ├── config.go
│   └── config_test.go
├── shared/
│   ├── database.go
│   ├── foundation_test.go
│   ├── migrations_test.go
│   └── security/
├── admin/
├── migrations/
└── go.mod
```

Current behavior includes:

```text
config loading
PostgreSQL connection with pgxpool
migration execution
development seed command
minimal Fiber app
GET /health -> "ok"
server listen address hard-coded to 127.0.0.1:3000
```

SPEC-02 must evolve this existing foundation rather than replacing it with a new architecture.

Do not rewrite the migration system or revised SPEC-01 schema unless a concrete SPEC-02 defect requires a minimal compatibility change.

---

# 3. Locked Technology

Use the project stack exactly as defined by `PROJECT_CONTEXT.md`.

Backend:

```text
Go
Fiber v3
context.Context
pgx v5
pgxpool
standard log
```

Database:

```text
PostgreSQL
explicit SQL
```

Do not add an ORM.

Do not replace Fiber.

Do not replace pgx/pgxpool.

Do not introduce a dependency injection framework.

---

# 4. Scope

SPEC-02 implements only shared backend infrastructure.

Required scope:

```text
configuration required by the HTTP server
AppState
Fiber app construction
central error response model
central Fiber error handling
request ID
request logging
panic recovery
CORS
liveness endpoint
readiness endpoint
/api/v1 routing foundation
server bind configuration
startup lifecycle
graceful shutdown
foundation tests
```

This SPEC may reorganize the small existing `backend/app` package when needed for clear responsibility, but must not create empty future feature folders.

---

# 5. Non-Goals

SPEC-02 must NOT implement:

```text
login
logout
/auth/me
password verification for login
session creation
session cookies
authorization middleware
user CRUD
ticket CRUD
New Request API
Accept
Assign
Close
ticket chat
WebSocket
checklist API
notifications
Staff Meal API
announcements API
Report
Settings API
Admin UI
file upload API
object storage
SvelteKit UI
Redis
background queues
Docker production deployment
Caddy
Prometheus
Grafana
Kubernetes
microservices
rate limiting
```

Do not create placeholder handlers for these features.

Do not return fake JSON to simulate future functionality.

---

# 6. Foundation Architecture

The desired foundation flow is:

```text
cmd/server
    │
    ├── load config
    ├── connect PostgreSQL
    ├── run migrations
    ├── build AppState
    ├── build Fiber app
    ├── start listener
    ├── wait for shutdown signal
    └── gracefully stop
            ↓
          app
            │
            ├── middleware
            ├── foundation routes
            └── future /api/v1 feature routes
                    ↓
                AppState
                    ↓
               pgxpool.Pool
                    ↓
                PostgreSQL
```

Persistent data must never be held in custom in-memory application state.

---

# 7. AppState

Create one small explicit application state structure.

Required concept:

```go
type AppState struct {
    DB *pgxpool.Pool
}
```

The exact file location may be:

```text
backend/app/state.go
```

or another clear location inside the app foundation.

Rules:

- `DB` is the existing shared `*pgxpool.Pool`.
- Do not wrap the pool in a mutex.
- Do not create a service container.
- Do not create a map of dependencies.
- Do not use reflection.
- Do not add future services to AppState before they exist.
- Do not store request-specific state in AppState.

Future specs may add concrete shared dependencies only when they actually exist.

---

# 8. Fiber Application Construction

Replace the current featureless constructor with an explicit constructor that can receive the application state and the HTTP configuration it actually needs.

A simple shape is preferred, for example:

```go
func New(state AppState, settings config.Config) *fiber.App
```

The exact signature may differ slightly if a simpler tested shape is found, but do not introduce factories or dependency injection abstractions.

`New(...)` must:

```text
create Fiber app
configure central error handling
register foundation middleware
register foundation routes
prepare /api/v1 routing for later specs
return the app
```

`New(...)` must NOT:

```text
open database connections
run migrations
load .env
start listening
spawn background workers
seed data
```

Those responsibilities belong outside app construction.

The application should remain easy to instantiate in tests without opening a network listener.

---

# 9. API Namespace

All normal application APIs added by future specs must live under:

```text
/api/v1
```

Examples for later specs:

```text
/api/v1/auth/login
/api/v1/auth/logout
/api/v1/auth/me
/api/v1/tickets
```

SPEC-02 should establish a clear routing pattern for this namespace.

Do not add fake placeholder endpoints merely to prove the group exists.

Infrastructure health endpoints may remain outside `/api/v1`.

---

# 10. Central Error Model

Use one explicit HTTP application error representation.

Preferred concept:

```go
type AppError struct {
    Code       string
    Message    string
    HTTPStatus int
}
```

A pointer receiver or constructor helpers may be used if they materially improve clarity.

Do not create a generic error framework.

Do not create a hierarchy of custom error interfaces.

Do not expose internal database/system errors directly to clients.

The stable external error response shape must be:

```json
{
  "error": {
    "code": "example_error",
    "message": "Human-readable safe message",
    "request_id": "..."
  }
}
```

`request_id` must be included when one exists.

The error response must use JSON content type.

---

# 11. Central Fiber Error Handler

Configure one Fiber error handler for the application.

It must distinguish at least:

```text
known AppError
Fiber HTTP error
unknown internal error
```

Expected behavior:

## Known application error

Return its intended safe status/code/message.

## Route not found

Return:

```http
404
```

with a stable JSON error such as:

```json
{
  "error": {
    "code": "not_found",
    "message": "Resource not found",
    "request_id": "..."
  }
}
```

Do not expose framework internals.

## Method not allowed

If Fiber produces a method-not-allowed error, return:

```http
405
```

with a stable safe JSON error.

## Unknown internal error

Return:

```http
500
```

with:

```json
{
  "error": {
    "code": "internal_server_error",
    "message": "Internal server error",
    "request_id": "..."
  }
}
```

The original internal error should be logged server-side with request context.

Never send stack traces, SQL details, credentials, or raw database errors to the client.

---

# 12. Request ID

Every HTTP request must have a request ID.

Requirements:

```text
each response exposes X-Request-ID
request ID is available to logging
request ID is available to the central error response
```

Use Fiber's maintained request-ID middleware if it satisfies the requirements without unnecessary behavior.

Do not add a new UUID dependency solely for request IDs if Fiber/current dependencies or a small standard-library implementation already solve the problem clearly.

The exact request-ID format is not part of the public API contract.

Tests must not depend on a specific UUID/random format.

---

# 13. Request Logging

Use the standard Go `log` package for SPEC-02.

For each completed HTTP request, log useful operational information such as:

```text
request_id
method
path
status
latency
```

Example conceptual log:

```text
request_id=... method=GET path=/health status=200 latency=...
```

Rules:

- Do not log request bodies.
- Do not log passwords.
- Do not log cookies.
- Do not log raw session tokens.
- Do not log Authorization headers.
- Do not dump every request header.
- Do not introduce structured logging dependencies in SPEC-02.

Future specs may add useful domain context such as ticket/user IDs where appropriate.

---

# 14. Panic Recovery

A panic in one request must not crash the entire HTTP process.

Use Fiber's maintained recovery middleware or a similarly small explicit solution.

Recovered failures must flow into the central error handling path and return a safe `500` response.

Do not expose panic stack traces to clients.

Server logs may contain diagnostic information appropriate for development and operations, but never secrets.

---

# 15. CORS

CORS must be explicit because future authentication uses browser cookies.

Add configuration for the frontend origin.

Required environment variable:

```text
FRONTEND_ORIGIN
```

Development behavior:

```text
default: http://localhost:5173
```

Production/non-development behavior:

```text
FRONTEND_ORIGIN must be explicitly configured
```

Reject invalid values rather than silently allowing every origin.

Rules:

```text
do not use Allow-Origin: *
allow the configured frontend origin only
credentials must be supported
```

This prepares the backend for future HttpOnly session cookies without implementing authentication in SPEC-02.

Allow only the ordinary HTTP methods/headers required by the current/future web API.

Do not add arbitrary permissive CORS configuration.

CORS configuration must be tested.

---

# 16. Server Bind Configuration

Remove the hard-coded listen address from `cmd/server/main.go`.

Add configuration:

```text
BACKEND_BIND_ADDRESS
```

Development default:

```text
127.0.0.1:3000
```

The server must listen on the configured value.

Examples that later deployments may use:

```text
127.0.0.1:3000
0.0.0.0:3000
:3000
```

Do not hard-code a production hostname or public domain into Go source.

Validation should reject an obviously unusable empty/invalid bind address.

Do not require the application to bind publicly during development.

---

# 17. Graceful Shutdown Configuration

Add one bounded shutdown duration.

Recommended environment variable:

```text
BACKEND_SHUTDOWN_TIMEOUT_SECONDS
```

Recommended default:

```text
10
```

Recommended safety ceiling:

```text
60
```

Reject:

```text
zero
negative values
values above the safety ceiling
```

Do not add many server timing knobs in SPEC-02.

Only add configuration that this SPEC actually consumes.

---

# 18. Liveness Endpoint

Keep a lightweight liveness endpoint:

```http
GET /health
```

Purpose:

> Is the HTTP process alive and able to respond?

It must NOT query PostgreSQL.

Return:

```http
200 OK
Content-Type: application/json
```

Body:

```json
{
  "status": "ok"
}
```

This intentionally replaces the current plain-text `"ok"` foundation response.

The liveness route must remain cheap and independent of database availability.

---

# 19. Readiness Endpoint

Add:

```http
GET /ready
```

Purpose:

> Is this backend ready to serve requests that depend on PostgreSQL?

Readiness must check the existing `pgxpool.Pool` using a bounded `context.Context`.

A simple database ping or `SELECT 1` is sufficient.

Do not run application queries.

Do not inspect every table.

Do not run migrations inside readiness.

## Ready

When PostgreSQL is reachable:

```http
200 OK
```

Body:

```json
{
  "status": "ready"
}
```

## Not ready

When PostgreSQL is unavailable or the readiness check times out:

```http
503 Service Unavailable
```

Return a safe response such as:

```json
{
  "error": {
    "code": "not_ready",
    "message": "Service not ready",
    "request_id": "..."
  }
}
```

Do not expose the database hostname, credentials, SQL error, or connection string to the client.

The underlying failure should be logged.

Use an existing bounded database timeout where practical rather than inventing another large configuration surface.

---

# 20. Startup Lifecycle

`cmd/server/main.go` must remain the process entry point.

Startup order:

```text
1. load local .env when available
2. load and validate configuration
3. connect pgxpool
4. verify PostgreSQL connectivity
5. run pending migrations
6. construct AppState
7. construct Fiber app
8. start HTTP listener
9. wait for termination or listener failure
```

If any required startup step fails:

```text
log the failure
exit non-zero
```

Do not start a partially initialized HTTP application.

Development seed data must NOT be run automatically by normal server startup.

The existing explicit command remains responsible for development seed:

```text
go run ./cmd/seed_development
```

---

# 21. Graceful Shutdown

Handle normal process termination signals appropriate for deployment, at least:

```text
SIGINT
SIGTERM
```

Desired shutdown sequence:

```text
termination signal
      ↓
stop accepting new requests
      ↓
allow in-flight HTTP work a bounded time to finish
      ↓
Fiber shuts down
      ↓
PostgreSQL pool closes
      ↓
process exits
```

The shutdown timeout must be bounded by `BACKEND_SHUTDOWN_TIMEOUT_SECONDS`.

If the timeout expires, the process must not wait forever.

Do not create a complicated lifecycle framework.

A small explicit `signal.NotifyContext` / channel-based implementation is preferred.

Do not add background worker shutdown infrastructure when no workers exist.

---

# 22. Listener Failure

A listener failure that is not the expected result of graceful shutdown must be treated as a startup/runtime error.

Examples:

```text
port already in use
invalid bind address
listener unexpectedly stops
```

The process must not silently continue after the HTTP server has stopped.

Tests should cover configuration validation; network-level listener behavior may be tested only where it remains deterministic and useful.

---

# 23. Middleware Order

Middleware order must be intentional.

The effective behavior should ensure:

```text
request ID exists before request logging/error output needs it
panic recovery protects downstream handlers
CORS headers apply correctly
request logging observes final status and latency
central error handling produces the final safe response
```

Do not stack duplicate middleware.

Document a non-obvious order only if future maintainers would otherwise be likely to break it.

---

# 24. No Global Request Timeout Middleware Yet

Do not add a global HTTP timeout middleware in SPEC-02 merely because timeouts sound production-ready.

Future realtime/WebSocket endpoints are long-lived and should not accidentally inherit a short global request timeout.

Use bounded `context.Context` values around operations such as PostgreSQL readiness checks and later database work.

If a later SPEC requires route-specific timeouts, add them deliberately there.

---

# 25. Configuration Changes

Extend the existing `config.Config` rather than introducing a second configuration system.

Expected new foundation fields conceptually include:

```go
BackendBindAddress            string
FrontendOrigin                string
BackendShutdownTimeoutSeconds int
```

Keep existing SPEC-01 database and seed configuration intact.

Do not rename existing environment variables without a real need.

Current relevant configuration includes:

```text
DATABASE_URL
DATABASE_MAX_CONNECTIONS
DATABASE_MIN_CONNECTIONS
DATABASE_ACQUIRE_TIMEOUT_SECONDS
APP_ENV
SEED_DEVELOPMENT_DATA
SEED_ADMIN_*
```

SPEC-02 adds only the HTTP/server configuration it consumes.

---

# 26. APP_ENV Behavior

Keep `APP_ENV` simple.

At minimum distinguish:

```text
development
non-development
```

Do not introduce:

```text
APP_ENV=test
TEST_ENV
```

as project environments only for automated testing.

For CORS:

```text
development:
FRONTEND_ORIGIN may default to http://localhost:5173

non-development:
FRONTEND_ORIGIN must be explicitly present and valid
```

Do not make production secretly fall back to localhost.

## Permanent local/test environment rule

SPEC-02 must follow the permanent project rule from `PROJECT_CONTEXT.md`.

The only local environment file is:

```text
backend/.env
```

Never generate:

```text
.env.example
.env.test
.env.testing
.env.local
.env.development
```

Never introduce test-only database variables such as:

```text
DATABASE_TEST_URL
TEST_DATABASE_URL
TEST_DB_URL
```

Database-backed tests must use:

```text
DATABASE_URL
```

from the real local environment and isolate their database work with temporary
PostgreSQL schemas.

Any legacy SPEC-01 test helper that still reads `DATABASE_TEST_URL` must be
updated during SPEC-02 to use `DATABASE_URL` while preserving the existing
temporary-schema/search-path isolation behavior.

Do not weaken, skip, or convert those PostgreSQL tests into mocks.

Config unit tests may use `t.Setenv` only to temporarily exercise official
runtime variables already defined by the application.

A legitimate new runtime setting required by SPEC-02, such as
`FRONTEND_ORIGIN`, `BACKEND_BIND_ADDRESS`, or
`BACKEND_SHUTDOWN_TIMEOUT_SECONDS`, is not a test-only variable. It belongs to
the real runtime configuration and may be added to the ignored local
`backend/.env` when needed.

Do not create another env file for it.

---

# 27. Frontend Origin Validation

`FRONTEND_ORIGIN` must represent one concrete HTTP(S) origin.

Accept conceptual forms such as:

```text
http://localhost:5173
https://staff.example.com
```

Reject unsafe/ambiguous values such as:

```text
*
empty non-development value
origin containing path/query/fragment
unsupported scheme
```

The exact validation helper should remain short and testable.

Do not implement a list of many production origins unless a later requirement needs multiple frontends.

---

# 28. App Package Structure

Keep the package small.

A reasonable result may look like:

```text
backend/app/
├── app.go
├── state.go
├── errors.go
├── middleware.go
├── health.go
└── app_test.go
```

This is guidance, not a requirement to create every listed file.

Prefer fewer clear files over empty/one-function fragmentation.

Do not create:

```text
controllers/
usecases/
adapters/
ports/
providers/
containers/
factories/
```

for this foundation.

---

# 29. Shared Package Boundary

`backend/shared/database.go` continues to own:

```text
pgxpool connection creation
migration runner
migration parsing/history
```

Do not move migration logic into the HTTP app package.

Do not turn `shared` into a dumping-ground package.

Only add a shared helper when more than one real package needs it.

---

# 30. No Database Schema Changes

SPEC-01 owns the database foundation.

SPEC-02 must not add a new database migration unless a genuine schema defect is discovered that blocks the required backend foundation.

The following must remain intact:

```text
migration versions 1..17, 19, 20
version 18 retired
multi-department ticket assignment
multi-user ticket assignment
ticket priority boolean
optional ticket due_at
ticket_attachments
```

Do not alter ticket schema for an HTTP-foundation task.

If a genuine SPEC-01 blocker is discovered, stop and report it instead of silently changing product data modeling.

---

# 31. No Development Seed Changes

The existing development seed workflow is outside normal server startup.

Do not add seed data for SPEC-02.

Do not add fake users or tickets.

Do not make `cmd/server` call:

```go
SeedDevelopmentFixtures(...)
SeedDevelopmentAdmin(...)
```

Development seed remains explicitly invoked by:

```bash
go run ./cmd/seed_development
```

---

# 32. Test Strategy

Use focused tests for foundation behavior.

Do not chase artificial coverage percentages.

Required categories:

```text
configuration tests
app construction tests
health tests
readiness tests
request ID tests
CORS tests
error handling tests
panic recovery test
shutdown/lifecycle tests where deterministic
existing database foundation regression tests
legacy SPEC-01 test-environment cleanup
```

As part of SPEC-02, update legacy PostgreSQL test helpers that still use
`DATABASE_TEST_URL` so they use the official `DATABASE_URL` while preserving
their isolated temporary-schema behavior.

Tests must not require production data.

---

# 33. Config Tests

Add or update tests for:

```text
BACKEND_BIND_ADDRESS default
BACKEND_BIND_ADDRESS override
invalid bind address rejection where applicable

FRONTEND_ORIGIN development default
FRONTEND_ORIGIN explicit override
FRONTEND_ORIGIN required outside development
wildcard origin rejection
invalid scheme rejection
path/query/fragment rejection

BACKEND_SHUTDOWN_TIMEOUT_SECONDS default
valid override
zero rejection
negative rejection
upper-bound rejection
```

Existing database/TLS/pool/seed config tests must continue to pass.

Avoid tests that depend on the developer's machine environment accidentally.

---

# 34. Health Tests

Verify:

```http
GET /health
```

returns:

```text
status 200
JSON content type
{"status":"ok"}
X-Request-ID present
```

Health must succeed even when the database state supplied to the application is not ready, because liveness must not depend on PostgreSQL.

Do not keep a test expecting the old plain-text `"ok"` response.

---

# 35. Readiness Tests

Use the real PostgreSQL instance configured by `DATABASE_URL`, with an isolated temporary PostgreSQL schema, where database behavior is involved.

Verify:

```text
healthy pool -> GET /ready -> 200 + {"status":"ready"}

unavailable/closed/unusable pool or controlled readiness failure
-> GET /ready -> 503 safe JSON error
```

The client response must not reveal the underlying database error.

Do not replace the PostgreSQL readiness behavior with a fake repository.

A very small test seam is acceptable only if a deterministic error-path test cannot otherwise be achieved safely; do not build an abstraction layer solely for one test.

---

# 36. Request ID Tests

Verify:

```text
ordinary successful response has X-Request-ID
error response contains a request ID
logged/error correlation can use the same request ID
```

Do not assert a specific random value format unless the implementation contract explicitly defines one.

Do not add a heavyweight ID library only for the test.

---

# 37. CORS Tests

Verify at minimum:

Configured origin:

```text
Origin: configured FRONTEND_ORIGIN
```

receives the expected CORS allow-origin response.

An unrelated origin must not receive permissive access.

Credentialed browser requests must be compatible with future HttpOnly session cookies.

Verify the backend never responds with credentialed:

```text
Access-Control-Allow-Origin: *
```

---

# 38. Error Handling Tests

Test at least:

```text
unknown route -> 404 stable JSON error
method not allowed -> 405 stable JSON error where supported
known AppError -> intended status/code/message
unknown handler error -> 500 safe generic JSON
```

Also verify:

```text
internal error text is not leaked
request_id is available in error response
```

Do not create feature endpoints solely to trigger errors in production routing; test-only routes may be registered in tests when appropriate.

---

# 39. Panic Recovery Test

Register a test-only route that panics.

Verify:

```text
server process/test app remains alive
response is safe 500
no panic detail is exposed in response body
```

Do not add a permanent `/panic` route.

---

# 40. Graceful Shutdown Tests

Test shutdown code at the smallest deterministic boundary possible.

At minimum, structure lifecycle code so shutdown behavior is not trapped inside one untestable giant `main()` function.

Do not create a large runtime abstraction.

A small function such as:

```go
run(...)
```

or similarly explicit lifecycle helper is acceptable if it materially improves testability.

Avoid flaky tests based on arbitrary sleeps.

Use contexts/channels/signals or local ephemeral listeners where needed.

---

# 41. Existing Regression Tests

All completed SPEC-01 tests must continue to pass.

In particular, SPEC-02 must not regress:

```text
fresh migrations
migration checksum validation
migration advisory lock behavior
version 18 retirement
migration 20
legacy assignment backfill
ticket constraints
multi-assignment
ticket attachments
development seed isolation
password hashing
database TLS validation
pool safety bounds
```

Do not weaken old tests to make SPEC-02 pass.

---

# 42. Performance Rules

SPEC-02 is infrastructure, not a throughput benchmark exercise.

The foundation must avoid obvious hot-path waste:

```text
no database query in /health
no unnecessary DB query middleware on every request
no per-request creation of expensive global objects
no giant response envelopes
no reflection-heavy DI
no synchronous filesystem reads per request
no unbounded goroutine creation
```

Request logging and request-ID generation should remain lightweight.

Do not add caching.

Do not add Redis.

Do not claim performance improvements without measurement.

---

# 43. Security Rules

SPEC-02 must preserve these security properties:

```text
no secrets in logs
no raw database errors in HTTP responses
explicit CORS origin
no wildcard credentialed CORS
bounded shutdown/readiness contexts
production configuration does not silently use localhost frontend origin
database TLS rules from SPEC-01 remain intact
```

Do not implement authentication early.

Do not add a temporary hard-coded API key.

Do not add fake authorization.

---

# 44. Logging Startup Information

Startup logs may include non-sensitive operational information such as:

```text
application environment
bind address
server started
migration completion
shutdown initiated
shutdown completed
```

Do not log:

```text
DATABASE_URL
database password
seed admin password
future session secret
cookies
authorization values
```

When logging the bind address, do not print unrelated configuration secrets.

---

# 45. HTTP Response Discipline

Foundation responses should be explicit and small.

Success examples:

```json
{"status":"ok"}
```

```json
{"status":"ready"}
```

Error example:

```json
{
  "error": {
    "code": "not_ready",
    "message": "Service not ready",
    "request_id": "..."
  }
}
```

Do not wrap every success response in:

```json
{
  "success": true,
  "data": {}
}
```

unless a later project-wide decision explicitly adopts such an envelope.

---

# 46. Context Propagation

Use standard `context.Context` for bounded application/database work.

Do not retain Fiber request-backed data after the handler returns.

When future code needs data beyond request lifetime, it must copy needed values and use an appropriate standard Go context.

SPEC-02 itself should use bounded contexts for readiness/database lifecycle operations where relevant.

Do not create custom context types.

---

# 47. Dependency Discipline

Prefer existing dependencies and standard library.

Fiber middleware packages that are part of the existing Fiber module are acceptable when they directly solve:

```text
CORS
request ID
panic recovery
```

Do not add dependencies for:

```text
logging framework
DI
UUIDs
config framework
error framework
lifecycle framework
```

unless the existing stack cannot satisfy a concrete requirement cleanly.

If any new dependency is added, the implementation report must state:

```text
dependency
purpose
why existing stack was insufficient
scope where used
```

---

# 48. Documentation

Update documentation only where the actual backend foundation changes current repository behavior.

At minimum, if applicable, keep these accurate:

```text
README.md
Context-Spec-BWP-SonaSea/PROJECT_CONTEXT.md current repository snapshot
```

Do not rewrite product requirements.

Do not change deferred Report/Admin rules.

Do not create large architecture documents for this small foundation.

This SPEC itself should live at:

```text
Context-Spec-BWP-SonaSea/Spec/SPEC-02-backend-foundation.md
```

---

# 49. Recommended Implementation Order

Use this order unless repository inspection reveals a concrete dependency:

```text
1. Read PROJECT_CONTEXT.md
2. Read SPEC-02
3. Inspect current app/config/server/database code and tests
4. Extend config for bind/origin/shutdown settings
5. Add AppState
6. Add central error model/handler
7. Add request ID
8. Add recovery
9. Add request logging
10. Add CORS
11. Replace /health with JSON liveness behavior
12. Add /ready PostgreSQL readiness
13. Establish /api/v1 routing pattern
14. Refactor server startup only as needed
15. Add graceful shutdown
16. Add focused tests
17. Run all existing backend verification
18. Review diff for scope creep
```

Do not start SPEC-03 work.

---

# 50. Verification Commands

Before SPEC-02 is considered complete, run from `backend/`:

```bash
gofmt -d .
go mod tidy
go vet ./...
go test ./... -count=1
go build ./...
```

Where the local toolchain supports it, also run:

```bash
go test -race ./... -count=1
```

Database-backed tests must use the real PostgreSQL instance configured by `DATABASE_URL` and isolate themselves with temporary PostgreSQL schemas.

Also verify manually or through tests:

```text
GET /health
GET /ready
404 error
CORS allowed origin
CORS unrelated origin
request ID response header
graceful shutdown
```

Do not claim a check passed unless it was actually executed.

---

# 51. Manual Runtime Verification

With a valid local PostgreSQL development environment:

Start the backend:

```bash
go run ./cmd/server
```

Verify:

```text
server binds to configured BACKEND_BIND_ADDRESS
migrations complete
development fixtures are NOT automatically seeded
```

Then verify:

```http
GET /health
```

returns JSON `200`.

Verify:

```http
GET /ready
```

returns JSON `200` while PostgreSQL is available.

Verify a nonexistent route returns the central JSON `404`.

Terminate the process normally and verify graceful shutdown occurs rather than an abrupt uncontrolled exit.

Do not use the production database for manual verification.

---

# 52. Definition of Done

SPEC-02 is complete only when all of the following are true:

- Existing SPEC-01 behavior still passes.
- `AppState` exists and contains the PostgreSQL pool without custom locking.
- Fiber app construction accepts the dependencies/configuration it actually needs.
- Fiber app construction does not connect to PostgreSQL or start the listener.
- Normal application APIs have a clear `/api/v1` routing foundation.
- No fake feature endpoint exists.
- Central safe JSON error handling exists.
- Unknown internal errors return safe `500` responses.
- 404 responses use the central error format.
- Request IDs exist on responses and error bodies.
- Request logging includes request ID/method/path/status/latency.
- Request logging does not expose secrets.
- Panic recovery prevents a request panic from crashing the service.
- CORS allows only the configured frontend origin.
- Credentialed CORS never uses wildcard origin.
- `/health` is a DB-independent JSON liveness endpoint.
- `/ready` checks PostgreSQL with a bounded context.
- `/ready` returns safe `503` when PostgreSQL is not ready.
- Server bind address is configurable.
- Production/non-development frontend origin must be explicit.
- Graceful shutdown handles SIGINT/SIGTERM.
- Shutdown wait is bounded.
- PostgreSQL pool closes during process shutdown.
- Normal server startup does not seed development data.
- No schema migration was added unless an actual blocking SPEC-01 defect was surfaced and explicitly justified.
- No authentication behavior was implemented.
- No ticket behavior was implemented.
- No frontend work was implemented.
- No Redis/queues/microservices/ORM/DI framework was introduced.
- `gofmt -d .` passes.
- `go mod tidy` passes.
- `go vet ./...` passes.
- `go test ./... -count=1` passes.
- `go build ./...` passes.
- Race tests pass where the environment supports them.
- Documentation reflects the actual resulting foundation.
- The final diff remains inside SPEC-02 scope.

---

# 53. Handoff to SPEC-03

After SPEC-02 is complete, the backend should be ready for:

```text
SPEC-03 — Authentication Backend
```

SPEC-03 will implement real authentication using the existing PostgreSQL schema and the reusable HTTP foundation from this SPEC.

Expected future authentication endpoints:

```text
POST /api/v1/auth/login
POST /api/v1/auth/logout
GET  /api/v1/auth/me
```

SPEC-03 will own:

```text
username/password validation
Argon2id verification
server-side session creation
session token hashing
HttpOnly auth cookie
session lookup
logout/revocation
authenticated user context
authorization foundation where required
```

Do not implement those items in SPEC-02.

---

# 54. Final Implementation Rule

The desired result is not a generic enterprise backend framework.

The desired result is this:

```text
Config
  ↓
PostgreSQL pool
  ↓
Migrations
  ↓
AppState
  ↓
Fiber
  ├── Request ID
  ├── Recovery
  ├── Logging
  ├── CORS
  ├── Central errors
  ├── /health
  ├── /ready
  └── /api/v1
  ↓
Graceful shutdown
```

Keep it boring.

Keep it explicit.

Keep it small.

Do not build future features early.

Do not sacrifice correctness or security for fewer lines.

Do not add abstraction simply because later specs may exist.
