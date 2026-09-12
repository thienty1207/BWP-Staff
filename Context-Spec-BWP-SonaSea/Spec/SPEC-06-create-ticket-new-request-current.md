# SPEC-06 — Create Ticket / New Request + Targeted Hardening

> Project: **BWP SonaSea**
>
> Repository: `thienty1207/BWP-Staff`
>
> Expected baseline: `main` at or after `1a31ed0dd39e74137e45a474b527f9ed4a00aa19`
>
> Status before this SPEC:
>
> - SPEC-01 ✅ CLOSED
> - SPEC-02 ✅ CLOSED
> - SPEC-03 ✅ CLOSED
> - SPEC-04 ✅ CLOSED
> - SPEC-04.1 ✅ CLOSED
> - SPEC-05 ✅ CLOSED
> - SPEC-05.1 ✅ CLOSED
- SPEC-05.2 ✅ CLOSED
>
> This SPEC introduces the first real ticket write flow. It also includes three small, targeted hardening fixes discovered during the post-SPEC-05 code audit.

---

## 1. Goal

Build a real **New Request / Create Ticket** flow backed by PostgreSQL.

A signed-in staff user must be able to:

1. open **New Request**;
2. load active departments and locations from PostgreSQL;
3. enter the request;
4. create a real ticket through the Go backend;
5. have the requester derived from the authenticated session;
6. have the ticket created as `pending`;
7. return to the **Open** list and see the new server-backed ticket.

No runtime mock data is allowed.

The intended data flow is:

```text
New Request UI
    ↓
GET /api/v1/departments
GET /api/v1/locations
    ↓
POST /api/v1/tickets
    ↓
Fiber handler
    ↓
Ticket service
    ↓
Ticket repository
    ↓
PostgreSQL transaction
    ├── validate/lock active department
    ├── validate/lock optional active location
    ├── INSERT ticket
    └── INSERT ticket_activity(action='created')
    ↓
201 Created
    ↓
reload Open first page
    ↓
new real ticket is visible
```

---

## 2. Source-of-truth order

Before implementation, read in this order:

1. explicit current user instruction
2. this SPEC
3. `Context-Spec-BWP-SonaSea/PROJECT_CONTEXT.md`
4. SPEC-01 through SPEC-05.1
5. current implementation
6. assumptions only when unavoidable

Do not silently rewrite already-closed behavior.

---

## 3. Architecture remains locked

Keep the existing backend architecture:

```text
HTTP
→ Handler
→ Service
→ Repository
→ pgx
→ PostgreSQL
```

Keep:

- Go
- Fiber v3
- `context.Context`
- `pgx/v5`
- `pgxpool`
- explicit SQL
- PostgreSQL
- SvelteKit / Svelte 5
- TypeScript
- Bun

Do not introduce:

- ORM
- Redis
- queue
- microservice
- generic repository framework
- DI container
- reflection-heavy abstraction
- global mutable session map

This SPEC should remain a boring modular-monolith change.

---

# PART A — TARGETED HARDENING INCLUDED IN SPEC-06

These fixes were found during the codebase audit. They are small enough to include here and must not become a broad refactor.

---

## 4. Hardening A — Argon2 verifier parameter caps

Current password verification parses Argon2 PHC parameters from the stored database hash.

The verifier must reject unsafe/absurd parameters **before** calling `argon2.IDKey`.

The current generated policy is:

```text
memory      = 65536 KiB
iterations  = 3
parallelism = 4
key length  = 32 bytes
salt length = 16 bytes
```

Required hardening:

- reject memory above the supported maximum before allocation;
- reject iterations above the supported maximum;
- reject parallelism above the supported maximum;
- reject an unreasonable decoded salt length;
- reject an unreasonable decoded expected-key length;
- malformed hashes still return `false`;
- verifier must never panic for malformed PHC strings.

Keep the implementation explicit.

A reasonable supported cap is the current application policy for the expensive parameters:

```text
memory      <= 65536 KiB
iterations  <= 3
parallelism <= 4
key length  <= 32 bytes
```

Salt should also have a small explicit upper bound, e.g. `64` bytes.

