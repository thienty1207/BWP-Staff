# SPEC-07 — Ticket Detail Read

> Project: **BWP SonaSea**
>
> Repository: `thienty1207/BWP-Staff`
>
> Baseline: `main` at or after `708fc80a00859bcdaf96231f7af5d7d7644a5066`
>
> Closed before this SPEC: SPEC-01 through SPEC-06.7.
>
> Status while implementing: **OPEN / NOT CLOSED**
>
> This SPEC is intentionally **read-only**. It adds real Ticket Detail retrieval and responsive Ticket Detail presentation. It does **not** implement Accept, Assign, Close, Chat, Checklist, attachments, notifications, realtime, or any ticket mutation.

---

# 1. Goal

Implement the first real Ticket Detail vertical slice:

```text
Ticket List
    ↓
select a real ticket
    ↓
GET /api/v1/tickets/:id
    ↓
PostgreSQL
    ↓
Ticket Detail UI
```

After SPEC-07, an authenticated staff user can open a ticket from the existing Open/Closed list and inspect its complete currently-supported data.

Runtime data flow must remain:

```text
PostgreSQL
→ Fiber
→ SvelteKit
```

No runtime mock data.

---

# 2. User-visible result

Desktop:

```text
Ticket List | Ticket Detail
```

The detail panel is on the right and **must not cover the table**.

Mobile / narrow tablet:

```text
Ticket List
→ Ticket Detail
→ Back to Tickets
```

The dedicated detail view replaces the list content until Back is pressed.

The user can inspect:

```text
Title
full Description
Status
Request Department
Location
Requester
Owner
Assigned Departments
Assigned Users
Created
Due
Accepted
Closed
Updated
```

No action workflow is introduced in this SPEC.

---

# 3. Locked scope

## In scope

```text
GET /api/v1/tickets/:id
positive-int64 route validation
404 ticket_not_found
real PostgreSQL detail query
existing Ticket model reuse
frontend getTicket(id)
strict response validation
desktop side-by-side detail layout
mobile/narrow-tablet dedicated detail view
loading
retryable error
not-found state
401 handling
stale-response protection
selection/close/back behavior
Open/Closed interaction
New Request interaction
accessibility
tests
responsive verification
```

## Out of scope

```text
Accept
Assign
Close
Chat
messages
Checklist
attachments / image upload
ticket activity timeline
notifications
WebSocket
polling
Report
Settings
Admin UI
new migration
schema change
new master data
Redis/cache
performance infrastructure
```

Do not add placeholders or disabled fake controls for out-of-scope features.

---

# 4. Existing product rules remain authoritative

Ticket statuses remain exactly:

```text
pending
accepted
closed
```

There is no `assigned` ticket status.

Assignment is separate from status.

Owner remains:

```text
accepted_by
```

meaning the first accepter.

`tickets.department_id` remains the original request destination/family.

Open contains:

```text
pending
accepted
```

Closed contains:

```text
closed
```

Do not create a separate Pending tab.

---

# 5. Backend endpoint

Add authenticated:

```http
GET /api/v1/tickets/:id
```

Access rule for SPEC-07:

```text
the same authenticated active staff population that can read the shared ticket list
may read Ticket Detail
```

Do not invent department-only or requester-only visibility.

Use the existing authentication middleware.

Authentication happens before the handler's detail logic. Therefore:

```text
unauthenticated request → 401
```

even if the supplied path ID would otherwise be malformed.

---

# 6. Path ID validation

For an authenticated request, `:id` must parse as a positive signed 64-bit integer.

Reject:

```text
abc
1.5
0
-1
9223372036854775808
other int64 overflow
```

with:

```http
400 Bad Request
```

Existing error envelope:

```json
{
  "error": {
    "code": "invalid_request",
    "message": "Invalid request"
  }
}
```

Do not rely on PostgreSQL casts to validate route parameters.

Do not treat malformed IDs as 404.

---

# 7. Not found

A syntactically valid positive ID that has no ticket:

```http
404 Not Found
```

Response:

```json
{
  "error": {
    "code": "ticket_not_found",
    "message": "Ticket not found"
  }
}
```

Do not return:

```text
200 + null
200 + {}
500
```

Use an explicit repository/service not-found distinction such as:

```go
ErrTicketNotFound
```

---

# 8. Success contract

Existing ticket:

```http
200 OK
```

Envelope:

```json
{
  "ticket": {
    "...": "existing explicit Ticket representation"
  }
}
```

Current detail payload must include:

```text
id
title
description
status
priority
due_at
created_at
updated_at
requester
department
location
accepted_by
accepted_at
assigned_departments
assigned_users
closed_at
```

Use existing Ticket/summary structures where practical.

Do not expose:

```text
password hash
session/token data
auth cookie
private credential fields
raw database rows
audit internals
```

---

# 9. Null/empty semantics

API:

```text
location absent            → null
description absent         → null
accepted_by absent         → null
accepted_at absent         → null
closed_at absent           → null
assigned_departments empty → []
assigned_users empty       → []
```

Never return assignment arrays as `null`.

UI:

```text
null/empty optional value → —
empty assignment array    → —
```

The detail description must render full plain text.

Do not parse description as HTML.

---

# 10. Historical references remain readable

Ticket Detail is historical data.

A ticket must remain readable when a referenced row has later become inactive.

Do **not** add `is_active = TRUE` filters to ticket-detail joins for:

```text
requester
request department
location
accepted user
assigned department
assigned user
```

where the schema/reference still exists.

The existing list behavior is the semantic baseline.

Inactive master/reference data must not turn an existing ticket into a false 404.

---

# 11. Repository behavior

Add a clear method, for example:

```go
FindByID(ctx context.Context, id int64) (Ticket, error)
```

Required query strategy:

```text
1 indexed ticket-by-primary-key query
+
bounded assignment queries
```

Reusing the existing assignment loaders with a one-ticket slice is acceptable.

Do not:

```text
call List() and search the returned page
load the complete ticket table
load all assignments
load all users/departments/locations
filter application-wide data in Go
introduce ORM
introduce generic repository framework
```

A small local row-scan/helper extraction is allowed only if it clearly reduces drift between list/detail SQL without making the code harder to read.

---

# 12. Query correctness guard

The detail base query must return the same core semantics as the list query:

```text
Requester = original requester
Department = original request destination
Location = persisted ticket location
Owner = accepted_by
Accepted At = tickets.accepted_at
Closed At = tickets.closed_at
```

Do not accidentally redefine Owner using `assigned_users`.

Do not accidentally replace original Department using assigned departments.

Do not require active destination/location/user rows.

---

# 13. Stable assignment ordering

Current list assignment loaders order assignments by assignment row ID.

Ticket Detail must preserve deterministic ordering.

Required:

```text
assigned_departments → stable insertion/assignment order
assigned_users       → stable insertion/assignment order
```

Tests must not depend on PostgreSQL's unspecified row order.

---

# 14. Service behavior

Add an explicit read-detail service method.

Service responsibilities:

```text
validate configured repository
call repository
preserve ErrTicketNotFound
return real Ticket
```

No mutation.

No transaction is required merely to read one ticket.

Do not add write locking for Ticket Detail.

---

# 15. Handler behavior

Register:

```text
GET /tickets/:id
```

Handler:

```text
parse + validate path ID
call service
map ErrTicketNotFound → 404 ticket_not_found
return {"ticket": ticket}
```

Use current app error handling.

Do not duplicate auth/session validation already handled by middleware.

---

# 16. No migration

SPEC-07 uses the existing schema.

Expected:

```text
active migrations: 0001..0017, 0019, 0020
0018 retired
no 0021
```

Do not add:

```text
table
column
enum
index
migration
fixture/master data
```

If implementation appears to need schema work:

```text
STOP and report the reason before changing database files.
```

---

# 17. Frontend API

Add:

```ts
getTicket(id: number): Promise<TicketSummary>
```

or a semantically equivalent name.

Request:

```text
GET /api/v1/tickets/:id
credentials: include
```

Reuse the existing strict ticket payload parser.

Do not introduce a weaker second parser that accepts malformed detail responses.

Expected mapping:

```text
200 valid payload   → TicketSummary
400                 → invalid_input
401                 → unauthenticated
404 ticket_not_found→ not_found
network error       → retryable
other HTTP error    → retryable
malformed 200 JSON  → retryable
```

Raw backend error text must not be rendered to users.

---

# 18. Frontend error type — known type trap

Current ticket client error code typing was originally designed around Create Ticket errors.

Do not force `ticket_not_found` into a type that only accepts:

```text
department_unavailable
location_unavailable
invalid_request
```

Use a clear compatible model, for example:

```text
CreateTicketErrorCode
TicketDetailErrorCode
or a broader TicketApiErrorCode union
```

Keep Create Ticket parsing behavior unchanged.

A refactor here must not weaken or regress existing Create Ticket error handling.

---

# 19. Separate list state from detail state

The existing page already has list-specific state such as:

```text
ticketRequestInFlight
ticketRequestSequence
ticketState
tickets
page
```

**Do not reuse those list request controls for Ticket Detail.**

Ticket Detail requires separate state, for example:

```text
selectedTicketID
detailState
detailTicket
detailErrorMessage
detailRequestSequence
or detailAbortController
```

Why:

```text
detail loading must not block Load More
detail loading must not corrupt list request sequencing
detail loading must not disable Open/Closed unnecessarily
detail loading must not disable New Request merely because a GET detail is running
```

This separation is a mandatory anti-regression rule.

---

# 20. Latest selection wins

Rapid selection must be safe.

Example:

```text
select A
select B
B response arrives
A response arrives late
```

Final UI must still show B.

Closing detail while a request is in flight:

```text
late response must not reopen detail
```

Switching Open/Closed while detail is in flight:

```text
late response must be ignored
```

Opening New Request while detail is in flight:

```text
detail is cleared
late response must be ignored
```

Use:

```text
dedicated request sequence
or AbortController
```

No new dependency.

---

# 21. Prevent stale-content flash

When selecting B after A:

```text
do not leave A's content visually presented as B
```

Immediately move detail to a loading state for B.

Allowed:

```text
panel remains open
heading/loading skeleton/status changes
```

Forbidden:

```text
A data remains visible under B selection until B arrives
```

This prevents users from acting on the wrong ticket in later SPECs.

---

# 22. Ticket selection semantics

Do not add an Action column merely to open detail.

There must be at least one real semantic, keyboard-focusable control per ticket.

Recommended:

```text
Title rendered as a button-like text control
```

Pointer row/card click may additionally open detail.

If both row/card click and nested Title control are implemented:

```text
one user click must produce only ONE detail request
```

Avoid event-bubbling double fetches.

Do not put conflicting nested interactive controls inside an invalid click target.

---

# 23. Desktop layout

Desktop contract:

> Ticket Detail appears on the right and **must not cover the ticket table**.

Use sibling layout, not overlay:

```text
Tickets workspace
├── Ticket List region
└── Ticket Detail region
```

Recommended sizing behavior:

```text
list column: minmax(0, 1fr)
detail column: clamp(20rem, 28vw, 26rem)
gap: existing spacing scale
```

The exact values can be adjusted to fit existing CSS, but:

```text
panel must not push outside viewport
list container must keep min-width: 0
existing table may retain its own horizontal scroll
```

Do not use:

```text
position: fixed over table
modal backdrop
absolute overlay covering rows
```

The detail region may be sticky and independently scrollable if needed.

---

# 24. Desktop 1366px regression guard

The app already has a fixed desktop sidebar and a ticket table with a large intrinsic width.

At `1366×768`:

```text
side panel must still not overlay the list
page must not cause body-level horizontal overflow
table-region horizontal scrolling is acceptable
detail region must remain usable
```

Do not "solve" the split view by shrinking fonts/data until unreadable.

Do not globally change the approved table column contract.

---

# 25. Responsive breakpoint guard

Current application switches to mobile navigation/cards around the existing responsive breakpoint.

SPEC-07 may introduce one additional layout breakpoint for detail if needed.

Recommended behavior:

```text
wide desktop (including 1366px) → list + right detail panel
narrow tablet / mobile          → dedicated detail view
```

Do not use user-agent/device detection.

Do not run JavaScript width polling.

Use CSS/media-query responsive layout.

---

# 26. Mobile / narrow-tablet detail

On narrow screens:

```text
Ticket List
→ select
→ full-width Ticket Detail content
```

Required:

```text
Back to Tickets visible
no horizontal scrolling
no desktop split view squeezed into narrow width
sidebar/mobile navigation behavior unchanged
```

Back:

```text
returns to already-loaded list
preserves Open/Closed tab
preserves loaded pagination data
does not refetch solely because Back was pressed
```

A new SvelteKit route is not required.

Do not perform a large routing refactor solely for SPEC-07.

---

# 27. Detail content hierarchy

Use existing visual language.

Suggested structure:

```text
Ticket Detail
Status

Title
Description

Request information
- Department
- Location
- Requester
- Owner

Assignment
- Assigned Departments
- Assigned Users

Timing
- Created
- Due
- Accepted
- Closed
- Updated
```

No chat composer.

No checklist.

No action buttons.

No ticket activity timeline.

---

# 28. Priority visual contract

Current canonical priority rule remains:

```text
priority=true:
normal Title text color
thin red/danger content-sized border
small radius
tight padding
natural wrapping

priority=false:
normal Title
```

Do not show a Priority badge.

Do not make the title red.

Do not stretch the red border to full panel width for a short title.

---

# 29. Timestamp consistency

Ticket Detail timestamps must use the same user-facing date/time style as the existing Ticket List.

Do not introduce a second visibly inconsistent format.

Avoid:

```text
raw RFC3339 strings in UI
UTC Z strings shown directly
mixed 24h/12h formats between list and detail
```

A small shared formatting helper is acceptable if needed.

Do not change timezone semantics of the existing list as part of this SPEC.

---

# 30. Description correctness

Detail description:

```text
full real content
plain text
normal wrapping
overflow-wrap where needed
```

Do not:

```text
reuse 2-line list clamp
render HTML
use {@html}
truncate long detail description
cause horizontal overflow
```

Null/blank displays `—`.

---

# 31. Assignments

Display actual:

```text
assigned_departments
assigned_users
```

Do not fabricate assignment data.

Do not treat an empty assignment array as an error.

Do not convert assignment state into Status or Owner.

---

# 32. Loading behavior

Desktop:

```text
list remains visible
detail region shows loading
```

Mobile:

```text
detail view shows loading
Back remains available
```

Do not blank the entire authenticated shell.

Loading UI must not contain stale previous-ticket values.

---

# 33. Retryable error behavior

On retryable failure:

```text
keep selected ticket ID
show safe inline message
show Retry
```

Retry performs only:

```text
GET /api/v1/tickets/:id
```

Do not:

```text
reload the whole browser page
reload all tickets
loop automatic retries
```

---

# 34. 404 frontend state

For `ticket_not_found`:

```text
Ticket not found.
```

Provide:

```text
Close on desktop
Back to Tickets on narrow screens
```

Do not convert 404 into generic "service unavailable".

Do not fake a ticket from list-summary data.

---

# 35. 401 behavior

If detail GET returns 401:

```text
invalidate/clear protected detail state
follow existing authenticated app behavior
goto /login
```

Do not leave previously loaded protected detail visible after session invalidation.

---

# 36. Open / Closed interaction

Switching tabs must:

```text
clear selected detail
invalidate in-flight detail GET
then continue existing list-switch behavior
```

Do not let detail from the old tab remain visible beside the new tab.

Do not make detail state mutate `activeView`.

---

# 37. New Request interaction

Opening New Request must:

```text
clear/close Ticket Detail
invalidate any detail GET
open the existing New Request dialog
```

Do not alter New Request business logic.

Do not make a detail GET set the existing `ticketRequestInFlight`, because that would incorrectly prevent New Request.

---

# 38. Load More interaction

Opening detail must not break existing pagination.

While detail is open:

```text
Load More may continue to operate
existing page cursor remains authoritative
selected ticket detail remains selected unless explicitly cleared
```

A Load More request and a detail GET must not invalidate each other's sequence/state.

This is why detail and list request sequencing must be separate.

---

# 39. List data is summary, detail API is authoritative

Do not simply display the already-loaded list object as "Ticket Detail" and skip the API request.

Opening detail must call:

```text
GET /api/v1/tickets/:id
```

The list object may be used only for non-authoritative selection context/loading labels if useful.

Real detail state comes from the detail endpoint.

---

# 40. No unnecessary list refetch

Do not refetch the full Ticket List on:

```text
open detail
close detail
Back to Tickets
detail Retry
detail 404
```

List refetch remains tied to existing list workflows, such as:

```text
initial load
tab change
New Request created
explicit Retry
Load More
```

---

# 41. Theme

Ticket Detail must render correctly in:

```text
Light
Dark
```

Use existing CSS variables:

```text
--bg
--surface
--surface-elevated
--text
--text-muted
--border
--accent
--danger
```

Do not introduce a disconnected color system.

---

# 42. Accessibility

Required:

```text
semantic open-detail control
keyboard open works
visible focus state
accessible detail heading
Close is a real button
Back is a real button
loading uses status/live semantics
error uses alert/status semantics appropriately
```

