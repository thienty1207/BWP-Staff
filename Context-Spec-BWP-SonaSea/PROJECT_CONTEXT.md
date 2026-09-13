# BWP SonaSea — PROJECT_CONTEXT (Current Canonical Context)

> **Last context refresh:** 2026-09-13
>
> **Repository:** `thienty1207/BWP-Staff`
>
> **Verified source baseline while this context was prepared:** `main` at `23bf1097808c7c4b4aa457332fb5a63dffc109c1`
>
> **Important:** SPEC-06.7 cleanup has been merged and reviewed. `ROOM-8020` and `ROOM-7309` remain stored but inactive, canonical `BWP-ROOM-8020` / `BWP-ROOM-7309` remain active, and the 06.7 integration test file now uses a descriptive domain filename. SPEC-06.7 is CLOSED.

---

# 0. How to use this file

This is the current canonical project context for **BWP SonaSea**.

Every coding agent must read:

1. `AGENTS.md`
2. this `PROJECT_CONTEXT.md`
3. the current SPEC
4. the existing source code in the affected modules

before editing.

Priority when instructions conflict:

```text
1. Latest explicit user instruction
2. Latest later SPEC / explicit supersession note
3. This PROJECT_CONTEXT.md
4. Earlier closed SPEC
5. Existing implementation behavior
6. Developer assumptions
```

Closed SPECs are historical implementation contracts. Later SPECs may deliberately supersede selected UI or behavior details. Do not blindly copy an older SPEC when a later SPEC explicitly changed that rule.

If a visual reference is supplied by the user and the user says to follow it, treat that image as a **binding visual/interaction contract** for the requested elements. Do not "improve" it into a materially different hierarchy or behavior unless the user explicitly permits deviation.

---

# 1. Project

```text
Name: BWP SonaSea
Type: hotel operations / internal support platform
Architecture: modular monolith
Deployment direction: real public domain
```

Primary goals:

```text
fast perceived UX
correct persistent state
high backend performance
maintainable boring code
clear architecture
responsive desktop/mobile UI
long-term stability
real PostgreSQL-backed behavior
```

This is a real product, not a mock/demo application.

---

# 2. Locked technology stack

## Frontend

```text
SvelteKit
Svelte 5
TypeScript
Bun
Vite
```

## Backend

```text
Go 1.27
Fiber v3
context.Context
standard log unless a later observability SPEC changes it
```

Current module baseline includes:

```text
github.com/gofiber/fiber/v3 v3.5.0
github.com/jackc/pgx/v5 v5.11.0
github.com/joho/godotenv v1.5.1
golang.org/x/crypto
```

## Database

```text
PostgreSQL
pgx v5
pgxpool
explicit parameterized SQL
no ORM
```

## Realtime direction

```text
WebSocket only when a later realtime feature actually needs it
```

Do not add Redis, Kafka, RabbitMQ, Kubernetes, microservices, CQRS, event sourcing, or distributed caching without a measured requirement and an explicit later architecture decision.

---

# 3. Core engineering philosophy

Default rule:

> **Boring, explicit code by default. Optimize from measurements.**

Code should be:

```text
readable
small
typed
explicit
predictable
safe
fast
easy to debug
low-magic
```

Do not introduce complexity to look "enterprise".

Avoid unless concretely required:

```text
generic repository frameworks
deep interface hierarchies
reflection-heavy DI
factory-heavy patterns
global mutable state
unnecessary goroutines
custom global locks
premature microservices
```

PostgreSQL is the persistent source of truth.

---

# 4. Backend layering

Canonical request flow:

```text
HTTP
↓
Handler
↓
Service
↓
Repository
↓
pgx / pgxpool
↓
PostgreSQL
```

## Handler

Owns:

```text
HTTP extraction
HTTP validation
auth context
response mapping
```

## Service

Owns:

```text
business rules
business workflows
authorization decisions where required
coordination
```

## Repository

Owns:

```text
SQL
transactions
database reads/writes
database-specific behavior
```

Do not put large SQL statements in handlers.

---

# 5. Source/test filename rule

Source and test filenames must describe the domain/behavior.

Good:

```text
bwp_room_locations_integration_test.go
location_search_integration_test.go
ticket_list_integration_test.go
authentication_integration_test.go
```

Do not create new code/test filenames based on SPEC numbers:

```text
spec065_*
spec067_*
spec07_*
phase_*
task_*
```