The goal is to prevent a corrupted/tampered DB hash from causing unexpectedly huge CPU/RAM allocation during login.

Do not weaken normal Argon2 verification.

---

## 5. Hardening B — seed/login password size consistency

Normal login already rejects password input larger than `1024` bytes.

Development admin seeding must use the same maximum.

Create one clear shared password-input limit rather than allowing the two flows to drift.

Recommended direction:

```go
security.MaxPasswordBytes = 1024
```

Use that limit in:

- normal login request validation;
- development admin seed validation.

Required behavior:

```text
password length 1..1024 bytes
→ allowed

password >1024 bytes
→ rejected before password hashing / DB insertion
```

Do not alter existing seeded users.

Do not add another env file.

---

## 6. Hardening C — frontend RFC3339 timestamp contract

The current ticket frontend parser is broader than the Go backend because browser `Date.parse()` accepts formats outside RFC3339.

While the ticket client is being extended for creation, tighten timestamp validation.

The frontend ticket API parser must accept the API timestamp contract only:

```text
YYYY-MM-DDTHH:mm:ssZ
YYYY-MM-DDTHH:mm:ss.sssZ
YYYY-MM-DDTHH:mm:ss+07:00
YYYY-MM-DDTHH:mm:ss.sss+07:00
```

Fractional seconds may contain the normal RFC3339 fractional precision.

Validation should require:

1. RFC3339-shaped string;
2. parseable real timestamp.

Apply the same helper to:

- ticket `created_at`
- `updated_at`
- optional `due_at`
- optional `accepted_at`
- optional `closed_at`
- pagination `next_before_created_at`

Do not add a date library solely for this.

---

## 7. Hardening non-goals

Do **not** expand this hardening pass into unrelated work.

Specifically defer:

- theme first-paint/FOUC cleanup;
- CI workflow;
- GitHub branch protection;
- production login rate limiting;
- Redis/caching;
- broad performance refactor.

Those are not blockers for Create Ticket.

---

# PART B — LOOKUP APIs

## 8. New endpoint — departments

Add:

```http
GET /api/v1/departments
```

Authentication:

```text
valid active session required
```

Return **active departments only**.

Recommended response:

```json
{
  "departments": [
    {
      "id": 1,
      "code": "IT",
      "name": "IT Department"
    }
  ]
}
```

Rules:

- explicit columns;
- parameterized SQL where parameters exist;
- `WHERE is_active = TRUE`;
- stable order:

```sql
ORDER BY name ASC, id ASC
```

Do not return:

- description
- timestamps
- inactive departments
- fake departments

---

## 9. New endpoint — locations

Add:

```http
GET /api/v1/locations
```

Authentication:

```text
valid active session required
```

Return **active locations only**.

Recommended response:

```json
{
  "locations": [
    {
      "id": 1,
      "code": "LOB",
      "name": "Lobby"
    }
  ]
}
```

`locations.code` is nullable in the schema, therefore the API type must support:

```text
string | null
```

Rules:

- explicit columns;
- `WHERE is_active = TRUE`;
- stable order:

```sql
ORDER BY name ASC, id ASC
```

Do not return inactive locations or fake locations.

---

## 10. Lookup module boundary

Do not put generic department/location reads into a giant ticket handler.

Preferred backend boundary:

```text
backend/client/lookups/
    handler.go
    service.go
    repository.go
    model.go
```

The layer can remain thin and explicit.

Do not build a generic “reference data engine”.

Preferred frontend boundary:

```text
frontend/src/lib/client/lookups/
    api.ts
    model.ts
```

---

# PART C — CREATE TICKET BACKEND

## 11. New endpoint

Add:

```http
POST /api/v1/tickets
```

Authentication is mandatory.

The requester comes from the authenticated principal.

The client must **not control**:

- `requester_id`
- `status`
- `accepted_by`
- `accepted_at`
- assignment
- closure data
- created/updated timestamps

---

## 12. Create request contract

Recommended JSON:

```json
{
  "department_id": 2,
  "location_id": 15,
  "title": "Lobby TV is not displaying content",
  "description": "The horizontal lobby TV shows no content after restart.",
  "priority": false,
  "due_at": "2026-09-12T09:30:00+07:00"
}
```

Nullable/optional fields:

```text
location_id
description
due_at
```

`priority` defaults to `false` when omitted.

---

## 13. Create validation

### department_id

Required:

```text
positive integer
```

The referenced department must exist and be active.

### location_id

Optional.

If present:

```text
positive integer
```

The referenced location must exist and be active.

### title

Required.

Rules:

- trim outer whitespace;
- must not be blank;
- maximum `255` Unicode characters/runes;
- store trimmed value.

This matches the existing `VARCHAR(255)` schema and nonblank DB constraint.

### description

Optional.

Rules:

- trim outer whitespace;
- empty after trimming becomes `NULL`;
- maximum `5000` Unicode characters/runes.

The application-level bound prevents accidentally huge product requests while leaving the existing `TEXT` schema unchanged.

### priority

Boolean.

Default:

```text
false
```

### due_at

Optional RFC3339 timestamp.

Do not invent a business rule requiring it to be in the future unless the user later requests that policy.

PostgreSQL stores it as `TIMESTAMPTZ`.

---

## 14. Mass-assignment safety

The server must derive requester/status/action state itself.

A request must not be able to spoof:

```json
{
  "requester_id": 999,
  "status": "closed",
  "accepted_by": 999,
  "closed_by": 999
}
```

Preferred behavior:

- decode only the explicit create-request DTO;
- reject unknown JSON fields for this create endpoint.

Keep strict decoding local and simple.

Do not create a generic reflection-based validator.

---

## 15. Create transaction

Ticket creation must be atomic.

Use one PostgreSQL transaction for:

1. validate/lock active department;
2. validate/lock optional active location;
3. insert ticket;
4. insert ticket activity;
5. commit.

Suggested reference validation:

```sql
SELECT id, code, name
FROM departments
WHERE id = $1
  AND is_active = TRUE
FOR SHARE
```

For location:

```sql
SELECT id, code, name
FROM locations
WHERE id = $1
  AND is_active = TRUE
FOR SHARE
```

This avoids a race where a reference becomes inactive between validation and insertion.

Do not use a serializable transaction unless evidence requires it.

---

## 16. Ticket insert invariants

New tickets must always begin as:

```text
status = pending
accepted_by = NULL
accepted_at = NULL
closed_by = NULL
closed_at = NULL
assigned departments = []
assigned users = []
```

Requester:

```text
authenticated principal user ID
```

Destination department:

```text
department_id from validated create request
```

Remember:

`tickets.department_id` is the original destination/family department, not assignment state.

---

## 17. Ticket activity

In the same transaction, create:

```text
ticket_activity.action = "created"
ticket_activity.actor_user_id = authenticated requester
```

Metadata may remain `NULL` for SPEC-06.

If activity insertion fails, rollback ticket creation.

No partial ticket creation.

---

## 18. Create response

Return:

```http
201 Created
```

Recommended response shape:

```json
{
  "ticket": {
    "id": 123,
    "title": "Lobby TV is not displaying content",
    "status": "pending",
    "priority": false,
    "due_at": "2026-09-12T02:30:00Z",
    "created_at": "2026-09-12T01:00:00Z",
    "updated_at": "2026-09-12T01:00:00Z",
    "requester": {
      "id": 4,
      "full_name": "Ho Thien Ty"
    },
    "department": {
      "id": 2,
      "code": "FO",
      "name": "Front Office"
    },
    "location": {
      "id": 15,
      "code": "LOB",
      "name": "Lobby"
    },
    "accepted_by": null,
    "accepted_at": null,
    "assigned_departments": [],
    "assigned_users": [],
    "closed_at": null
  }
}
```

Keep this compatible with the existing ticket-summary representation where practical.

Do not return the password/session/token material.

---

## 19. Backend create errors

Required categories:

### 400 — malformed/invalid request

Examples:

- invalid JSON;
- unknown create field;
- department ID <= 0;
- location ID <= 0;
- blank title;
- title too long;
- description too long;
- invalid due timestamp.