Desktop Ticket Detail is not a modal.

Therefore:

```text
no dialog role solely because it looks like a panel
no modal backdrop
no modal focus trap
```

---

# 43. Performance

A detail read is bounded.

Required:

```text
primary-key ticket lookup
bounded joins
bounded assignment reads
small explicit JSON
no polling
no global lookup-table downloads
```

Do not add:

```text
Redis
cache
queue
WebSocket
new DB index without measured evidence
```

---

# 44. Backend integration tests

Add semantic filename:

```text
backend/client/tickets/ticket_detail_integration_test.go
```

Never use a SPEC-number filename.

Use existing real PostgreSQL temporary-schema infrastructure and official `DATABASE_URL`.

Required cases:

```text
unauthenticated valid detail request → 401

authenticated:
malformed ID → 400 invalid_request
zero ID → 400 invalid_request
negative ID → 400 invalid_request
overflow ID → 400 invalid_request
missing positive ID → 404 ticket_not_found

pending ticket:
core fields correct
null accepted fields
null closed_at
assignment arrays []

accepted ticket:
accepted_by correct
accepted_at correct
Owner semantics preserved

ticket with department/user assignments:
correct rows
stable ordering
arrays not null

closed ticket:
closed status
accepted owner preserved
closed_at correct

historical/inactive reference:
ticket remains readable
no false 404 caused by inactive relation

cross-department staff:
same shared-detail access model

response JSON:
correct {"ticket": ...} envelope
```

---

# 45. Frontend tests

Add/extend semantic tests, preferably:

```text
frontend/tests/ticket-detail.test.ts
```

Required coverage:

```text
getTicket success
strict payload validation
400 mapping
401 mapping
404 mapping
malformed 200 → retryable

selection → loading → success
full description displayed
null fields show —
empty assignments show —
priority title border contract

A→B rapid selection:
A late response cannot overwrite B

close during request:
late response cannot reopen

tab switch:
detail cleared
late response ignored

New Request:
detail cleared

detail Retry:
only detail request retried

Back:
list state retained
no list refetch solely from Back
```

Do not write fragile pixel-position tests for normal responsive CSS.

---

# 46. Manual verification without polluting the real dev database

The current restored development database may contain zero tickets.

Do **not** silently insert permanent "mock" runtime tickets merely to take screenshots.

Automated integration tests may create deterministic tickets in isolated temporary schemas.

For manual app verification:

```text
use an existing real development ticket if available
```

If no real ticket exists and manual UI verification requires creating one:

```text
use the normal New Request product flow only with explicit user approval
```

Otherwise report that the manual detail UI could not be exercised against the persistent dev DB because it contained no ticket.

Never claim manual checks that were not actually performed.

---

# 47. Responsive manual checks

When real ticket data is available, verify:

```text
1920×1080
1366×768
narrow tablet
mobile
```

Desktop:

```text
list visible
detail right-side
no overlay
no body horizontal overflow
table internal horizontal scroll acceptable
panel independently usable
```

Mobile/narrow:

```text
full-width detail
Back works
no horizontal overflow
list state preserved
```

Verify both Light and Dark themes.

---

# 48. Existing working-tree safety

The developer may have unrelated tracked changes, including earlier Baron/UI work.

Before editing:

```bash
git status
git diff --name-only
git diff
```

Record the pre-existing changed paths.

Preserve them exactly.

Never run:

```bash
git reset --hard
git clean -fd
git clean -fdx
git restore .
git checkout -- .
```

Do not stash without explicit permission.

Do not use:

```bash
git add .
```

If a SPEC-07 target file already contains unrelated local hunks:

```text
edit surgically
stage only SPEC-07 hunks
```

If safe isolation is impossible:

```text
STOP before commit and report overlap
```

---

# 49. Source naming discipline

Good:

```text
TicketDetail.svelte
ticket_detail_integration_test.go
ticket-detail.test.ts
ticket-presentation.ts
```

Bad:

```text
Spec07.svelte
spec07_handler.go
spec070_integration_test.go
phase7_*
```

Do not rename unrelated historical files.

---

# 50. Common failure modes — implementation must explicitly avoid all

Before declaring implementation ready, audit for these known mistakes:

```text
1. Detail implemented by reusing list summary only, with no GET /:id.
2. GET /:id loads List() and searches Go memory.
3. 404 becomes 500 because pgx.ErrNoRows is not mapped.
4. Malformed path becomes 404 instead of 400.
5. Unauthenticated malformed path incorrectly bypasses auth.
6. Inactive location/department/user filters make historical ticket disappear.
7. Owner accidentally uses assigned_users instead of accepted_by.
8. Original request department accidentally replaced by assigned department.
9. Empty assignment arrays serialize as null.
10. Assignment order is nondeterministic.
11. getTicket duplicates a weaker JSON parser.
12. ticket_not_found is jammed into the old create-only error-code type and breaks TS.
13. Detail reuses ticketRequestSequence and cancels/corrupts list requests.
14. Detail reuses ticketRequestInFlight and blocks tabs/New Request/Load More.
15. Late A response overwrites newer B.
16. Late response reopens a detail panel the user closed.
17. Old ticket content stays visible while a new ticket loads.
18. Row + nested Title click sends two detail requests due to bubbling.
19. Detail panel is implemented as a modal/overlay covering the table.
20. Split layout causes body-level horizontal scrolling.
21. Mobile squeezes desktop panel beside cards.
22. Back refetches the entire list and loses pagination state.
23. Open/Closed switch leaves stale old-tab detail visible.
24. New Request opens on top of active detail without clearing it.
25. Full description accidentally inherits the list 2-line clamp.
26. Raw RFC3339 strings appear in UI.
27. Priority becomes red text/badge again.
28. Detail Retry reloads the entire app/list.
29. Runtime mock ticket is inserted because local DB has zero tickets.
30. SPEC implementation modifies migration/master-data files.
31. Agent stages pre-existing unrelated working-tree changes.
32. Agent claims viewport/manual verification it did not actually perform.
```

Any occurrence is a SPEC-07 defect.

---

# 51. Verification

Backend:

```bash
cd backend
gofmt -d .
go mod tidy
go vet ./...
go test ./... -count=1
go test -race ./... -count=1
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

Database/repository:

```text
no migration 0021
applied migrations unchanged
14 departments unchanged
141 96 Villas unchanged
155 BWP areas unchanged
564 BWP rooms unchanged
legacy ROOM-8020/7309 inactive
canonical BWP-ROOM-8020/7309 active
```

Do not claim commands not executed.

---

# 52. Acceptance criteria

SPEC-07 implementation is review-ready only if:

```text
backend:
GET /api/v1/tickets/:id exists
auth required
positive int64 validation
missing → 404 ticket_not_found
real PostgreSQL data
historical refs readable
explicit [] assignments
no schema/migration change

frontend:
real detail GET
strict parser
separate list/detail request state
latest-selection-wins
close/tab/new-request invalidation safe
retry + 404 + 401 correct

desktop:
right-side sibling panel
table remains visible
no overlay
1366 regression acceptable

mobile:
dedicated full-width detail
Back keeps list state
no horizontal overflow

presentation:
full description
correct owner/department semantics
consistent timestamps
priority red-border rule
Light/Dark

regression:
Open/Closed unchanged
Load More unchanged
New Request unchanged
Location picker unchanged
auth unchanged

quality:
tests pass
build passes
no runtime mock data
no unrelated changes staged
```

---

# 53. Commit discipline

The implementation commit may include:

```text
SPEC-07 markdown
SPEC-07 source changes
SPEC-07 tests
```

Do not update `PROJECT_CONTEXT.md` yet.

Review final staged diff.

Never stage unrelated pre-existing changes.

Suggested commit:

```text
feat: add ticket detail read
```

Do not mark SPEC-07 CLOSED until an independent review confirms the implementation against this document.

---

# 54. Final implementation report

Report:

```text
Backend endpoint
Repository/query strategy
HTTP/error behavior
Frontend API/error typing
List/detail state separation
Desktop layout
Mobile/tablet behavior
Stale-response protection
Accessibility
Tests
Backend verification
Frontend verification
Manual verification
Database/migration check
Files changed
Pre-existing working-tree changes preserved
Commit SHA
```

Explicitly confirm:

```text
GET /api/v1/tickets/:id reads real PostgreSQL data.
No migration 0021 was created.
No schema/master-data changes were made.
Accept/Assign/Close were not implemented.
Chat/Checklist were not implemented.
No runtime mock data was added.
Pre-existing unrelated working-tree changes were not staged.
```

End with:

```text
SPEC-07 implementation is ready for independent review; it is not marked CLOSED yet.
```
