# BWP SonaSea — Project Context

> **Purpose of this file**
>
> This is the permanent project context for **BWP SonaSea**.
>
> Every AI coding agent must read this file **before reading or implementing any SPEC**.
>
> This document defines the project's locked architecture, engineering philosophy, folder structure, coding rules, performance expectations, UI behavior, data rules, and implementation workflow.
>
> Individual SPEC files define the scope of a particular implementation phase. A SPEC may add details, but it must not silently replace the locked decisions in this file.
>
> If a SPEC appears to conflict with this file, stop and identify the conflict instead of inventing a new architecture.

## Current repository snapshot

As of 2026-09-10, the repository contains the SPEC-01 PostgreSQL foundation and
the initial Go/Fiber v3 backend bootstrap. Feature packages are added only when
their corresponding SPEC implements behavior; empty future folders are not
generated.

```text
BWP-SonaSea/
├── backend/
│   ├── go.mod
│   ├── go.sum
│   ├── migrations/
│   ├── cmd/
│   │   ├── server/main.go
│   │   └── seed_development/main.go
│   ├── app/app.go
│   ├── app/app_test.go
│   ├── config/config.go
│   ├── config/config_test.go
│   ├── shared/
│   │   ├── database.go
│   │   ├── migrations_test.go
│   │   ├── foundation_test.go
│   │   └── security/
│   │       ├── password.go
│   │       └── password_test.go
│   └── admin/
│       ├── seed.go
│       └── fixtures.go
├── frontend/
│   ├── package.json
│   ├── bun.lock
│   ├── vite.config.ts
│   ├── tsconfig.json
│   ├── .gitignore
│   ├── static/
│   └── src/
│       ├── app.d.ts
│       ├── app.html
│       ├── lib/
│       │   ├── assets/
│       │   └── index.ts
│       └── routes/
│           ├── +layout.svelte
│           └── +page.svelte
├── database/
│   ├── README.md
│   ├── bwp-sonasea.dump
│   ├── full_app_schema.sql
│   └── scripts/
│       ├── README.md
│       ├── backup_database.ps1
│       ├── verify_foundation.sql
│       └── performance_baseline.sql
├── docker/
│   └── README.md
├── img/
├── docs/baron/
├── .agents/
├── .baron/
├── .codex/
├── Context-Spec-BWP-SonaSea/
│   ├── PROJECT_CONTEXT.md
│   └── Spec/
│       └── SPEC-01-database-foundation.md
├── .gitignore
├── AGENTS.md
└── README.md
```

`backend/.env` is local-only and ignored by Git. `frontend/node_modules/` and
Go build/test output under `backend/bin/`, `backend/*.exe`, and
`backend/*.test` are also ignored. The legacy `target/` ignore remains as a
guard against accidentally committing legacy or other build output; the Go
backend does not generate it.
The local PostgreSQL backup is kept at `database/bwp-sonasea.dump` and is
ignored by Git because it contains data. Backend and frontend tests stay with
the project that owns them; there is no root `test/` folder. The `backend`
feature folders follow the locked layout below, while behavior is added only
by the relevant SPEC.

---

# 1. Project Name

**BWP SonaSea**

BWP SonaSea is a hotel operations and internal support platform inspired by the existing SARA-style hotel ticketing workflow, but rebuilt with a modern architecture focused on:

- Fast perceived UX
- Fast backend response
- Realtime synchronization
- Maintainable code
- Clear architecture
- Low operational complexity
- Long-term stability
- Responsive desktop/mobile experience
- Production use through a real public domain

This is not a toy project and must not be implemented as a demo application.

---

# 2. Product Goal

The application is intended to support real hotel operations.

Core product areas include:

```text
Authentication

Tickets
├── Open
│   ├── Pending
│   └── Accepted
└── Closed

Ticket Detail
├── Accept
├── Assign
├── Close
├── Chat
├── Attachments
└── Checklist

Staff Meal

Announcements

Report

Settings

Admin functionality
```

The application should feel closer to a responsive desktop/mobile application than a traditional page-refresh-heavy website.

---

# 3. Locked Technology Stack

These technologies are **already decided**.

Do not propose replacing them unless the user explicitly asks to reconsider the stack.

## Frontend

```text
SvelteKit
Svelte
TypeScript
Bun for package management and script execution
```

## Backend

```text
Go (current supported version)
Fiber v3
standard context.Context
Fiber middleware only when required
standard log for the foundation
```

## Database

```text
PostgreSQL
pgx v5
pgxpool
parameterized SQL
```

## Realtime

```text
Fiber-compatible WebSocket in the realtime SPEC
```

Use WebSocket only where realtime behavior is actually required.

## Files and images

Production binary files should eventually be stored using:

```text
Object Storage
+
CDN where appropriate
```

PostgreSQL stores metadata and storage references, not large binary images/files.

## Optional infrastructure

The following are **not default requirements**:

```text
Redis
Docker
Caddy
Background queues
Object storage provider
Prometheus
Grafana
Kubernetes
Microservices
```

Introduce them only when a concrete requirement justifies them.

---

# 4. Architecture Philosophy

The project follows this principle:

> **Boring, explicit code by default. Optimize based on measurement.**

"Boring code" does **not** mean careless, naive, or slow code.

It means:

- Easy to read
- Easy to trace
- Easy to debug
- Minimal hidden behavior
- Minimal abstraction
- Minimal magic
- Predictable control flow
- Clear responsibility per module
- Performance-conscious without premature complexity

The goal is production-quality software that can still be understood several years later.

---

# 5. Performance Philosophy

The project is expected to be fast, but performance must come from correct architecture rather than complicated Go or framework syntax.

Do not assume:

```text
more generics = faster
more traits = faster
Arc<Mutex<...>> = faster
more abstraction = more professional
more crates = better architecture
```

Performance should primarily come from:

