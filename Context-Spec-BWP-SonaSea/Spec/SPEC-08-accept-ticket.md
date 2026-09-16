# SPEC-08 — Accept Ticket

> Project: **BWP SonaSea**
>
> Repository: `thienty1207/BWP-Staff`
>
> Baseline: `main` at or after `89be98275c92bf704186f9519af703fab0bed063`
>
> Required before implementation: **SPEC-07 must be independently reviewed and marked CLOSED, then `PROJECT_CONTEXT.md` must be aligned to the final SPEC-07 Chat contract.**
>
> Status: **OPEN / NOT CLOSED**
>
> SPEC-08 activates the first real ticket lifecycle action in the Ticket Chat shell: **Accept**.

---

# 1. Goal

Implement the complete Accept vertical slice:

```text
Pending Ticket
    ↓
staff clicks Accept in Ticket Chat
    ↓
POST /api/v1/tickets/:id/accept
    ↓
atomic PostgreSQL transaction
    ↓
status = accepted
accepted_by = authenticated staff
accepted_at = server timestamp
ticket_activity(action='accepted')
    ↓
Chat + Ticket List update immediately
```

After SPEC-08, Accept is a real persisted business action.

No fake frontend-only status change.

---

# 2. User-visible result

Before:

```text
[Pending]

Requester
Location
                  [Accept] [Assign] [Close]
```

After successful Accept:

```text
[Accepted]

Requester                    by Current Staff
Location                     Date / Time

Chats
...
✓ Accepted by Current Staff

                  [Accept disabled] [Assign shell] [Close shell]
```

The ticket remains in **Open** because Open contains:

```text
pending
accepted
```

Accept must not move it to Closed.

---

# 3. Scope

## Included

```text
POST /api/v1/tickets/:id/accept
authenticated actor from server session
pending → accepted transition
accepted_by
accepted_at
updated_at
ticket_activity(action='accepted')
atomic transaction
concurrency protection
same-user idempotent replay
different-user acceptance conflict
closed-ticket conflict
frontend acceptTicket(id)
Accept button activation
in-flight state
safe errors
Chat update
Ticket List row update
Accepted system activity
tests
manual verification
```

## Explicitly excluded

```text
Assign mutation
Close mutation
Chat message sending
Checklist backend
attachments
notifications
WebSocket
polling
Report
Settings
Admin UI
new permission/RBAC subsystem
migration 0021
schema changes
new master data
Redis / queue / cache
```

Do not implement excluded features opportunistically.

---

# 4. Existing ticket semantics remain authoritative

Canonical statuses:

```text
pending
accepted
closed
```

There is no:

```text
assigned
on_hold
in_progress
```

Assignment remains separate from status.

Owner remains:

```text
accepted_by
```

meaning the **first staff member who successfully accepted the ticket**.

`tickets.department_id` remains the original request destination/family.

Accept must not modify department or assignment state.

---

# 5. Authorization rule for SPEC-08

SPEC-08 does not introduce a new RBAC model.

Product rule for this feature:

```text
any authenticated active staff member who can access the shared Tickets workspace
may Accept a pending ticket
```

Do not invent:

```text
same-department-only acceptance
admin-only acceptance
requester cannot accept
assigned-user-only acceptance
```

unless a later explicit product requirement changes this rule.

The backend is authoritative.

Frontend visibility alone is not authorization.

---

# 6. Endpoint

Add authenticated:

```http
POST /api/v1/tickets/:id/accept
```

Use existing auth middleware.

No client-controlled actor fields are accepted.

The actor must come exclusively from:

```text
auth.CurrentPrincipal(c)
```

The client must never choose:

```text
accepted_by
accepted_at
status
actor_user_id
```

---

# 7. Request body

The frontend sends no business payload.

Example:

```http
POST /api/v1/tickets/123/accept
```

No JSON body is required.

If a client sends unrelated body content, it must never influence:

```text
accepted_by
accepted_at
status
activity actor
```

Do not build a generic patch endpoint.

---

# 8. Ticket ID validation

For an authenticated request, `:id` must be a positive signed 64-bit integer.

Reject:

```text
abc
1.5
0
-1
int64 overflow
```

with:

```http
400 Bad Request
```

Existing envelope:

```json
{
  "error": {
    "code": "invalid_request",
    "message": "Invalid request"
  }
}
```