SPEC numbers belong in documentation.

Historical files with SPEC-number names may be renamed in bounded housekeeping work; do not opportunistically rename unrelated files inside a feature SPEC.

---

# 6. Current repository feature modules

Current implemented backend feature packages include:

```text
backend/client/auth/
backend/client/lookups/
backend/client/tickets/
backend/admin/
backend/shared/
backend/config/
backend/app/
```

Current frontend implemented areas include:

```text
login
authenticated Tickets shell
Open / Closed ticket list
New Request dialog
Department lookup
searchable Location picker
Light/Dark authenticated shell
responsive desktop/mobile ticket list
```

Planned modules such as Chat, Checklist, Announcements, Staff Meal, Report, Settings, and Admin UI must not be created as empty scaffolding merely because they are future features.

---

# 7. Current implemented HTTP API

Currently implemented API surface:

```text
POST /api/v1/auth/login
POST /api/v1/auth/logout
GET  /api/v1/auth/me

GET  /api/v1/departments
GET  /api/v1/locations
GET  /api/v1/tickets
POST /api/v1/tickets
```

Not implemented yet:

```text
GET  /api/v1/tickets/:id
POST /api/v1/tickets/:id/accept
POST /api/v1/tickets/:id/assign
POST /api/v1/tickets/:id/close
ticket chat
checklist APIs
attachment upload APIs
notifications
staff meal
announcements
report
settings
admin UI flows
```

Do not fake these endpoints.

---

# 8. Authentication — implemented and locked

Authentication uses:

```text
username + password
server-side session
HttpOnly cookie
```

Email is not a login identifier.

There is:

```text
no public signup
no forgot-password flow
no email reset flow
```

Cookie:

```text
name = bwp_session
HttpOnly
SameSite=Lax
Secure outside development
```

Session token rules:

```text
32 random bytes
base64url raw token
only SHA-256 token hash stored in PostgreSQL
```

Session TTL:

```text
AUTH_SESSION_TTL_HOURS
default 12
accepted range 1..720
```

Inactive users invalidate session access.

Logout revokes the database session before clearing the browser cookie.

Request IDs are generated by the backend. Never trust an inbound client `X-Request-ID` as the authoritative request ID.

Password hashing:

```text
Argon2id
password max = 1024 bytes
PHC parameters are bounded before expensive verification
```

Remember Me on login stores **username only**. Never persist password or raw session token in localStorage.

---

# 9. Database migrations

Active schema migration sequence:

```text
0001..0017
0019
0020
```

`0018` is permanently retired and must never be reused.

Current forward ticket amendment:

```text
0020_ticket_assignment_and_request_fields.sql
```

It adds/locks:

```text
boolean priority
optional due_at
ticket attachment metadata foundation
normalized ticket department assignments
normalized ticket user assignments
removal of legacy single-assignment columns after migration/backfill
```

There is currently **no migration 0021**.

Never rewrite an already-applied migration. Future schema work is forward-only.

Reference/master data changes belong to the explicit development seed workflow unless a real schema change is required.

---

# 10. Local environment and PostgreSQL test contract

There is exactly one local environment file:

```text
backend/.env
```

Do not create:

```text
.env.example
.env.test
.env.testing
.env.local
.env.development
DATABASE_TEST_URL
TEST_DATABASE_URL
TEST_DB_URL
APP_ENV=test
```

Database-backed tests use:

```text
backend/.env
→ DATABASE_URL
→ real local PostgreSQL instance
→ unique temporary schema
→ run migrations/test state
→ drop only temporary schema
```

Never drop/truncate/reset the developer's normal public schema from automated tests.

---

# 11. No mock application data

Runtime application features must use:

```text
PostgreSQL
→ Fiber API
→ SvelteKit
```

Forbidden for product behavior:

```text
fake tickets
fake users
fake departments
fake locations
fake chat
static JSON pretending to be API data
frontend fallback arrays
fake success responses
```

Deterministic test fixtures in isolated temporary schemas are allowed.

Development master/reference data is allowed only through the real PostgreSQL seed workflow.

---

# 12. Ticket lifecycle

Locked statuses:

```text
pending
accepted
closed
```

No `assigned` status.

Assignment is separate from status.

`tickets.department_id` means:

```text
original request destination/family
```

It is not the current assignment state.

Current assignment supports relational:

```text
one or many departments
one or many users
departments + users simultaneously
```

Owner semantics in the Ticket List:

```text
Owner = first user who accepted the ticket = accepted_by
```

Assignment must not redefine Owner.

Requester semantics:

```text
Requester = original ticket creator = requester_id
```

---

# 13. Ticket list backend

Implemented:

```http
GET /api/v1/tickets
```

Views:

```text
view=open
view=closed
```

Open contains:

```text
pending
accepted
```

Closed contains:

```text
closed
```

Authenticated active staff currently read the shared ticket list; there is no current per-user read filter.

Pagination:

```text
keyset pagination
ORDER BY created_at DESC, id DESC

default limit = 50
max limit = 100

cursor:
before_created_at
before_id
```

Repository behavior must avoid N+1 queries. Current assignment data is bulk-loaded, not fetched once per ticket.

---

# 14. Current Ticket List UI

Desktop columns are currently:

```text
Requester
Location
Title
Description
Status
Owner
Created On
Due Date
```

Do not reintroduce:

```text
Ticket ID
Department column
Assignment column
Action column
```

until a later SPEC explicitly changes the table.

Requester label:

```text
Full Name (DepartmentCode)
```

Owner:

```text
accepted_by only
pending → —
accepted/closed → first accepter
```

Description is a compact real database preview.

Created On / Due Date:

```text
date on first visual line
time on second visual line
```

Null Due Date:

```text
—
```

Priority presentation:

```text
NO visible "Priority" badge
NO red Title text

priority=true:
normal theme Title text
+ thin red/danger content-sized border
+ small radius/tight padding
+ max-width 100%
+ natural wrapping

priority=false:
normal Title
```

Short Title → short border.
Long Title → border grows/wraps with content.

This priority rule supersedes older SPEC-06.1 and SPEC-06.5 visual rules.

---

# 15. Mobile Ticket List

Mobile uses compact cards.

Do not revert to the old tall developer-style card that stacks every backend field vertically.

Do not display Ticket ID.

The mobile card remains a summary. Ticket Detail/Chat navigation is intentionally deferred to SPEC-07+.

Priority uses the same rule as desktop:

```text
normal text
red content-sized border
no Priority badge
```

---

# 16. Create Ticket — implemented

Implemented:

```http
POST /api/v1/tickets
```

Request fields currently accepted by the backend:

```text
department_id  required
location_id    optional
title          required
description    optional
priority       boolean
due_at         optional RFC3339 timestamp
```

Requester always comes from the authenticated principal. The frontend must never choose/requester spoof a requester ID.

New tickets are:

```text
status = pending
accepted_by = NULL
accepted_at = NULL
closed fields = NULL
```

Create transaction verifies active department/location with PostgreSQL locking and atomically writes:

```text
ticket
+
ticket_activity(created)
```

If activity insertion fails, the ticket creation transaction rolls back.

Create endpoint must not automatically retry POST on network/server failures.

Frontend preserves form input on safe validation errors.

---

# 17. New Request — current UI

Current New Request includes:

```text
Department
Location
Request title
Description
Priority boolean
Due time
```

Department values come from PostgreSQL.

Location values come from PostgreSQL.

Current create request does **not yet implement actual image attachment upload**, even though ticket attachment schema/product direction exists. Do not fake image upload and do not claim attachment creation is implemented until a dedicated upload SPEC adds it.

---

# 18. Department master data

Approved requestable development departments:

```text
Concierge
Damaged Asset
F&B
Finance Request
Front Office
Housekeeping
Housekeeping PPM
IT
Kitchen
Laundry
Lost & Found
Maintenance
REC
Security
```

These are PostgreSQL-backed.

Legacy development fixture departments:

```text
Engineering
Human Resources
```

are non-requestable/inactive only when safely identified as development fixture rows.

Do not deactivate unrelated real department rows.

---

# 19. Location master data ownership

Location data is PostgreSQL-backed and separated by explicit fixture ownership markers.

## 96 Villas

Marker:

```text
Development location seed data: 96 Villas
```

Expected:

```text
42 named areas
99 room rows 1001..1099
141 total fixture rows
```

## BWP main-property areas

Marker:

```text
Development location seed data: BWP
```

Expected:

```text
155 rows
```

Source labels are preserved exactly, including original capitalization/punctuation/typos.

## BWP rooms

Marker:

```text
Development location seed data: BWP Rooms
```

Expected:

```text
564 exact approved room rows
```

Visible room name:

```text
plain numeric room number
```