```text
good PostgreSQL schema
correct indexes
efficient SQL
avoiding N+1 queries
bounded result sets
pagination
connection pooling
efficient serialization
small payloads
realtime push where appropriate
optimistic frontend updates
good asset delivery
correct caching when actually needed
avoiding lock contention
avoiding unnecessary clones/allocations
measuring before optimizing
```

---

# 6. Performance Priorities

When a feature feels slow, investigate in this order where applicable:

```text
1. Network latency
2. Frontend data flow / unnecessary refetch
3. SQL query design
4. Database indexes
5. N+1 query behavior
6. Payload size
7. Image/file handling
8. Serialization
9. Connection pool behavior
10. Lock contention
11. Actual CPU bottleneck
```

Do not immediately rewrite code into advanced Go or framework-specific tricks.

Use evidence.

For PostgreSQL performance work, prefer:

```text
EXPLAIN
EXPLAIN ANALYZE
query timing
index inspection
realistic load tests
```

For Go/backend performance work, use profiling and database measurements before
changing architecture.

---

# 6A. Extreme Database Performance Bar

Database performance is a first-class product requirement, not a cleanup task
for later. The database foundation must make the fast path obvious and keep
future Fiber + pgx/pgxpool handlers fast under real data volume.

Required database performance discipline:

```text
design indexes from real WHERE, JOIN, and ORDER BY patterns
use composite indexes in predicate/order-column order
make list ordering deterministic with a stable tie-breaker such as id
use partial indexes for hot filtered subsets such as unread/active rows
avoid redundant or unused indexes because every index costs write time and storage
use bounded result sets and stable pagination for every growing list
use explicit column lists instead of SELECT * in application queries
avoid N+1 queries and unnecessary round trips
keep transactions short and lock scope narrow
bound pool size and acquire time instead of opening unbounded connections
measure representative queries with EXPLAIN (ANALYZE, BUFFERS)
verify performance assumptions with realistic data before claiming them
```

Do not trade data correctness or recoverability for a benchmark number. Do not
add Redis, read replicas, sharding, partitioning, materialized views, or other
distributed complexity without measured evidence and an explicit later
architecture decision.

---

# 7. Go Coding Style

Go code must remain intentionally simple.

The target style is:

```text
explicit
small
predictable
boring
typed
safe
fast
```

A developer who understands basic Go should be able to follow most application code.

---

# 8. No Over-Engineering Rule

Do not introduce advanced Go or framework constructs merely because they are
idiomatic in a large codebase.

Avoid unless there is a concrete requirement:

```text
complex generic hierarchies
generic repositories
generic services
trait factories
deep trait-object architecture
reflection-heavy dependency injection
code generation without a concrete need
unsafe code
deep middleware abstractions
global mutable state
```

These constructs are not forbidden when technically necessary.

If one is introduced, the implementation must have a clear reason.

Example of an acceptable reason:

```text
A genuinely shared mutable in-memory structure must be accessed safely
from multiple concurrent tasks.
```

Example of an unacceptable reason:

```text
"This is how advanced Go projects usually look."
```

---

# 9. Shared State and Concurrency

Do not add synchronization or background workers by default.

Do not imitate a database with in-memory state such as:

```go
var tickets []Ticket
```

to imitate a database.

The project uses PostgreSQL.

Use PostgreSQL and `pgxpool.Pool` for persistent application state. A pool is
designed for concurrent use and must not be wrapped in a custom global lock.

When a handler needs application state, keep it explicit and small:

```go
type AppState struct {
    DB *pgxpool.Pool
}
```

Only add shared mutable memory when a concrete feature requires it. Do not
retain Fiber request-backed values after a handler returns; copy data before
passing it to longer-lived work and use `context.Context` for that work.

---

# 10. Interfaces

Interfaces are useful, but they must solve a real problem.

Do not automatically create:

```text
TicketRepository interface
PostgresTicketRepository
TicketRepositoryFactory
GenericTicketService[T]
```

for ordinary SQL operations.

Prefer a direct function when sufficient:

```go
func FindByID(ctx context.Context, db *pgxpool.Pool, id int64) (Ticket, error) {
    // SQL
}
```

An interface may be introduced later when there is a real requirement such as:

```text
multiple implementations
test boundary that materially benefits from an interface
runtime polymorphism
shared behavior across meaningful implementations
```

Do not use interfaces only to make the project look more enterprise.

---

# 11. Framework and Code Generation

Use normal Go code and Fiber handlers. Do not introduce reflection-heavy
frameworks, code generation, or hidden binding behavior for core business
logic. SQL remains visible and is executed with parameterized `pgx` queries.

---

# 12. Fiber Style

Fiber should remain thin.

Handlers should mainly:

```text
extract request data
validate HTTP-level input
read authenticated context
call service
return response
```

Preferred handler shape:

```go
func GetTicket(c fiber.Ctx) error {
	id := c.Params("id")
	ticket, err := ticketService.GetByID(c.Context(), db, id)
	if err != nil {
		return err
	}
	return c.JSON(ticket)
}
```

Do not put large SQL queries directly inside handlers.

Do not put unrelated business workflows directly inside handlers.

---

# 13. Backend Layering

Default backend flow:

```text
HTTP Request
     ↓
Handler
     ↓
Service
     ↓
Repository
    ↓
pgx/pgxpool
     ↓
PostgreSQL
```

Responsibilities:

## Handler

```text
HTTP extraction
HTTP validation
authentication context
HTTP response mapping
```

## Service

```text
business rules
business workflow
authorization decisions where appropriate
coordination between repositories/services
```

## Repository

```text
SQL
database reads
database writes
transactions
database-specific behavior
```

## Model

```text
domain structs
request DTOs
response DTOs
enums
database row models where appropriate
```

Do not force every trivial operation to have unnecessary wrapper layers.

For very small features, a service function can remain short.

The point is separation of responsibility, not artificial line count.

---

# 14. Backend Folder Structure

The current backend contains only the SPEC-01 foundation. The layout is
deliberately small and follows Go package boundaries. Client feature packages
and shared helpers are created only when a SPEC needs real code; empty future
folders and placeholder files are not generated.