Do not send malformed IDs to PostgreSQL.

---

# 9. Successful first acceptance

For a real pending ticket:

```text
status       pending → accepted
accepted_by  NULL    → authenticated user ID
accepted_at  NULL    → server/database timestamp
updated_at           → same acceptance timestamp
```

Acceptance timestamp is server-controlled.

Never trust browser time.

---

# 10. Atomic business transaction

A successful first acceptance must atomically perform:

```text
1. lock / conditionally claim the ticket
2. verify current lifecycle state
3. update tickets
4. insert ticket_activity(action='accepted')
5. return the accepted ticket state
6. commit
```

Ticket update and activity insertion must either both commit or both roll back.

Forbidden:

```text
ticket accepted but no accepted activity
accepted activity inserted but ticket still pending
```

---

# 11. Ticket activity

Insert exactly one business-history row for the real transition:

```text
ticket_id     = ticket.id
actor_user_id = authenticated staff ID
action        = 'accepted'
metadata      = NULL unless a concrete need is discovered
created_at    = acceptance timestamp
```

Use the existing `ticket_activity` table.

Do not create a new activity schema.

Do not insert a second accepted activity for an idempotent replay.

---

# 12. Timestamp consistency

For the first successful transition:

```text
tickets.accepted_at
tickets.updated_at
ticket_activity.created_at
```

should represent the same acceptance event.

Prefer one database-generated timestamp shared within the transaction.

Do not generate one timestamp in JavaScript and another in Go.

---

# 13. Concurrency — first accepter wins

Two staff may click Accept at nearly the same time.

The backend must guarantee:

```text
exactly one first accepter
accepted_by never flips to the loser
exactly one accepted activity
```

No lost update.

No last-write-wins overwrite.

Use PostgreSQL transaction/locking or an atomic conditional update.

Do not use application-global mutexes.

Do not rely on the frontend to prevent concurrency.

---

# 14. Same-user idempotent replay

If the ticket is already:

```text
status = accepted
accepted_by = current authenticated user
```

then Accept acts as an idempotent replay.

Return:

```http
200 OK
```

with the current accepted ticket.

Do NOT:

```text
change accepted_at
change accepted_by
insert another accepted activity
```

Why:

```text
double click
network timeout after commit
browser retry
```

must not duplicate the business event.

---

# 15. Different-user acceptance conflict

If the ticket is already accepted by another user:

```http
409 Conflict
```

Error:

```json
{
  "error": {
    "code": "ticket_already_accepted",
    "message": "Ticket has already been accepted"
  }
}
```

Do not overwrite the first accepter.

Do not insert another accepted activity.

---

# 16. Closed-ticket conflict

If ticket status is:

```text
closed
```

Accept returns:

```http
409 Conflict
```

Error:

```json
{
  "error": {
    "code": "ticket_closed",
    "message": "Ticket is closed"
  }
}
```

Do not reopen or mutate a closed ticket.

---

# 17. Missing ticket

Valid positive ID that does not exist:

```http
404 Not Found
```

```json
{
  "error": {
    "code": "ticket_not_found",
    "message": "Ticket not found"
  }
}
```

---

# 18. Success response

Successful first acceptance and same-user idempotent replay return:

```http
200 OK
```

Shape:

```json
{
  "ticket": {
    "...": "canonical Ticket representation"
  }
}
```

The returned ticket must contain at minimum the existing fields, with:

```text
status = accepted
accepted_by = current user
accepted_at != null
```

Reuse the existing strict Ticket response model.

Do not introduce a second weaker ticket representation.

---

# 19. Historical reference behavior

Accept must not fail merely because historical ticket references such as:

```text
requester
original department
location
```

have become inactive.

The existing ticket is still valid historical data.

Do not add active-only filters to the acceptance read/return path.

---

# 20. Repository API

Add a clearly named method, for example:

```go
Accept(ctx context.Context, ticketID int64, actor IdentitySummary) (Ticket, error)
```

or equivalent.

Use explicit domain errors such as:

```text
ErrTicketNotFound
ErrTicketAlreadyAccepted
ErrTicketClosed
```

Same-user idempotent replay is not an error.

Do not encode business states by parsing PostgreSQL error strings.

---

# 21. Transaction strategy

Implementation may use either:

```text
SELECT ... FOR UPDATE
```