Internal code:

```text
BWP-ROOM-<number>
```

The room set is sparse and source-driven. Never infer missing numbers from ranges.

## Generic development locations

Historical generic fixture rows include:

```text
Lobby
Ballroom
Back Office
Room 8020
Room 7309
Villa
```

Canonical post-SPEC-06.7-cleanup state:

```text
Lobby        active
Ballroom     active
Back Office  active
Villa        active

ROOM-8020    still exists but inactive
ROOM-7309    still exists but inactive

BWP-ROOM-8020 active, visible name "8020"
BWP-ROOM-7309 active, visible name "7309"
```

Do not delete the legacy room rows because existing foreign keys may reference them.

The cleanup above is merged and verified on the current `main` baseline.

---

# 20. Development fixture aggregate

After SPEC-06.7 room data:

```text
generic development location rows = 6 stored
96 Villas                    = 141
BWP areas                    = 155
BWP rooms                    = 564
-----------------------------------
stored development locations = 866
```

After the legacy-room cleanup, two of the six generic rows are inactive, so the development fixture set contributes:

```text
864 active location rows
```

This is **not** a guarantee that `GET /api/v1/locations` has exactly 864 rows because unrelated real active locations may coexist.

Do not write production logic that depends on a hardcoded global row count.

---

# 21. Location lookup/search API

Implemented:

```http
GET /api/v1/locations
```

Optional query parameters:

```text
q
limit
```

`q`:

```text
trim whitespace
case-insensitive
max 100 Unicode code points
literal text semantics
% and _ are NOT SQL wildcard behavior
```

Search ranking:

```text
0 exact name
1 full-name prefix
2 token/word prefix
3 contains
4 name ASC
5 id ASC
```

Only active rows are returned.

`limit`:

```text
optional
when omitted → no artificial result cap
when provided → integer 1..30
```

Do not restore the older SPEC-06.5 behavior that silently defaulted Location search to 10 rows.

Filtering/ranking belongs in PostgreSQL, not by loading the table into Go and filtering in memory.

---

# 22. Searchable Location picker

The current Location control is a searchable combobox/listbox style interaction.

Canonical behavior:

```text
focus/click
→ full active catalogue is available when query is empty
→ panel remains visually bounded
→ internal vertical scrolling
→ typing sends real backend search
→ ~180 ms debounce
→ newest request wins
→ stale responses ignored
→ loading state cannot select an old result
→ select stores real location.id
→ visible label = location.name only
→ clear means location_id = null
```

Do not display:

```text
internal location code
database ID
debug metadata
```

Mobile:

```text
bounded viewport-aware result panel
internal scrolling
no giant dropdown overflowing the modal
```

This searchable control supersedes the native `<select>` restriction from SPEC-06.4. The SPEC-06.4 **name-only label rule remains valid**.

---

# 23. Location search concurrency correctness

When a new query is scheduled:

```text
highlighted result index resets
old result set is not selectable
```

While the current query is loading:

```text
ArrowUp/ArrowDown/Enter cannot select stale hidden results
```

Request sequence protection remains latest-response-wins.

No polling.
No automatic search retry loops.

---

# 24. Login UI

Current login direction is the later visual-polish result, not the original SPEC-04 mock.

Canonical assets include:

```text
frontend/static/images/login-background.png
frontend/static/images/hotel-login-icon.svg
```

Login is dark branded with the background image and glass panel treatment.

The form remains:

```text
username
password
Remember me (username only)
Sign in
```

No:

```text
forgot password
alternative account login
email login
SSO
```

Earlier SPEC-04 / SPEC-04.1 visual details are historical where SPEC-05.1 / SPEC-05.2 changed them.

---

# 25. Desktop shell

Desktop sidebar:

```text
Tickets
Report     disabled/deferred
Settings   disabled/deferred
```

User/profile area remains in sidebar footer.

Ticket page:

```text
Open
Closed
New Request
```

Report and Settings are not implemented merely because they exist in navigation.

---

# 26. Ticket Detail / actions are NOT implemented yet

Next major intended product phase remains Ticket Detail Read.

Current list must not pretend these actions exist:

```text
Accept
Assign
Close
Chat
Checklist
```

Ticket Detail desktop direction remains:

```text
right-side integrated panel
does not cover the table
full-height allowed
```

Mobile direction remains:

```text
dedicated detail/chat screen
```

Do not implement fake click-through content or dead action controls.