```text
backend/
├── go.mod
├── go.sum
├── migrations/
├── cmd/
│   ├── server/main.go
│   └── seed_development/main.go
├── app/
│   ├── app.go
│   └── app_test.go
├── config/
│   ├── config.go
│   └── config_test.go
├── shared/
│   ├── database.go
│   ├── foundation_test.go
│   ├── migrations_test.go
│   └── security/
│       ├── password.go
│       └── password_test.go
└── admin/
    ├── seed.go
    └── fixtures.go
```

When later client behavior is implemented, use these high-level packages:

```text
backend/client/
├── auth/
├── tickets/
├── chat/
├── checklist/
├── announcements/
├── staff_meal/
├── reports/
└── settings/
```

The current `backend/shared/` package owns the database pool and migration
runner because those are shared infrastructure. Add other shared helpers only
when multiple implemented packages need a concrete small utility. The
existing local `backend/.env` is the only development environment file. Never
generate or commit `.env.example` or another example env file.

The normal server migration path discovers schema migrations only from
`backend/migrations/`. Development fixtures are not migrations. The historical
fixture-only version `0018` has been retired from the active migration
directory; existing databases may retain its old `_sqlx_migrations` ledger row
and fixture data, which must not be reset or deleted as part of startup.
`cmd/seed_development` explicitly inserts the development departments,
locations, and admin through real PostgreSQL transactions and is guarded by
`APP_ENV=development` plus `SEED_DEVELOPMENT_DATA=true`.

The Go packages are not a promise that every product feature is implemented.
Later SPECs add only their required handlers, business code, SQL, and tests.

---

# 15. File Size and Readability

Avoid giant source files.

A file should have one clear responsibility.

Do not artificially split tiny functions across dozens of files, but also do not allow a single file to become a dumping ground.

If a source file becomes difficult to scan or contains multiple unrelated responsibilities, split it.

Prefer descriptive names:

```text
create_ticket
find_ticket_by_id
list_open_tickets
accept_ticket
assign_ticket
close_ticket
verify_password
create_session
```

Avoid vague names:

```text
handle
process
do_it
manager
helper
utils2
data
stuff
```

---

# 16. Error Handling

Use one clear application error model.

Preferred concept:

```go
type AppError struct {
    Code    string
    Message string
}
```

Exact implementation may evolve.

Do not scatter arbitrary status-code tuples throughout business logic.

Do not use `unwrap()` or `expect()` for normal request-path failures.

`unwrap()` may be acceptable in:

```text
tests
guaranteed startup invariants
developer-only scripts
```

when failure is intentionally fatal and obvious.

---

# 17. Logging and Observability

Use the standard `log` package for the foundation and add structured logging
only when the application needs it. Fiber middleware should remain explicit
and limited to the behavior the current SPEC requires.

Logs should help diagnose real production problems.

Log useful context such as:

```text
request id
route
HTTP status
latency
ticket id
authenticated user id where appropriate
database error category
WebSocket lifecycle
```

Do not log:

```text
passwords
password hashes
raw session tokens
secret keys
authorization cookies
sensitive credentials
```

---

# 18. Database Rules

PostgreSQL is the source of truth for persistent application data.

Use:

```text
pgx v5
pgxpool
sequential PostgreSQL SQL migrations
parameterized SQL
transactions where atomicity is required
```

Do not introduce a full ORM unless the user explicitly changes the decision.

SQL should remain visible and understandable.

---

# 19. SQL Rules

Prefer:

```text
explicit columns
parameterized queries
bounded results
purpose-built indexes
transactions for multi-step writes
```

Avoid:

```sql
SELECT *
```

for production endpoints unless there is a strong reason.

Prefer:

```sql
SELECT id, title, status, created_at
FROM tickets
WHERE status = $1
ORDER BY created_at DESC
LIMIT $2
```

Do not concatenate untrusted input into SQL.

---

# 20. Database Performance

The database is more likely to become a bottleneck than Fiber for this project.

Every high-frequency query should eventually be evaluated for:

```text
index usage
row count
sort behavior
join behavior
query plan
returned payload
```

Common access patterns must receive appropriate indexes.

Do not add indexes blindly because unnecessary indexes increase write cost and storage.

---

# 21. Transactions

Use database transactions when multiple writes must succeed or fail as one operation.

Examples:

```text
accept ticket
+
insert ticket activity
+
create notification
```

or:

```text
close ticket
+
insert activity
+
notification
```

These should not leave half-completed state.

Do not wrap unrelated operations in giant transactions.

Keep transactions short.

---

# 22. No Mock Data — Global Rule

This is a critical project rule.

> **Application features must use real backend + real PostgreSQL data from the beginning.**

Do not implement UI features using:

```text
hard-coded arrays
fake ticket objects
fake users
fake chat messages
fake notifications
static JSON pretending to be API data
temporary mock API routes
random generated UI data
frontend-only in-memory data pretending to be persisted
```

Example of forbidden frontend code:

```ts
const tickets = [
    {
        id: 1,
        title: "AC not cooling",
        status: "pending"
    }
];
```

when that array is being used as temporary application data.

The correct development flow is:

```text
PostgreSQL
    ↓
Fiber API
    ↓
SvelteKit
```

The frontend should consume real API responses.

---

# 23. What "No Mock Data" Means for Development

Development data must also go through the real database.

If the developer needs a user to test login:

```text
create that user in PostgreSQL
```

The application uses an admin-provisioned account model:

```text
username + password -> login
admin creates users
no public signup
no forgot-password or email-reset flow
```

Email is optional user profile/contact data only. It is never the login identifier.

The initial development administrator is provisioned with username `hothienty`.
Its password is supplied only through local development configuration and is
hashed with Argon2id before it is stored in PostgreSQL. Never commit that
password or a production credential to the repository.

If the developer needs tickets to test the ticket list:

```text
insert them into PostgreSQL
```

Then retrieve them through the real Fiber endpoint.

Do not bypass the application architecture for convenience.

---