or a carefully designed:

```text
UPDATE ... WHERE status='pending' RETURNING ...
```

with explicit follow-up state resolution.

The required semantics matter more than the exact SQL shape.

However, the implementation must be easy to read and prove correct under concurrency.

Do not implement:

```text
SELECT without lock
then unconditional UPDATE
```

because two users could both become "first" accepter.

---

# 22. Service behavior

Add a dedicated service method:

```go
Accept(...)
```

Responsibilities:

```text
validate repository configuration
call repository acceptance transaction
preserve domain errors
return canonical updated Ticket
```

Do not put SQL in handler/service.

---

# 23. Handler behavior

Register:

```text
POST /tickets/:id/accept
```

Handler responsibilities:

```text
validate :id
read authenticated principal
build actor identity from session
call service
map domain errors
return {"ticket": ...}
```

Do not trust any frontend user ID.

---

# 24. Error mapping

Required:

```text
400 invalid_request
401 unauthenticated
404 ticket_not_found
409 ticket_already_accepted
409 ticket_closed
500 safe internal error
```

Do not leak:

```text
SQL
constraint names
database driver errors
stack traces
```

---

# 25. No migration

Current schema already contains:

```text
tickets.status
tickets.accepted_by
tickets.accepted_at
tickets.updated_at
ticket_activity
```

Therefore SPEC-08 should require:

```text
no migration 0021
no schema change
no enum change
```

If implementation appears to require a migration:

```text
STOP and explain the blocking deficiency before changing schema.
```

---

# 26. Do not touch assignment

Accept must NOT:

```text
insert ticket_assigned_departments
insert ticket_assigned_users
delete assignments
change tickets.department_id
```

After Accept:

```text
status = accepted
assignments may still be empty
```

This is valid.

---

# 27. Do not touch Close

Accept must NOT set:

```text
closed_by
closed_at
```

Accepting does not close a ticket.

---

# 28. Frontend API

Add:

```ts
acceptTicket(id: number): Promise<TicketSummary>
```

Request:

```text
POST /api/v1/tickets/:id/accept
credentials: include
no business body
```

Use the existing strict ticket parser for the response.

Do not duplicate/paraphrase the ticket parser.

---

# 29. Frontend error model

Frontend must distinguish:

```text
unauthenticated
not_found
already_accepted
closed
retryable
invalid_input
```

Keep Create Ticket and existing selected-ticket error typing safe.

Do not let mutation-specific error codes leak into unrelated New Request UX.

---

# 30. Activate the Chat Accept button

The existing Chat shell already has:

```text
Accept
Assign
Close
```

SPEC-08 activates **Accept only**.

Remain non-operational:

```text
Assign
Close
```

Accept enabled only when:

```text
ticket.status == pending
and accept request is not in flight
```

For:

```text
accepted
closed
```

Accept is disabled.

Do not hide the button and cause layout shift.

---

# 31. Accept in-flight UI

After click:

```text
disable Accept immediately
prevent second submission
```

Use a small local in-flight treatment.

Do not:

```text
block the entire Tickets page
disable Load More
disable theme switch
show full-page spinner
```

No optimistic accepted state before server success.

The database response remains authoritative.

---

# 32. Success UI update

On `200`:

Chat must immediately reflect returned ticket:

```text
Pending badge → Accepted
Owner row → by <accepted_by>
Accepted system activity appears
Accept button becomes disabled
```

Use server-returned:

```text
accepted_by
accepted_at
```

Do not synthesize timestamp with `new Date()`.

---

# 33. Ticket List row update

The selected ticket row/card must also update in place:

```text
Status → Accepted
Owner → accepted_by
```

Do not refetch the entire ticket list solely because Accept succeeded.

Patch only the matching ticket ID using the server-returned Ticket.

Do not reorder unrelated rows.

---

# 34. Chat stays open

Successful Accept must not:

```text
close Chat
reopen Chat
lose scroll unnecessarily
switch tab
```

The same ticket remains selected.

---

# 35. Accepted system activity

The current Chat can derive:

```text
✓ Accepted by <accepted_by>
```

from returned:

```text
accepted_by
accepted_at
```

After success the event must appear without full-page/list reload.

Do not add fake activity.

The backend still inserts the real `ticket_activity('accepted')` history row.

---

# 36. Different-user conflict UI

If API returns:

```text
409 ticket_already_accepted
```

show a compact safe message such as:

```text
This ticket was already accepted by another staff member.
```

Then refresh the selected ticket from:

```text
GET /api/v1/tickets/:id
```

so Chat receives the actual accepter/timestamp.

Patch the corresponding list row with that refreshed Ticket.

Do not keep displaying Pending after the server says another staff member won.

---

# 37. Closed conflict UI

If API returns:

```text
409 ticket_closed
```

show:

```text
This ticket is already closed.
```

Refresh the selected ticket once so Chat/list display actual closed state.

Do not retry Accept automatically.

---

# 38. Unknown network failure

If POST outcome is unknown because of network failure:

```text
show safe inline action error
allow user to Retry
```

A Retry is safe because:

```text
same-user accepted replay → 200 without duplicate activity
different-user winner     → 409 + refresh
```

Do not implement uncontrolled automatic retry loops.

---

# 39. Selected-ticket race protection

User may:

```text
Accept ticket A
then click ticket B before response A arrives
```

Required:

```text
response A may update list row A
response A must NOT replace Chat B
```

Likewise:

```text
Accept A
close Chat
late A success
```

must not reopen Chat.

The mutation result is tied to the ticket ID that was submitted.

---

# 40. Tab interaction during Accept

If Accept is in flight and user switches:

```text
Open ↔ Closed
```

the request may finish.

Required:

```text
do not reopen old Chat
do not corrupt newly loaded list
do not apply A response to an unrelated row
```

If the matching ticket is still present in current list state, it may be patched by ID.

Otherwise ignore the local list patch.

Server state remains authoritative on later load.

---

# 41. New Request interaction during Accept

If New Request opens while Accept is in flight:

```text
do not fake-cancel a mutation that may already commit
```

The Chat can close according to existing behavior.

A late Accept response must not reopen it.

No global mutation cancellation assumption.

---

# 42. Button semantics after success

After a ticket is accepted:

```text
Accept button remains visible but disabled
```

Use normal disabled styling.

Do not change it into another action.

Do not enable Assign/Close in SPEC-08 unless those buttons are already visual shells; they remain non-mutating.

---

# 43. Accessibility

Accept button:

```text
real <button>
keyboard operable
visible focus
disabled state exposed correctly
in-flight state communicated appropriately
```

Errors:

```text
use suitable role/status/alert semantics
```

Do not announce successful Accept multiple times.

Accepted activity can serve as visible confirmation.

---

# 44. No realtime requirement

SPEC-08 does not add WebSocket or polling.

Two different browser sessions do not need instant cross-client push in this SPEC.

Concurrency is protected server-side.

A stale other client becomes current when it:

```text
attempts Accept
reloads/list refreshes
reopens/refreshes ticket
```

Realtime comes later if required.

---

# 45. Backend integration tests

Add semantic filename:

```text
accept_ticket_integration_test.go
```

Do not create:

```text
spec08_integration_test.go
```

Use real Fiber routing + real PostgreSQL temporary-schema isolation with official `DATABASE_URL`.

Required cases:

```text
unauthenticated → 401

authenticated invalid IDs:
text → 400
zero → 400
negative → 400
overflow → 400

missing ticket:
404 ticket_not_found

first acceptance:
200
status accepted
accepted_by actor
accepted_at non-null
updated_at reflects transition
original requester unchanged
original department unchanged
location unchanged
assignments unchanged

activity:
exactly one action='accepted'
actor_user_id = authenticated actor
activity timestamp represents same acceptance event

same actor replay:
200
same accepted_by
same accepted_at
no second accepted activity

different actor after acceptance:
409 ticket_already_accepted
first accepted_by preserved
first accepted_at preserved
no second activity

closed ticket:
409 ticket_closed
no mutation

actor spoofing:
request body cannot choose accepted_by

concurrent two-user acceptance:
exactly one first transition
one user wins
loser receives conflict
one accepted activity
```

---

# 46. Transaction rollback test

Add at least one regression that proves the business operation is atomic if reasonably possible using the existing isolated test infrastructure.

Goal:

```text
if accepted activity insertion fails
ticket must not remain accepted
```

Do not add production-only hooks merely to make this test possible.

If safe failure injection would require invasive production abstractions, document the transaction-boundary source review instead and keep implementation simple.

---

# 47. Frontend tests

