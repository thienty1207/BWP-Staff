# SPEC-05 — Authenticated Shell + Tickets Read/List Foundation

> Project: BWP SonaSea
> Baseline: `main` at or after `ad25b2cc337e0b6430a3e6537fd4c90f8cb8673b`
> This SPEC starts the real authenticated staff UI.
> **Runtime mock data is forbidden.**

## 1. Read order

1. `AGENTS.md`
2. `Context-Spec-BWP-SonaSea/PROJECT_CONTEXT.md`
3. SPEC-01
4. SPEC-02
5. SPEC-03
6. SPEC-04
7. SPEC-04.1
8. this SPEC
9. current repository implementation

Current explicit user instruction and current SPEC override older assumptions.

## 2. Goal

Implement:

```text
authenticated app shell
+ desktop sidebar
+ mobile hamburger/drawer
+ GET /api/v1/tickets
+ Open / Closed ticket listing
+ real PostgreSQL data
+ bounded keyset pagination
```

Do not implement ticket creation/actions/detail/chat yet.

## 3. Absolute no-mock-data rule

Forbidden in runtime/product code:

```text
hard-coded ticket arrays
sample ticket JSON
fake requester/department/location/staff names
fake ticket counts
random/demo tickets
frontend fallback demo data
mock API success responses
development seed tickets created only to populate UI
```

If PostgreSQL has no tickets, show the real empty state.

Automated tests may create isolated deterministic records only inside temporary test schemas/test memory. Those fixtures must never become runtime/dev product data.

Do not expand the normal development seed with tickets.

## 4. Existing ticket schema

Use existing tables only.

Core fields include:

```text
tickets.id
tickets.requester_id
tickets.department_id
tickets.location_id
tickets.title
tickets.description
tickets.status
tickets.accepted_by
tickets.accepted_at
tickets.closed_by
tickets.closed_at
tickets.created_at
tickets.updated_at
tickets.priority
tickets.due_at
```

Assignments:

```text
ticket_assigned_departments
ticket_assigned_users
```

Statuses exactly:

```text
pending
accepted
closed
```

Never create `assigned` status.

## 5. Open / Closed meaning

```text
Open   = pending + accepted
Closed = closed
```

Open displays Pending/Accepted badges.

No separate Pending tab.
No My Tickets tab.

## 6. Backend package

Create the real read/list ticket feature, preferably:

```text
backend/client/tickets/
  handler.go
  service.go
  repository.go
  model.go
  tests...
```

Use:

```text
HTTP -> Handler -> Service -> Repository -> pgx/pgxpool -> PostgreSQL
```

No ORM or generic query framework.

## 7. Endpoint

Implement only:

```http
GET /api/v1/tickets
```

Require existing `RequireAuth`.

Do not implement create/detail/accept/assign/close/edit/delete endpoints in SPEC-05.

## 8. Read visibility

Do not invent RBAC/department filtering yet.

For SPEC-05, every active authenticated staff user may read the ticket list.

Do not filter by current department/requester/assignment/role unless authoritative requirements are explicitly changed.

## 9. Query parameters

Support:

```text
view=open|closed
limit=<integer>
before_created_at=<RFC3339>
before_id=<positive integer>
```

Rules:

```text
view default = open
limit default = 50
limit min = 1
limit max = 100
```

Cursor fields must be both present or both absent.

Invalid query -> safe `400 invalid_request`.

## 10. Ordering / pagination

Order:

```sql
created_at DESC, id DESC
```

Use keyset pagination, not large OFFSET pagination.

Next-page comparison should be equivalent to:

```sql
(created_at, id) < ($before_created_at, $before_id)
```

Fetch `limit + 1`, return at most `limit`, set `has_more`.

## 11. Response shape

Recommended:

```json
{
  "tickets": [
    {
      "id": 123,
      "title": "TV not working",
      "status": "pending",
      "priority": false,
      "due_at": null,
      "created_at": "2026-09-11T10:00:00Z",
      "updated_at": "2026-09-11T10:00:00Z",
      "requester": {
        "id": 10,
        "full_name": "..."
      },
      "department": {
        "id": 1,
        "code": "IT",
        "name": "IT Department"
      },
      "location": null,
      "accepted_by": null,
      "accepted_at": null,
      "assigned_departments": [],
      "assigned_users": [],
      "closed_at": null
    }
  ],
  "page": {
    "has_more": false,
    "next_before_created_at": null,
    "next_before_id": null
  }
}
```

Do not return chat/checklist/message history/audit data/full attachment data.

This is a list endpoint, not detail.

## 12. Assignment semantics

Use real rows from:

```text
ticket_assigned_departments
ticket_assigned_users
```

`tickets.department_id` remains original destination/family department and is not current assignment state.

## 13. No N+1

Do not query assignments once per ticket.

Acceptable boring pattern:

```text
query 1: base page + requester/department/location/accepted user
query 2: all assigned departments for page IDs
query 3: all assigned users for page IDs
```

A readable aggregate query is also acceptable.

Database query count must stay bounded per page.

## 14. SQL discipline

Use explicit parameterized SQL and explicit columns.

No `SELECT *`.
No ORM.
No user-provided SQL fragments.
Use request context.

## 15. Database migration

No migration should be needed.

Do not create `0021` merely for SPEC-05.

Do not rewrite 0001–0017, 0019, 0020. 0018 remains retired.

If a genuine blocking schema/index issue is discovered, stop and report it instead of silently changing schema.

## 16. Backend tests

Use existing permanent pattern:

```text
backend/.env
DATABASE_URL
unique temporary PostgreSQL schema
run migrations
insert isolated test records
assert
drop only temp schema
```

Never create `DATABASE_TEST_URL`, `APP_ENV=test`, `.env.test`, etc.

Cover at minimum:

- unauthenticated -> 401
- open returns only pending + accepted
- closed returns only closed
- newest-first ordering
- ID tie-breaker
- priority/due_at
- null location
- requester/department/location mapping
- assigned departments/users
- no assignment duplication
- limit
- has_more
- second page with no duplicate
- invalid view/limit/cursor -> 400

## 17. Authenticated root

Replace the temporary SPEC-04 authenticated identity card.

`/` becomes the real Tickets screen.

Keep `/login` behavior unchanged.

## 18. Desktop shell

Build a persistent left sidebar with:

```text
real BWP logo
Tickets
Report
Settings
profile/user area at bottom
theme control
logout
```

Rules:

- Tickets is active and functional.
- Report is visible but not implemented.
- Settings is visible but not implemented.
- Report/Settings must not navigate to fake pages.
- Use disabled/non-interactive semantics until their real SPECs exist.
- Sidebar may use subtle blur/translucency.

## 19. Profile area

Use real `/me` data.

Show concise real identity such as full name + department.

If `avatar_url` exists, use it.
If absent, derived initials from real `full_name` are allowed.

Never use a fake profile photo.

## 20. Tickets header

Use heading:

```text
Tickets
```

Tabs exactly:

```text
Open
Closed
```

Do not show:

```text
Dashboard
My Request
Pending tab
search header
fake totals
fake KPI cards
```

## 21. Desktop table

Render real API rows.

Recommended columns:

```text
ID
Request
Department
Location
Requester
Assignment
Status
Created
Due
```

Display DB ID as `#<id>`.

Do not invent a ticket-number scheme.

Show priority indicator only when `priority=true`.

## 22. Status display

Only:

```text
Pending
Accepted
Closed
```

No Assigned / In Progress / Resolved / Done.

Assignment and lifecycle status stay separate.

## 23. Assignment display

Use real assignment arrays.

Allowed derived empty label:

```text
Unassigned
```

Never fabricate names.

## 24. No actions yet

SPEC-05 is read-only.

Do not render fake/disabled future action buttons:

```text
Accept
Assign
Close
New Request
Edit
Delete
```

These appear only when real backend behavior exists.

## 25. No detail/chat yet

Do not implement:

```text
ticket detail panel
clickable fake row
chat
checklist
three-dot chat menu
WebSocket
attachment viewer
```

## 26. No Staff Meal / Announcements yet

Do not show placeholder or mock Staff Meal/Announcements.

They require real persisted APIs later.

## 27. Frontend ticket client

Add a focused module such as:

```text
frontend/src/lib/client/tickets/
  api.ts
  model.ts
```

Only implement current needs, e.g. `listTickets()`.

Use relative:

```text
/api/v1/tickets
```

with:

```ts
credentials: 'include'
```

No generic SDK.

## 28. Auth behavior

On `/`:

1. call `/me`;
2. `401` -> `/login`;
3. network/500 -> session Retry state, not fake logout;
4. after auth succeeds -> load Open tickets.

If ticket list returns 401 -> redirect `/login`.

If ticket list returns 500/network -> keep shell, show Retry.

## 29. Loading state

No fake rows containing fake business data.

Neutral skeleton shapes with no names/ticket values are allowed, or use a simple loading state.

## 30. Empty state

Real empty response must show:

```text
Open:   No open tickets.
Closed: No closed tickets.
```

Do not insert sample tickets.

## 31. Error state

Ticket list network/500:

```text
concise message
Retry
preserve current tab
keep authenticated shell
```

Never replace failures with demo data.

## 32. Tab behavior

Default = Open.

Switching:

```text
Open   -> view=open
Closed -> view=closed
```

Do not load all historical tickets upfront.

## 33. Load More

When `has_more=true`, show `Load more`.

Use returned cursor pair and append real rows.

Prevent concurrent duplicate loads.

No button when `has_more=false`.

No fake page/count numbers.

## 34. Ticket reload behavior