# 24. What "No Mock Data" Means for Tests

Automated tests are allowed to create isolated test records because a test must control its own state.

However:

> Tests should use a **real PostgreSQL test database** where database behavior is under test.

Do not replace PostgreSQL with an in-memory fake repository simply to make tests easier unless a SPEC explicitly requires a pure unit test.

Integration tests should test the real stack:

```text
test
 ↓
Fiber router
 ↓
pgx/pgxpool
 ↓
test PostgreSQL database
```

Test fixtures are acceptable.

Mock production behavior is not.

Never connect automated tests to the production database.

---

# 25. No Fake API Success

Do not implement unfinished backend code like:

```go
return c.JSON(fiber.Map{
    "success": true,
})
```

just to make the frontend appear functional.

If a feature is unfinished:

```text
leave it unimplemented within the current SPEC scope
```

or return an appropriate explicit error during development.

Do not fake completion.

---

# 26. Frontend Architecture

The current frontend is a minimal SvelteKit scaffold using Bun for dependency installation and script execution. It currently contains the root layout/page, shared library entrypoint, static assets, and Vite/TypeScript configuration.

The feature-specific API, component, store, and route-group directories below are planned additions. Create them as the relevant SPECs are implemented rather than treating them as current folders.

Planned structure:

```text
frontend/
├── package.json
├── svelte.config.js
├── vite.config.ts
│
└── src/
    ├── lib/
    │   ├── api/
    │   ├── components/
    │   ├── stores/
    │   ├── types/
    │   ├── utils/
    │   ├── shared/
    │   ├── client/
    │   └── admin/
    │
    └── routes/
        ├── (auth)/
        │   └── login/
        │
        ├── (app)/
        │   ├── tickets/
        │   ├── staff-meal/
        │   ├── announcements/
        │   ├── report/
        │   └── settings/
        │
        └── admin/
```

Exact SvelteKit route-group details may evolve, but responsibilities should remain clear.

---

# 27. Frontend Component Rules

Prefer small components with clear roles.

Examples:

```text
TicketTable.svelte
TicketRow.svelte
TicketStatusBadge.svelte
TicketDetailPanel.svelte
TicketChat.svelte
ChecklistPanel.svelte
AnnouncementCard.svelte
StaffMealCard.svelte
MobileDrawer.svelte
```

Do not create one massive page component containing all ticket logic, chat logic, dialog logic, API calls, and state management.

Do not over-componentize trivial markup.

---

# 28. Frontend Data Access

Centralize API access.

Prefer:

```text
src/lib/api/auth.ts
src/lib/api/tickets.ts
src/lib/api/chat.ts
src/lib/api/announcements.ts
```

Do not scatter raw `fetch()` calls randomly across many visual components.

Visual components should not need to know backend URL construction details.

---

# 29. Shared Types

Keep frontend types explicit.

Examples:

```ts
export type TicketStatus =
    | 'pending'
    | 'accepted'
    | 'closed';
```

Do not use `any` to avoid defining data structures.

The frontend contract should match backend API response structures.

Avoid unnecessary duplicated transformation layers.

---

# 30. UX Performance Goal

The application should feel immediate.

Perceived performance matters as much as backend benchmark throughput.

Prioritize:

```text
instant button feedback
optimistic UI where safe
minimal full-page loading states
skeletons for data loading
no unnecessary page refresh
no unnecessary full-table refetch
small payloads
realtime updates
responsive interactions
stable layout
fast route transitions
```

---

# 31. Optimistic UI

Use optimistic updates where rollback is safe and behavior is obvious.

Example:

```text
User clicks Accept
        ↓
UI changes Pending -> Accepted immediately
        ↓
request sent to Fiber
        ↓
success -> keep state
failure -> rollback + show error
```

Do not wait unnecessarily for the entire page to reload.

However, do not use optimistic updates for operations where pretending success would be dangerous or confusing.

---

# 32. Realtime Rules

Realtime is required for areas where multiple users should see changes promptly.

Primary examples:

```text
ticket status changes
ticket assignment
ticket chat messages
possibly checklist updates
notifications
```

WebSocket should push relevant events.

Do not poll the entire application every few seconds if a targeted realtime event solves the problem better.

Do not send the entire ticket list through WebSocket for every small update.

Prefer compact events such as:

```json
{
  "type": "ticket.accepted",
  "ticket_id": 8020,
  "accepted_by": 42
}
```

---

# 33. Realtime Data Ownership

PostgreSQL remains the source of truth.

WebSocket is a delivery mechanism, not the authoritative database.

Correct model:

```text
write to PostgreSQL
      ↓
commit
      ↓
broadcast realtime event
      ↓
clients update
```

Do not make important ticket state exist only in WebSocket memory.

---

# 34. Authentication Direction

Authentication will use server-side sessions unless a later SPEC explicitly changes the decision.

Authentication identity is the user `username`, not email.

The account lifecycle is admin-controlled:

```text
IT admin creates and disables users
users log in with username + password
there is no public account creation
there is no forgot-password or email-reset flow
```

Email remains optional contact/profile data and must not be accepted as an
alternative login identifier.

Expected endpoints:

```text
POST /api/v1/auth/login
POST /api/v1/auth/logout
GET  /api/v1/auth/me
```

Expected browser mechanism:

```text
HttpOnly cookie
Secure in production
SameSite configured appropriately
```

Raw session tokens must not be stored in frontend local storage.

Passwords must never be stored in plaintext.

Use Argon2id for password hashing unless a later security decision explicitly replaces it.

---

# 35. Authorization

Authentication answers:

```text
Who is this user?
```

Authorization answers:

```text
May this user perform this action?
```

Do not treat them as the same thing.

Authorization rules belong in the backend.

Frontend button visibility is UX only and is not a security boundary.

An unauthorized user must still be rejected by Fiber even if they manually call an API endpoint.

---

# 36. Ticket Lifecycle

Locked ticket states:

```text
pending
accepted
closed
```

Meaning:

## Pending