Use safe public messaging.

### 400 — department unavailable

Missing or inactive department.

Recommended code:

```text
department_unavailable
```

### 400 — location unavailable

Missing or inactive location.

Recommended code:

```text
location_unavailable
```

### 401

No valid session:

```text
unauthenticated
```

### 500

Unexpected DB/internal error:

```text
internal_server_error
```

The central backend error handler must continue to hide internal details.

---

## 20. No automatic POST retry

A network failure during a write can be ambiguous:

```text
server may have committed
but client may have lost the response
```

SPEC-06 does **not** add a database idempotency-key subsystem.

Therefore:

- do not automatically retry `POST /tickets`;
- do not use an automatic request library that retries writes;
- on retryable/ambiguous create failure, keep form values and show a clear generic error;
- do not fabricate success.

A future idempotency feature may be added if production behavior requires it.

---

# PART D — FRONTEND NEW REQUEST

## 21. New Request button

Add a real:

```text
New Request
```

control to the Tickets page.

Desktop:

- place it in the Tickets page header where it fits naturally;
- this is a real product action, unlike the removed manual Refresh control.

Mobile:

- keep it clearly accessible without duplicating unrelated navigation controls.

Do not reintroduce Refresh.

---

## 22. Component boundary

Do not dump the entire form into the already-large root page.

Create a focused component, e.g.:

```text
frontend/src/lib/components/NewRequestDialog.svelte
```

The root page should own orchestration:

- open/close;
- successful-create refresh;
- auth redirect.

The form component should own:

- fields;
- lookup loading;
- validation presentation;
- submit state.

Do not refactor the entire existing Tickets shell during this SPEC.

---

## 23. Desktop form presentation

Desktop New Request should be a clean temporary modal/dialog.

It may cover part of the shell because creation is a temporary action.

This restriction from the product remains specific to future persistent ticket-detail/chat behavior:

```text
ticket detail/chat must not cover the ticket table
```

Do not pre-build the future chat panel here.

---

## 24. Mobile form presentation

On narrow screens, the same New Request UI should become:

```text
full-width / near-full-screen sheet
```

Requirements:

- no horizontal overflow;
- all controls accessible;
- submit button visible/reachable;
- keyboard-friendly;
- scrolling works for small-height devices.

Do not create a separate duplicated form implementation.

---

## 25. Form fields

Required form:

### Department

Required `<select>`.

Data source:

```http
GET /api/v1/departments
```

No hard-coded options.

### Location

Optional `<select>`.

Data source:

```http
GET /api/v1/locations
```

Include:

```text
No location
```

as a real null choice.

### Request

Required title input.

### Description

Optional textarea.

### Priority

Boolean checkbox/toggle.

Default:

```text
false
```

### Due time

Optional `datetime-local`.

Before API submission:

- convert the browser-local value into a valid RFC3339/ISO timestamp;
- send `null` when empty.

### Actions

```text
Cancel
Create request
```

---

## 26. Lookup loading behavior

Do not fetch lookups continuously.

Preferred behavior:

```text
open New Request
→ lazy-load departments + locations
→ reuse them while that form instance remains open
```

No:

- polling;
- setInterval;
- realtime subscription;
- Redis cache.

If lookup load fails:

- keep dialog open;
- show Retry for lookup failure;
- do not show fake options.

If 401:

- redirect to `/login`.

If no active departments exist:

- show a real empty state;
- disable Create request.

If no active locations exist:

- Location remains optional with only `No location`.

---

## 27. Client validation

Frontend validation is for UX only.

Backend remains authoritative.

Before submit, validate:

- department selected;
- title nonblank;
- title <=255 characters;
- description <=5000 characters;
- optional due time convertible to valid timestamp.

Do not trust frontend validation for security.

---

## 28. Create success behavior

On successful `201`:

```text
close New Request
→ set active view to Open
→ reload first Open page from server
→ new ticket appears from real PostgreSQL data
```

Do not manually fabricate/prepend a fake local ticket.

The extra GET is intentional:

- simple;
- server-authoritative;
- keeps pagination state correct;
- avoids client-side list mutation complexity.

At current hotel workload, this is preferable to clever optimistic cache mutation.

---

## 29. Create error behavior

### 400 validation/reference error

Keep form open.

Keep user-entered values.

Show a safe actionable message.

### 401

Redirect to `/login`.

### network/500

Keep form open.

Keep values.

Show:

```text
Unable to create the request right now. Please check your connection and try again.
```

Do not auto-submit again.

Do not clear the form.

---

## 30. Request concurrency

While create is in flight:

- disable submit;
- prevent double submission;
- show clear loading text such as `Creating…`.

Do not allow duplicate click submissions.

Do not create a global request manager.

---

## 31. Interaction with ticket-list requests

Avoid awkward create/list races.

The New Request control may be disabled while the ticket list is in a list/pagination request.

While the New Request modal is active, the user should not be able to trigger Load More behind it.

After successful creation, the Open first-page reload must run from a clean list-request state.

Do not rewrite the current ticket sequence mechanism unless required for a concrete race.

---

# PART E — PERFORMANCE RULES

## 32. Performance target

No new cache layer is required.

The create path is low-frequency compared with read traffic.

Correctness/readability take priority over shaving one DB round trip.

Expected create transaction is small and bounded.

Lookup queries operate over reference tables.

---

## 33. SQL rules

Keep:

- explicit column lists;
- parameterized values;
- request context;
- bounded queries;
- no N+1;
- no `SELECT *`;
- no dynamic user-provided SQL fragments.

No migration is expected.

The current schema already contains:

- `tickets.priority`
- `tickets.due_at`
- relational assignment tables
- `ticket_activity`
- departments/locations with `is_active`

Do not create migration `0021` unless an actual blocking schema deficiency is demonstrated and reported first.

---

# PART F — NO MOCK DATA

## 34. Permanent no-mock rule

Runtime/product mock data is forbidden.

Do not add:

- fake ticket arrays;
- fake departments;
- fake locations;
- fake users;
- fake counts;
- fake create success;
- hard-coded dropdown data;
- demo request rows;
- development seed tickets.

An empty DB must remain genuinely empty.

Automated tests may use deterministic fixtures only inside isolated temporary PostgreSQL schemas.

Frontend unit tests may stub `fetch` only inside tests.

---

# PART G — FILES

## 35. Backend files likely to change

Expected:

```text
backend/shared/security/password.go
backend/shared/security/password_test.go

backend/admin/seed.go
backend/admin/...test...

backend/client/auth/handler.go
backend/client/auth/model.go
backend/client/auth/...test...

backend/client/lookups/handler.go
backend/client/lookups/service.go
backend/client/lookups/repository.go
backend/client/lookups/model.go

backend/client/tickets/handler.go
backend/client/tickets/service.go
backend/client/tickets/repository.go
backend/client/tickets/model.go
backend/client/tickets/...tests...

backend/app/app.go
```

Exact test file names may follow the current repository convention.

---

## 36. Frontend files likely to change

Expected:

```text
frontend/src/lib/client/lookups/api.ts
frontend/src/lib/client/lookups/model.ts

frontend/src/lib/client/tickets/api.ts
frontend/src/lib/client/tickets/model.ts

frontend/src/lib/components/NewRequestDialog.svelte

frontend/src/routes/+page.svelte
frontend/src/lib/styles/app.css

frontend/tests/...
```

Do not redesign `/login`.

---

# PART H — TESTS

## 37. Backend hardening tests

Required:

### Argon2

- valid existing generated hash still verifies;
- wrong password fails;
- malformed PHC fails;
- memory above cap fails without huge allocation;
- iterations above cap fails;
- parallelism above cap fails;
- oversized decoded key/salt fails safely.

### Seed password

- <=1024-byte password can proceed to normal seed behavior;
- >1024-byte password is rejected before hashing/insertion.

---

## 38. Lookup API integration tests

Using isolated temp schema:

- unauthenticated departments -> 401;
- unauthenticated locations -> 401;
- only active departments returned;
- only active locations returned;
- stable ordering;
- nullable location code serialized correctly;
- no extra sensitive fields.