The normal Tickets header must not contain a manual Refresh control or an equivalent reload control.

Retry remains available only when session verification or ticket loading fails.

No aggressive polling.
No realtime/WebSocket yet.

## 35. Mobile shell

Use approved mobile model:

```text
hamburger
drawer/menu
logo inside drawer
Tickets
Report disabled
Settings disabled
profile
theme
logout
```

Important:

```text
do not duplicate logo in mobile header and drawer
```

Logo belongs inside the drawer.

## 36. Mobile tickets

Only Open/Closed.

Render responsive cards/list rows, not a wide table.

Show concise real data:

```text
#id
title
department/location
status
priority if true
created
assignment summary
```

## 37. Responsive verification

Must work at:

```text
360
390
430
768
1280/desktop
```

Desktop -> sidebar + table.
Mobile -> hamburger drawer + cards.

No horizontal page overflow.

## 38. Theme

Reuse existing `light`, `dark`, and `bwp-theme`.

No new theme system.

Dark = red/charcoal.
Light = white/blue.

## 39. No search

Do not add ticket search or search query params.

The approved header search remains removed.

## 40. No fake counts

Do not show tab counts unless a real count query/API exists.

SPEC-05 does not implement counts.

Use plain Open / Closed.

## 41. Do not expand dev seed

Do not add ticket examples to:

```text
backend/admin/
backend/cmd/seed_development/
```

No runtime demo tickets.

Empty dev database must produce empty state.

## 42. Test fixtures are not runtime mock data

Temporary test fixtures are allowed only for deterministic automated tests.

They must never:

- populate normal development DB;
- be imported by runtime product code;
- become frontend fallback data;
- become committed sample product JSON.

## 43. Performance

Required:

```text
bounded page size
keyset pagination
no N+1
DB-side filtering
explicit payload
do not load all tickets into browser
```

Do not add Redis.

## 44. Security

Endpoint requires auth.

Never expose passwords/session tokens/session hashes.

No JWT/Bearer.
No cookie logging.
Preserve request-ID and safe-error behavior.

## 45. Backend verification

From `backend/`:

```bash
gofmt -d .
go mod tidy
go vet ./...
go test ./... -count=1
go build ./...
```

Where supported:

```bash
go test -race ./... -count=1
```

Run real PostgreSQL integration tests against `DATABASE_URL` using temporary schema isolation.

Only report commands actually run.

## 46. Frontend verification

From `frontend/`:

```bash
bun install --frozen-lockfile
bun run check
bun run build
bun test
```

At minimum existing auth/branding tests and new ticket-client tests must pass.

Only report actual results.

## 47. Manual real-data verification

Use real backend + real PostgreSQL + real login.

Verify:

```text
/ -> Tickets shell
Open -> GET /api/v1/tickets?view=open
Closed -> GET /api/v1/tickets?view=closed
```

If real dev tickets exist, verify them.

If no tickets exist, verify correct empty state.

**Do not create fake dev tickets solely for screenshots/manual verification.**

## 48. Auth regression

Verify:

```text
unauthenticated / -> /login
login -> /
refresh / -> authenticated
logout -> /login
back/direct / after logout -> /login
```

## 49. Documentation

Commit SPEC at:

```text
Context-Spec-BWP-SonaSea/Spec/SPEC-05-authenticated-shell-tickets-list.md
```

Update PROJECT_CONTEXT current repository snapshot after implementation.

Update README only if run/API docs become stale.

## 50. Final diff review

Remove anything matching:

```text
runtime mock tickets
sample product JSON
dev ticket seed data
fake counts
fake departments/users/locations
fake Report/Settings pages
fake action buttons
detail/chat scope creep
fake Staff Meal/Announcements
migration/schema changes
new frontend env files
JWT/Bearer
N+1 ticket queries
unbounded SELECT
OFFSET-heavy pagination
SPEC-06+ work
```

## 51. Definition of Done

Complete only when:

- `/` is real Tickets shell;
- desktop sidebar exists;
- mobile hamburger/drawer exists;
- real BWP branding reused;
- real `/me` data powers profile;
- `GET /api/v1/tickets` exists and requires auth;
- Open = pending + accepted;
- Closed = closed;
- keyset pagination works;
- response uses real relational data;
- assignments come from assignment tables;
- no N+1;
- desktop table uses real API data;
- mobile cards use real API data;
- no Assigned status;
- no ticket actions/detail/chat;
- no Staff Meal/Announcements mock content;
- empty DB shows empty state;
- **no runtime mock ticket data exists**;
- **no development ticket seed data added**;
- no fake counts;
- dark/light work;
- auth regressions pass;
- no migration added by default;
- backend/frontend verification passes;
- responsive verification passes;
- docs are current;
- no later-SPEC scope creep remains.

## 52. Next phase

After SPEC-05, add ticket creation/detail/actions in later SPECs.

Do not implement them here.