Ticket exists but has not yet been accepted.

## Accepted

Ticket has been accepted.

## Closed

Ticket has been completed/closed.

Assignment is separate:

```text
assigned_to
assigned_at
```

Therefore:

```text
Accepted + unassigned
```

and:

```text
Accepted + assigned
```

are both valid.

Do not create an `assigned` ticket status unless the user explicitly changes this product rule.

---

# 37. Ticket Actions

Expected actions include:

```text
Accept
Assign
Close
```

Actions must update both persistent state and relevant history/audit data where the corresponding SPEC requires it.

Buttons should become disabled or visually inactive when an action is no longer valid.

Backend validation remains authoritative.

---

# 38. Ticket Detail — Desktop UX

On desktop:

> Ticket detail/chat opens in a right-side panel and must **not cover the ticket table**.

The overall ticket list context should remain visible.

The panel may be full-height.

The chat/detail panel should feel integrated into the desktop workspace rather than behaving like a generic modal.

---

# 39. Ticket Detail — Mobile UX

On mobile:

> Ticket chat/detail is a dedicated mobile screen.

Do not squeeze the desktop side panel onto a narrow screen.

---

# 40. Checklist UX

Checklist is hidden by default.

It becomes visible through the checklist control / three-dot action according to the final UI implementation.

On mobile, checklist can occupy its own ticket view/tab.

Checklist content belongs to the ticket and must come from the real backend/database.

---

# 41. Mobile Ticket Tabs

Mobile ticket navigation uses only:

```text
Open
Closed
```

There is no separate Pending tab.

Inside Open:

```text
Pending
Accepted
```

appear as ticket statuses.

---

# 42. Desktop Ticket Tabs

Desktop ticket navigation also centers on:

```text
Open
Closed
```

Open contains both:

```text
Pending
Accepted
```

---

# 43. Staff Meal

Staff Meal displays a complete menu image uploaded by admin.

It is not a separate menu image per weekday.

The image comes from real persisted application data.

No hard-coded placeholder meal content should remain in production UI.

---

# 44. Announcements

Announcements are created by admin and displayed as real persisted content.

Mobile and desktop may use different responsive layouts, but they consume the same backend data.

Pagination should be supported.

---

# 45. Report

Report must query real application data.

Never build charts from hard-coded numbers.

Report performance is important.

Avoid:

```text
loading all tickets into Go application memory
then calculating every report in application memory
```

when PostgreSQL can aggregate efficiently.

Prefer SQL aggregation for database-owned data.

Optimize report queries using real measurements.

Caching may be introduced later when justified.

---

# 46. Settings

Expected settings include:

```text
profile display
theme
language where implemented
notification preferences where implemented
```

User profile, account status, and password changes are admin-controlled unless
a later SPEC explicitly introduces a separate self-service rule. The product
does not include a forgot-password flow.

Do not implement editable fields that the product does not permit the user to edit.

---

# 47. Theme

The product supports:

```text
Dark theme
Light theme
```

Design direction:

```text
Dark:
red + black / charcoal
modern hotel operations feel

Light:
white + blue
clean SARA-inspired direction
```

Do not create a completely different product layout between themes.

Theme affects presentation, not product behavior.

---

# 48. Responsive Design

Responsive behavior is part of every frontend feature SPEC.

Do not create:

```text
one desktop-only implementation
then a separate unrelated mobile application
```

Use one responsive SvelteKit web application.

Every relevant UI SPEC should define and verify:

```text
Desktop
Tablet
Mobile
```

A feature is not complete if it only looks correct on desktop.

---

# 49. Public Deployment Direction

BWP SonaSea is intended to run through a real public domain.

Do not architect the product around:

```text
192.168.x.x-only access
hard-coded LAN URLs
hard-coded localhost URLs
```

Development may use localhost.

Production configuration must come from environment/configuration.

---

# 50. API Design

Use a consistent versioned API namespace:

```text
/api/v1
```

Examples:

```text
POST /api/v1/auth/login
GET  /api/v1/auth/me

GET  /api/v1/tickets
GET  /api/v1/tickets/:id
POST /api/v1/tickets
POST /api/v1/tickets/:id/accept
POST /api/v1/tickets/:id/assign
POST /api/v1/tickets/:id/close
```

Exact REST shape may be refined by the relevant SPEC.

Keep naming consistent.

Do not create multiple styles for similar endpoints.

---

# 51. API Response Design

Prefer explicit stable response models.

Do not expose raw database rows accidentally.

Avoid wrapping every endpoint in meaningless structures if unnecessary.

Good:

```json
{
  "id": 8020,
  "title": "fix vòi lỏng",
  "status": "accepted"
}
```

A consistent envelope may be used only if there is a clear project-wide reason.

Errors should have a consistent format.

Example concept:

```json
{
  "error": {
    "code": "ticket_not_found",
    "message": "Ticket not found"
  }
}
```

---

# 52. Pagination

Large collections must be paginated.

Examples:

```text
tickets
messages
announcements
notifications
audit logs
```

Do not load thousands of records just because the current database is small.

Initial pagination may be offset-based where appropriate.

Use cursor pagination later if scale/query behavior justifies it.

---

# 53. Search and Filtering

Search/filter should be done at the appropriate layer.

Do not fetch the complete database table into SvelteKit and filter everything in the browser for production-scale lists.

For tickets, the backend/database layer should handle query filters such as:

```text
status
department
requester
assignee
date range
search text
```

when those filters are implemented.

---

# 54. File Uploads

Do not store large uploaded image/file binaries directly in PostgreSQL.

Store:

```text
object key
URL if appropriate
original filename
MIME type
size
dimensions where useful
metadata
```

Validate:

```text
allowed MIME types
maximum size
ownership/authorization
```

Do not trust only the filename extension.

---

# 55. Image UX

Images should not make chat or Staff Meal feel slow.

Where appropriate:

```text
upload once
store externally
serve efficiently
use thumbnails/previews
lazy-load large images
avoid repeatedly transferring full-resolution originals
```