Use semantic filenames.

Required:

```text
acceptTicket API:
success parser
401
404
409 already_accepted
409 closed
network/retryable
malformed 200

Chat action:
pending → Accept enabled
accepted → Accept disabled
closed → Accept disabled
double click cannot submit twice
success applies server Ticket
status changes Accepted
Owner appears from server
accepted activity appears
no client-generated accepted timestamp

list:
matching row patched by ID
status/owner update
no full list refetch on success

race:
Accept A then Chat B → A does not replace B
Accept A then close → response does not reopen
conflict refreshes actual ticket
```

Do not remove existing SPEC-07 tests.

---

# 48. Manual verification

With a real pending development ticket:

```text
1. Open its Chat.
2. Confirm Accept enabled.
3. Click Accept once.
4. Confirm badge → Accepted.
5. Confirm "by <current user>" appears.
6. Confirm accepted activity appears.
7. Confirm list Owner/Status update.
8. Confirm Chat remains open.
9. Confirm Accept disabled.
10. Refresh browser.
11. Confirm accepted state persisted.
```

Database verification:

```text
tickets.status = accepted
accepted_by = current user
accepted_at != null

ticket_activity:
exactly one accepted row for the transition
```

Do not use direct SQL to fake the acceptance for the manual product test.

---

# 49. Multi-user concurrency manual check

If two authenticated test users are available:

```text
both open same pending ticket
both click Accept near-simultaneously
```

Expected:

```text
one user becomes Owner
other receives already-accepted conflict
Owner never changes
one accepted activity
```

If two sessions/users are not available, do not claim this manual check.

Automated concurrency coverage remains required.

---

# 50. Performance

Accept is a small bounded transaction.

Expected work:

```text
one ticket lifecycle lookup/lock or conditional update
one ticket update
one ticket_activity insert
bounded response read
```

Do not add:

```text
Redis
queue
distributed lock
global Go mutex
list-wide scans
```

PostgreSQL is the source of truth and concurrency authority.

---

# 51. Common failure modes — must avoid

Any of these is a SPEC-08 defect:

```text
1. Frontend changes Pending → Accepted without server success.
2. Client sends accepted_by and backend trusts it.
3. accepted_at comes from browser time.
4. Two concurrent users can overwrite accepted_by.
5. Last accepter wins.
6. Same-user retry inserts duplicate accepted activity.
7. Same-user retry changes accepted_at.
8. Different-user retry returns 200 as though they own it.
9. Closed ticket becomes accepted/reopened.
10. Accept accidentally modifies assignments.
11. Accept accidentally changes original department.
12. accepted event is fake frontend data only.
13. ticket_activity is inserted outside the acceptance transaction.
14. ticket updates but activity insert failure still commits.
15. Accept success refetches the entire list unnecessarily.
16. Accept response for ticket A replaces currently open Chat B.
17. Closing Chat while Accept is in flight allows late response to reopen it.
18. Accept button remains enabled after accepted state.
19. Accepted timestamp is synthesized on frontend.
20. Assign/Close are accidentally activated in this SPEC.
21. New permission system is invented.
22. Migration 0021 is created unnecessarily.
23. Runtime mock data is added.
24. Source/test filenames use SPEC numbers.
25. Unrelated working-tree changes are staged.
```

---

# 52. Existing working-tree safety

Before editing:

```bash
git status
git diff --name-only
git diff
```

Preserve unrelated local work.

Never run:

```bash
git reset --hard
git clean -fd
git clean -fdx
git restore .
git checkout -- .
git add .
```

Do not stash without explicit permission.

Use surgical staging if target files contain unrelated hunks.

---

# 53. Expected implementation shape

Likely backend:

```text
backend/client/tickets/handler.go
backend/client/tickets/service.go
backend/client/tickets/repository.go
backend/client/tickets/accept_ticket_integration_test.go
```

Likely frontend:

```text
frontend/src/lib/client/tickets/api.ts
frontend/src/lib/client/tickets/model.ts
frontend/src/lib/components/TicketChat.svelte
frontend/src/routes/+page.svelte
frontend/tests/ticket-chat.test.ts
frontend/tests/tickets-api.test.ts
```

Only edit files actually needed.

Do not create empty modules.

---

# 54. Verification

Backend:

```powershell
$env:Path = "C:\msys64\ucrt64\bin;$env:Path"
$env:CGO_ENABLED = "1"
$env:CC = "C:\msys64\ucrt64\bin\gcc.exe"
$env:CXX = "C:\msys64\ucrt64\bin\g++.exe"

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

No dependency upgrades.

---

# 55. Database verification

Confirm:

```text
no migration 0021
existing migrations unchanged
schema unchanged
master data unchanged
```

Still expected:

```text
14 active requestable departments
141 96 Villas
155 BWP areas
564 BWP rooms
legacy ROOM-8020 / ROOM-7309 inactive
canonical BWP-ROOM-8020 / BWP-ROOM-7309 active
```

---

# 56. Acceptance criteria

SPEC-08 is implementation-ready for independent review only when:

```text
POST /api/v1/tickets/:id/accept exists
auth required
actor always comes from session
pending → accepted persisted
accepted_by correct
accepted_at server-controlled
ticket_activity accepted inserted atomically

same-user replay:
200
no duplicate activity
timestamp unchanged

different-user:
409
first owner preserved

closed:
409
no mutation

frontend:
Accept enabled only for pending
double submit prevented
server response updates Chat
server response updates list row
Accepted activity appears
Accept disables
no list-wide refetch
race-safe if selected ticket changes

quality:
tests pass
race tests pass
no migration
no runtime mock data
no unrelated changes staged
```

---

# 57. Commit discipline

Do not update `PROJECT_CONTEXT.md` in the implementation commit.

Do not mark SPEC-08 CLOSED until independent review + manual product verification.

Suggested commit:

```text
feat: add ticket acceptance
```

Stage only SPEC-08 implementation/documentation.

Never use `git add .` in a dirty worktree.

---

# 58. Final implementation report

Return:

```text
Backend endpoint
Authorization
Acceptance transaction
Concurrency behavior
Idempotency
Activity history
HTTP errors
Frontend Accept action
Chat update
Ticket List update
Race protection
Tests
Backend verification
Frontend verification
Manual verification
Database/migration check
Files changed
Pre-existing changes preserved
Commit SHA
```

Explicitly confirm:

```text
accepted_by always comes from the authenticated server session.
accepted_at is server/database controlled.
Same-user replay does not duplicate acceptance.
Different users cannot overwrite the first accepter.
Exactly one accepted activity is created for the transition.
Accept does not modify assignment.
Assign/Close remain non-mutating.
No migration 0021 was created.
No runtime mock data was added.
```

End with:

```text
SPEC-08 implementation is ready for independent review; it is not marked CLOSED yet.
```

---

# Latest UI supersession for SPEC-08 hardening

SPEC-08 remains **OPEN / NOT CLOSED**. The following UI rules are the latest
binding corrections for the implementation and supersede any earlier list
presentation assumption that actions only live inside Chat.

## Desktop ticket list

- The far-right `Action` column is visible on the desktop table.
- Every desktop row shows `Accept`, `Assign`, and `Close` without layout shift.
- `Accept` is the real SPEC-08 lifecycle action for pending tickets and may be
  used without opening Chat.
- `Assign` and `Close` remain disabled, non-mutating shells.
- Chat continues to render its own `Accept`, `Assign`, and `Close` action row.
- Action-button activation is isolated from whole-row Chat activation.

## Mobile ticket list and Chat

- Mobile cards do not receive the desktop `Action` column or a separate action
  group; their whole-card interaction still opens Chat.
- Mobile and desktop Chat close with the compact `X` control. Do not render a
  `Back to Tickets` text control.
- Closing Chat returns to the already-loaded list and preserves its active tab,
  pagination, and loaded data without a close-only refetch.

## Priority and metadata hierarchy

- Priority styling applies only to the primary ticket title in the desktop
  list, mobile card, and Chat summary.
- Priority uses a subtle theme-aware soft-red background with a small radius;
  it must not be an outline-only red box and must not make the title bold.
- The Chat summary keeps a single-line title ellipsis and places the status
  beside it.
- Location is the strongest/bold ticket metadata. Other metadata remains
  compact and normal-weight.
- The created-request activity line `Title: ...` remains conversation content
  and does not receive primary-title priority styling.

The shared Accept flow remains ID-based and race-safe: row and Chat actions
must patch only the submitted ticket, must not duplicate a request, and must
not replace a newer Chat selection.