---

# 27. Attachments product direction vs current implementation

Product direction supports:

```text
New Request image attachments
Ticket chat attachments
Staff Meal image
```

Database foundation includes ticket attachment metadata capability.

However, current `POST /api/v1/tickets` and current New Request frontend do **not** yet implement binary/image upload.

When implemented later:

```text
object storage for binary content
PostgreSQL stores metadata/references
MIME/type/size/authorization validation required
```

Do not store large binaries directly in PostgreSQL.

---

# 28. Staff Meal / Announcements / Report / Settings / Admin

Not currently implemented.

Staff Meal direction:

```text
one complete menu image uploaded by admin
not one image per weekday
```

Announcements:

```text
real admin-created persistent posts
pagination later
```

Report:

```text
deferred until explicit supervisor-provided UI/functionality
```

Settings:

```text
deferred
```

Admin UI:

```text
not yet designed
do not infer final admin screens from staff screenshots
```

---

# 29. Visual-reference rule

When the user supplies a screenshot/reference and asks to match it:

```text
follow the visible hierarchy
follow the content placement
follow the requested interaction
follow spacing/density intent
```

Do not change the requested design merely because another UI pattern seems more modern.

A later explicit user instruction overrides an older screenshot.

---

# 30. Error handling

Use the existing application error model.

Do not display raw backend error messages directly in the frontend.

Known safe create-ticket codes currently preserved by frontend:

```text
department_unavailable
location_unavailable
invalid_request
```

Frontend owns user-facing strings.

Unknown/malformed backend error payloads fall back to a safe generic message.

---

# 31. Logging/security

Do not log:

```text
passwords
password hashes
raw session tokens
auth cookies
secret keys
credentials
```

Useful context includes:

```text
request id
route
status
latency
user/ticket IDs where safe
database error category
```

---

# 32. Performance philosophy

Performance comes first from:

```text
correct SQL
correct indexes
small payloads
no N+1
short transactions
keyset pagination for growing ticket lists
bounded pool behavior
good frontend data flow
no unnecessary refetch
measurement
```

Do not optimize by adding abstractions.

The current pgx pool configuration default is approximately:

```text
max connections = 10
min connections = 1
acquire timeout = 5s
```

Configuration supports a bounded higher max; do not map request concurrency 1:1 to PostgreSQL connections.

---

# 33. Future high-concurrency performance program

After the functional application is substantially complete, perform a dedicated performance/hardening phase.

The user's desired stress target is:

```text
evaluate individual important functions/endpoints up to ~10,000 concurrent requests/users where meaningful
```

This is a **future test target**, not a statement that the current application already supports 10,000 concurrent requests.

For each important endpoint/flow, measure progressively:

```text
50
100
200
500
1,000
2,000
5,000
10,000
```

where technically meaningful.

Record:

```text
throughput/RPS
p50
p95
p99
error rate
CPU
RAM
PostgreSQL CPU/RAM
pgx pool acquire wait
active connections
query timing
query plans
lock waits
payload size
```

Then optimize the measured bottleneck.

Potential later optimizations may include:

```text
index/query improvements
pool tuning
payload reduction
request coalescing
cache only if justified
async work only if justified
object storage/CDN tuning
```

Do not add Redis or distributed architecture preemptively just to claim a 10k target.

---

# 34. Current verification philosophy

Backend feature completion should run, where applicable:

```bash
gofmt -d .
go vet ./...
go test ./... -count=1
go test -race ./... -count=1
go build ./...
```

Frontend:

```bash
bun install --frozen-lockfile
bun run check
bun test
bun run build
```

Repository:

```bash
git diff --cached --check
git status
```

Only claim commands actually executed.

Manual viewport checks must only be claimed if they were actually performed.

GitHub currently does not provide required CI/status checks for these commits; local verification reports must not be represented as GitHub CI evidence.

---

# 35. Current SPEC status

Canonical status:

```text
SPEC-01   ✅ CLOSED
SPEC-02   ✅ CLOSED
SPEC-03   ✅ CLOSED
SPEC-04   ✅ CLOSED
SPEC-04.1 ✅ CLOSED
SPEC-05   ✅ CLOSED
SPEC-05.1 ✅ CLOSED
SPEC-05.2 ✅ CLOSED
SPEC-06   ✅ CLOSED
SPEC-06.1 ✅ CLOSED
SPEC-06.2 ✅ CLOSED
SPEC-06.3 ✅ CLOSED
SPEC-06.4 ✅ CLOSED
SPEC-06.5 ✅ CLOSED
SPEC-06.6 ✅ CLOSED
SPEC-06.7 ✅ CLOSED
```