Do not prematurely invent image pipelines before the upload SPEC defines the actual requirements.

---

# 56. Dependency Rule

Keep the dependency graph small.

Before adding a crate/package, ask:

```text
What problem does it solve?
Can the standard library/current stack already solve it clearly?
Is it maintained?
Does it materially reduce correct implementation effort?
```

Do not add dependencies for tiny helpers that can be expressed clearly in a few lines.

Do not reimplement serious security/crypto primitives manually.

---

# 57. Backend Dependency Baseline

The backend foundation uses only the packages required by the current SPEC:

```text
github.com/gofiber/fiber/v3
github.com/jackc/pgx/v5
github.com/joho/godotenv
golang.org/x/crypto/argon2
```

Add a dependency only when the relevant SPEC requires it and its purpose is
clear. Use the Go standard library when it solves the problem clearly.

Additional dependencies are added only as corresponding features require them.

Examples:

```text
argon2     -> password hashing
cookie-related support -> sessions if needed
uuid       -> if a SPEC chooses UUID identifiers
```

Do not add every possible package on day one.

---

# 59. Fiber and Go Runtime Rule

Fiber v3 owns the HTTP server and routing. Use `context.Context` for bounded
database and application work. Do not enable global allocation-heavy options,
add middleware, or start goroutines unless a concrete requirement justifies it.

Fiber-compatible middleware may provide:

```text
CORS
request IDs
timeouts
compression where justified
```

---

# 60. SvelteKit Rendering Strategy

Do not blindly force every page into SSR or CSR.

Choose rendering/data-loading behavior based on the feature.

Authentication, protected routes, realtime tickets, and chat should be designed for a smooth authenticated application experience.

Avoid full page reloads for routine ticket operations.

---

# 61. State Management

Use the simplest Svelte state mechanism that clearly solves the feature.

Do not introduce a large external state-management library unless required.

Keep server state and UI state conceptually separate.

Persistent source of truth:

```text
backend/shared/database
```

Temporary interactive state:

```text
Svelte component/store state
```

---

# 62. Accessibility and Interaction

Interactive UI should remain usable with:

```text
keyboard
visible focus states
semantic buttons
appropriate labels
reasonable contrast
```

Do not sacrifice usability purely to match a visual mockup.

---

# 63. Security Baseline

Security-sensitive behavior must be handled intentionally.

Required principles:

```text
Argon2id password hashing
server-side authorization
HttpOnly auth cookie
Secure cookie in production
SameSite policy
session expiration
session revocation
parameterized SQL
input validation
upload validation
no secret logging
no production secrets in git
rate limiting where later justified
CORS restricted to intended origins
```

Do not implement custom cryptography.

---

# 64. CORS

Development CORS may allow the local SvelteKit origin.

Production must use explicit allowed origin(s).

Do not ship:

```text
Allow-Origin: *
```

with credentialed authentication.

---

# 65. Configuration

Environment-specific values belong in environment/configuration.

Examples:

```text
DATABASE_URL
APP_ENV
BACKEND_BIND_ADDRESS
FRONTEND_ORIGIN
SESSION_SECRET / SESSION CONFIG
OBJECT_STORAGE CONFIG
LOG FILTER
```

Do not hard-code deployment-specific domains or credentials in source code.

Use the existing local `.env` file or the deployment secret manager. Never
generate `.env.example` files, and never commit local credentials.

---

# 66. Testing Philosophy

Testing should focus on important behavior, not artificial coverage percentages.

Priority:

```text
auth
authorization
ticket lifecycle
database constraints
transactions
chat persistence
critical API behavior
report correctness
```

Backend integration tests should prefer real Fiber routing + a real PostgreSQL test database for important flows.

Frontend tests should focus on behavior that is easy to regress.

End-to-end tests can be introduced for critical user journeys.

---

# 67. Critical End-to-End Journeys

Eventually, the system should be able to verify flows like:

```text
Login
 ↓
Open Tickets
 ↓
Create/View Ticket
 ↓
Accept
 ↓
Assign
 ↓
Chat
 ↓
Checklist
 ↓
Close
```

and:

```text
Admin publishes announcement
 ↓
staff sees announcement
```

and:

```text
Admin uploads Staff Meal
 ↓
staff sees current menu
```

These flows must use real application APIs and database state.

---

# 68. Completion Checks — Backend

Before a backend SPEC is called complete, run where applicable:

```bash
gofmt -d .
go vet ./...
go test ./...
go build ./...
```

Also verify:

```text
migrations
database queries
HTTP status behavior
authorization behavior
error paths
```

Do not claim completion without verification.

---

# 69. Completion Checks — Frontend

Before a frontend SPEC is complete, run the relevant project checks, such as:

```text
format
lint
Svelte/TypeScript checks
tests where present
production build
```

Exact commands should follow the package configuration in the repository.

Also manually verify responsive behavior for:

```text
desktop
tablet
mobile
```

---

# 70. Scope Discipline

Each SPEC has a defined scope.

Do not use one SPEC as permission to refactor the entire repository.

When implementing a SPEC:

```text
read this context
read the SPEC
inspect existing code
change only what the SPEC requires
preserve existing working behavior
verify the result
```

Avoid speculative future work.

---

# 71. No Premature Architecture

Do not introduce these only because the application may grow someday:

```text
microservices
Kafka
RabbitMQ
Kubernetes
CQRS
event sourcing
service mesh
distributed cache
read replicas
sharding
generic plugin architecture
```

Start with:

```text
one SvelteKit frontend
one Fiber v3 backend
one PostgreSQL database
```

This is the default architecture.

Scale vertically and optimize actual bottlenecks before introducing distributed complexity.

---

# 72. Monolith Strategy

The backend is a modular monolith.

Meaning:

```text
one deployable Fiber v3 service
```

with clear internal modules.

Example:

```text
auth
tickets
chat
checklist
announcements
staff_meal
reports
settings
admin
```

Modules should have clear boundaries without pretending they are networked microservices.