---

## 39. Create API integration tests

Using isolated temp schema:

1. unauthenticated POST -> 401;
2. valid request -> 201;
3. DB ticket row exists;
4. requester is authenticated user;
5. status is `pending`;
6. priority persists;
7. due_at persists;
8. description persists;
9. optional location null works;
10. active location works;
11. inactive/missing department -> 400;
12. inactive/missing location -> 400;
13. blank title -> 400;
14. >255-char title -> 400;
15. >5000-char description -> 400;
16. invalid IDs -> 400;
17. unknown/forbidden JSON fields are rejected;
18. ticket activity `created` exists with requester actor;
19. ticket and activity rollback together on transaction failure where practical to exercise;
20. response contains no password/session fields.

Do not use normal development rows for these tests.

---

## 40. Frontend tests

Required focused tests:

- departments response parser;
- locations response parser including nullable code;
- create request mapping;
- create 201 parsing;
- 401 mapping;
- 400 mapping;
- retryable 5xx/network mapping;
- strict RFC3339 acceptance/rejection;
- no runtime mock lookup arrays;
- New Request component does not persist fake data.

Do not add a large component test framework solely for this SPEC.

---

# PART I — MANUAL VERIFICATION

## 41. Manual browser verification

Using real frontend/backend:

- login still works;
- Tickets still loads;
- Open/Closed still works;
- New Request opens;
- active departments come from API;
- active locations come from API;
- form responsive at 360/390/430/768/1280;
- show no horizontal overflow;
- Cancel works;
- client validation works;
- logout still works;
- no Refresh control returns.

Do not pollute the normal development DB with fake tickets only to demonstrate screenshots.

If a legitimate test submission is not appropriate, rely on the isolated integration tests for DB-create proof and state that manual final submission was intentionally not performed.

---

# PART J — NON-GOALS

## 42. Explicitly out of scope

Do not implement:

- ticket detail page/panel;
- Accept;
- Assign;
- Close;
- chat;
- checklist;
- attachment upload;
- object storage;
- Staff Meal;
- Announcements;
- Report;
- Settings;
- admin UI;
- notifications;
- WebSocket;
- polling;
- Redis;
- idempotency-key database subsystem;
- CI/branch protection;
- production rate limiter;
- broad theme refactor.

### Attachments note

The schema already contains `ticket_attachments`, but SPEC-06 intentionally does **not** implement file upload.

Upload requires a storage contract and should be implemented with the later attachment/chat work instead of silently inventing infrastructure here.

---

# PART K — DEFINITION OF DONE

## 43. Completion criteria

SPEC-06 is complete only when:

- hardening A/B/C is implemented and tested;
- `GET /api/v1/departments` works from PostgreSQL;
- `GET /api/v1/locations` works from PostgreSQL;
- only active lookup rows are returned;
- `POST /api/v1/tickets` works;
- requester comes from session;
- new status is always `pending`;
- ticket + `created` activity are atomic;
- inactive references are rejected;
- New Request uses real lookups;
- no runtime mock data exists;
- frontend prevents duplicate click submission;
- successful creation reloads real Open tickets;
- errors preserve form values;
- 401 redirects to login;
- existing ticket pagination still works;
- auth regression remains green;
- frontend responsive verification passes;
- backend/frontend tests pass;
- no migration was added unless a genuine blocker was reported first;
- no SPEC-07+ functionality was implemented.

---

## 44. Verification commands

Backend:

```bash
cd backend
gofmt -w <changed-go-files>
go vet ./...
go test ./...
go build ./...
```

Frontend:

```bash
cd frontend
bun install --frozen-lockfile
bun run check
bun test
bun run build
```

Repository:

```bash
git diff --check
git status
```

Only report commands actually executed.

---

## 45. Suggested commit

```text
feat: implement SPEC-06 create ticket flow
```

Do not rewrite closed SPEC history.

---

## 46. Next phase

After SPEC-06 closes:

```text
SPEC-07 — Ticket Detail Read
```

Ticket actions such as Accept/Assign/Close remain later SPECs.