Verified SPEC-06.7 cleanup state:

```text
ROOM-8020 legacy generic row exists but inactive
ROOM-7309 legacy generic row exists but inactive
BWP-ROOM-8020 active
BWP-ROOM-7309 active
bwp_room_locations_integration_test.go uses a descriptive domain filename
all 564 BWP room rows remain unchanged
```

---

# 36. SPEC supersession map

Later rules override older conflicting clauses.

```text
SPEC-05
  ticket-list visual hierarchy later superseded by SPEC-06.1

SPEC-06.1
  Due/Priority placement later superseded by SPEC-06.5 and SPEC-06.6
  current priority = red content-sized title border, normal text

SPEC-06.3
  96 Villas dataset remains valid
  preservation of generic Room 8020 / Room 7309 as active is superseded by 06.7 cleanup

SPEC-06.4
  name-only visible location label remains valid
  native <select>/no-combobox constraint superseded by SPEC-06.5

SPEC-06.5
  BWP 155-row dataset remains valid
  date/time stacking remains valid
  default location search limit=10 is superseded by SPEC-06.6
  red/bold priority text is superseded by SPEC-06.6

SPEC-06.6
  current authoritative location-picker/search/priority visual behavior

SPEC-06.7
  current authoritative BWP room master data
  cleanup deactivates only the obsolete generic Room 8020 / Room 7309 placeholders
```

Agents must read the later SPEC when an older rule appears contradictory.

---

# 37. Documentation policy for closed SPECs

Do not rewrite closed SPECs as though the later feature existed at the time they were implemented.

Preferred documentation approach:

```text
keep historical acceptance criteria
+
add a short "Superseded / amended by" note at the top of the affected SPEC
```

This preserves implementation history while preventing future agents from following stale UI instructions.

Do not duplicate entire later SPECs into older files.

---

# 38. Scope discipline

Every SPEC:

```text
read context
read current SPEC
inspect source
change only scope
use real database/API
test happy path + errors
verify
report honestly
```

Do not use one feature SPEC as permission for broad refactoring.

---

# 39. Permanent prohibitions

Do not silently:

```text
change SvelteKit
change Go/Fiber
replace pgx/pgxpool
introduce ORM
switch sessions to JWT
add Redis
add microservices
change ticket statuses
invent mock data
invent a final Admin UI
implement deferred Report
rewrite applied migrations
create test-only env files
hardcode production domains/IPs
```

without explicit approval.

---

# 40. Next intended phase

After PROJECT_CONTEXT/spec supersession notes are synchronized:

```text
SPEC-07 — Ticket Detail Read
```

Do not bundle Accept/Assign/Close/Chat/Checklist unless the SPEC explicitly includes them.

---

# 41. Quick permanent summary

```text
PROJECT:
BWP SonaSea

FRONTEND:
SvelteKit + Svelte 5 + TypeScript + Bun

BACKEND:
Go 1.27 + Fiber v3

DATABASE:
PostgreSQL + pgx v5 + pgxpool

ARCHITECTURE:
modular monolith

AUTH:
server-side username/password sessions

TICKET STATUS:
pending / accepted / closed

OWNER:
first accepter (accepted_by)

ASSIGNMENT:
separate from status
multi-department + multi-user capable

TICKET LIST:
Requester / Location / Title / Description / Status / Owner / Created On / Due Date

PRIORITY UI:
normal Title text + red content-sized border
no Priority badge

CREATE TICKET:
department + optional location + title + optional description + boolean priority + optional due time
requester from auth principal
image upload not implemented yet

LOCATION:
PostgreSQL-backed searchable picker
full active list when no explicit limit
literal search
bounded scrollable UI
name-only display
real location ID submitted

MASTER DATA:
14 approved requestable departments
141 96 Villas rows
155 BWP area rows
564 BWP room rows

MOCK DATA:
forbidden for runtime application

MIGRATIONS:
0001..0017, 0019, 0020
0018 retired
no 0021 currently

PERFORMANCE:
measure first
future dedicated load-test phase up to ~10k concurrency where meaningful

NEXT:
SPEC-06.7 CLOSED
then SPEC-07 Ticket Detail Read
```