---

# 73. Admin Strategy

Admin functionality belongs to the same overall system unless a later SPEC defines a separate deployable application.

Frontend code may separate:

```text
shared
client
admin
```

for maintainability.

Backend may separate admin handlers/services where behavior materially differs.

Do not duplicate common business logic between client and admin modules.

---

# 74. Source of Truth

When two representations disagree, use this priority:

```text
1. Current explicit user instruction
2. Current SPEC
3. This PROJECT_CONTEXT.md
4. Existing implementation behavior
5. Developer assumptions
```

However, a SPEC should not silently alter a locked technology decision.

If a conflict is meaningful, surface it.

Do not guess.

---

# 75. Current Product UI Direction

The UI reference currently establishes:

## Desktop

```text
Left sidebar:
- Tickets
- Report
- Settings

Bottom sidebar:
- user profile/avatar

Tickets:
- Open
- Closed

Open statuses:
- Pending
- Accepted

Ticket table:
- requester
- location
- title
- status
- created time
- actions

Actions:
- Chat
- Accept
- Assign
- Close

Right side:
- Ticket detail
- Chat
- Checklist

Bottom content:
- Staff Meal
- Announcements
```

Ticket detail on desktop must not cover the ticket table.

---

# 76. Current Mobile UI Direction

Mobile includes:

```text
Login

Hamburger menu

Tickets
├── Open
└── Closed

Ticket Chat

Ticket Checklist

Staff Meal

Announcements

Report

Settings
```

The logo belongs in the menu/drawer design rather than being duplicated unnecessarily in both drawer and header.

Open tickets display:

```text
Pending
Accepted
```

Closed tickets live under Closed.

---

# 77. Staff Meal UI Rule

Staff Meal uses **one complete menu image** uploaded by admin.

Do not split it into one menu image per day.

Do not add a "View All" workflow unless a later explicit UI requirement adds one.

---

# 78. Announcement UI Rule

Announcements are separate content entries created by admin.

On mobile they should be presented as their own page/list rather than being artificially merged with Staff Meal.

---

# 79. Report UX Rule

Report should load quickly enough to feel like part of the application.

The existing inspiration system may have very slow reports; BWP SonaSea must not reproduce that behavior.

Report speed should be solved through:

```text
correct SQL
indexes
appropriate aggregation
pagination
bounded date ranges
caching only if later needed
```

not by simply changing frameworks.

---

# 80. User Experience Target

The desired perception is:

```text
open application
→ immediate shell

open ticket
→ detail appears quickly

accept ticket
→ status changes immediately

send message
→ message appears immediately

another user changes ticket
→ current screen updates without reload

switch Open/Closed
→ fast transition

open Staff Meal
→ image loads efficiently

open Report
→ no multi-minute waiting
```

---

# 81. Performance Targets

These are engineering targets, not absolute guarantees over every internet connection.

Aim for:

```text
UI interaction feedback:
near-instant, generally < 50 ms perceived

ordinary backend API work:
typically tens of milliseconds where DB work is simple

simple indexed DB queries:
typically low milliseconds under healthy local conditions

chat/ticket realtime propagation:
near realtime

route transitions:
no unnecessary full reload
```

Do not manipulate measurements to satisfy targets.

Measure end-to-end behavior.

---

# 82. No Benchmark Theater

Do not choose implementation techniques solely because they look good in synthetic framework benchmarks.

Real application performance matters more than maximum empty-request throughput.

For BWP SonaSea, optimize:

```text
actual ticket endpoints
actual report queries
actual chat behavior
actual image handling
actual concurrent users
```

---

# 83. Readability Over Cleverness

Prefer:

```go
ticket, err := tickets.FindByID(ctx, state.DB, id)
if err != nil {
    return err
}
```

over a clever abstraction requiring several files and generic constraints to understand.

Prefer:

```go
switch ticket.Status {
case TicketStatusPending:
    // ...
case TicketStatusAccepted:
    // ...
case TicketStatusClosed:
    // ...
}
```

when it is clearer than hidden dynamic dispatch.

Readable code is a long-term performance feature for the development team.

---

# 84. Explicit Business Names

Use domain language.

Prefer:

```text
ticket
requester
assignee
department
announcement
staff_meal
checklist_item
session
```

Avoid generic abstractions like:

```text
entity
resource_manager
data_processor
generic_item
```

unless the concept is genuinely generic.

---

# 85. Comments

Comments should explain:

```text
why
business constraints
non-obvious performance decisions
security decisions
edge cases
```

Do not comment obvious syntax.

Bad:

```go
// Get ticket
ticket, err := getTicket(ctx, id)
if err != nil {
    return err
}
```

Useful:

```go
// Assignment does not transition ticket status.
// A ticket remains `accepted` while assigned to a technician.
```

---

# 86. Documentation

Important architecture decisions should be captured in documentation/specs rather than existing only in chat history.

The current repository stores these project documents at:

```text
Context-Spec-BWP-SonaSea/PROJECT_CONTEXT.md
Context-Spec-BWP-SonaSea/Spec/
README.md
docs/baron/
```

The `Context-Spec-BWP-SonaSea/Spec/` directory is the current SPEC location. Keep new SPEC documents there unless the project explicitly adopts a different documentation layout.

Recommended project docs:

```text
Context-Spec-BWP-SonaSea/PROJECT_CONTEXT.md
Context-Spec-BWP-SonaSea/Spec/
README.md
```

Optional later:

```text
docs/architecture/
docs/api/
docs/deployment/
docs/ui/
```

Do not create documentation directories full of empty placeholders.

---

# 87. SPEC Workflow

The repository is currently at the foundation-scaffold stage. `SPEC-01-database-foundation.md` is stored under `Context-Spec-BWP-SonaSea/Spec/`; later SPECs should add the corresponding production modules incrementally.

The intended implementation sequence is approximately:

```text
SPEC-01
Database Foundation

SPEC-02
Backend Foundation

SPEC-03
Authentication Backend

SPEC-04
Login UI + Authentication Integration

SPEC-05
Application Shell + Navigation + Responsive Foundation

SPEC-06
Ticket List

SPEC-07
Ticket Detail + Accept/Assign/Close

SPEC-08
Realtime Chat + Attachments

SPEC-09
Checklist

SPEC-10
Staff Meal

SPEC-11
Announcements

SPEC-12
Report

SPEC-13
Settings

SPEC-14
Production Hardening / QA / Performance / Security
```

Admin-specific work may add additional SPECs.

The exact number of SPECs may evolve.

---

# 88. Vertical Slice Rule

Do not build the entire backend first and postpone frontend integration until the end.

Preferred feature workflow:

```text
database foundation
        ↓
backend feature
        ↓
frontend feature
        ↓
connect end-to-end
        ↓
verify real behavior
        ↓
next feature
```

Example:

```text
auth database
 ↓
auth backend
 ↓
login UI
 ↓
real login integration
 ↓
verified login
```

Then move to tickets.

---

# 89. Implementation Order Inside a SPEC

Unless the SPEC requires otherwise:

```text
1. Read PROJECT_CONTEXT.md
2. Read current SPEC
3. Inspect existing repository
4. Identify affected modules
5. Implement database/migration changes if needed
6. Implement backend behavior
7. Implement frontend behavior if in scope
8. Connect using real API/database data
9. Test error states
10. Run verification commands
11. Report exactly what changed
```

---

# 90. AI Coding Agent Rules

An AI agent working on this repository must:

```text
read before editing
respect scope
preserve architecture
avoid speculative abstractions
avoid mock data
use real PostgreSQL-backed flows
keep code readable
run verification
report blockers truthfully
```

The agent must not silently:

```text
change framework
replace pgx/pgxpool
introduce an ORM
introduce microservices
introduce Redis
replace session auth with JWT
rewrite folder structure
add a UI library
add mock data
change ticket states
```

unless explicitly instructed.

---

# 91. AI Must Not Hide Complexity

Do not solve a simple application problem with architecture the human maintainer cannot reasonably understand.

If advanced Go is genuinely necessary:

1. Keep it isolated.
2. Explain why it is required.
3. Prefer the smallest advanced construct that solves the problem.
4. Add a concise code comment if future maintainers need context.

---

# 92. AI Must Not Fake Completion

If something cannot be completed:

```text
state the blocker
state what remains
do not simulate success
```

Never:

```text
hard-code a successful response
insert fake UI data
disable a failing test
remove validation to make tests pass
swallow an error
```

just to claim the SPEC is complete.

---

# 93. AI Dependency Discipline

When an agent wants to add a dependency, it should be able to state:

```text
dependency:
purpose:
why existing stack is insufficient:
scope where used:
```

Do not add large dependency sets preemptively.

---

# 94. AI Refactoring Discipline

Refactor only when:

```text
required by the current SPEC
needed to prevent duplication introduced by the current work
needed to fix a concrete defect
needed for measurable performance/security correctness
```

Do not perform broad aesthetic refactors while implementing an unrelated feature.

---

# 95. Version Upgrade Discipline

Major framework/library releases do not require immediate migration.

If the current version is:

```text
working
supported
secure
stable
```

do not upgrade solely because a newer major version exists.

Upgrade deliberately:

```text
read migration notes
create dedicated branch/work
run full tests
measure behavior
deploy safely
```

---

# 96. Long-Term Maintainability

The codebase should remain understandable even if the original author returns after a long absence.

That means:

```text
predictable folders
domain names
explicit SQL
small modules
few hidden abstractions
stable API patterns
consistent error handling
consistent tests
documented business constraints
```

This goal is as important as raw throughput.

---

# 97. Final Engineering Rule

The default question is not:

> "What is the most advanced Go way to implement this?"

The default questions are:

> "What is the simplest correct implementation?"

> "Will it remain fast under the expected workload?"

> "Can we measure it?"

> "Will the maintainer still understand it later?"

Only introduce complexity when the simpler implementation no longer satisfies correctness, security, performance, or maintainability requirements.

---

# 98. Permanent Project Summary

```text
PROJECT:
BWP SonaSea

CURRENT STAGE:
Foundation scaffold; feature modules and migrations are added incrementally by SPEC

FRONTEND:
SvelteKit + Svelte + TypeScript

FRONTEND TOOLING:
Bun

BACKEND:
Go + Fiber v3

DATABASE:
PostgreSQL + pgx v5 + pgxpool

ARCHITECTURE:
Modular monolith

REALTIME:
WebSocket where needed

AUTH:
Server-side sessions

STYLE:
Boring, explicit, readable Go

PERFORMANCE:
Measure first; optimize real bottlenecks

DATA:
Real PostgreSQL-backed application data from day one

MOCK DATA:
Forbidden for application implementation

UI:
Responsive desktop + tablet + mobile

THEMES:
Dark + Light

DEPLOYMENT:
Public domain

TICKET STATUS:
pending
accepted
closed

ASSIGNMENT:
Separate from ticket status

PREMATURE INFRA:
Avoid

PRIMARY GOAL:
Fast UX + maintainable production system

CONTEXT LOCATION:
Context-Spec-BWP-SonaSea/PROJECT_CONTEXT.md

SPEC LOCATION:
Context-Spec-BWP-SonaSea/Spec/
```

---

# 99. Instruction to the AI Before Every SPEC

Before implementing any SPEC, confirm internally that you understand these non-negotiable rules:

```text
Do not change the locked stack.

Do not use mock application data.

Do not fake API success.

Use real PostgreSQL-backed flows.

Keep Go boring and explicit.

Do not add Arc/Mutex/generics/traits without a concrete reason.

Do not sacrifice performance through naive SQL or data flow.

Do not sacrifice readability for theoretical optimization.

Do not over-engineer.

Respect responsive desktop/mobile requirements.

Stay inside the current SPEC scope.

Verify before claiming completion.
```

Then read the relevant SPEC and implement only that scope.
