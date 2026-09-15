# SPEC-07 — Ticket Chat Shell + Conversation Foundation

> Project: **BWP SonaSea**
>
> Repository: `thienty1207/BWP-Staff`
>
> Current implementation baseline: `main` at or after `ed2b47a46d3d65a7e65a868b8e627f7a0ec8a53a`
>
> Closed before this SPEC: SPEC-01 through SPEC-06.7.
>
> Status: **OPEN / NOT CLOSED**
>
> This SPEC **replaces the previous SPEC-07 — Ticket Detail Read contract**. The previous SPEC-07 was never CLOSED and used the wrong user-facing concept. Git history preserves it; it must not remain the active canonical SPEC document.

---

# 1. Canonical product correction

The correct user flow is:

```text
Ticket List
    ↓ click/select a ticket
Ticket Chat
```

Not:

```text
Ticket List
    ↓
Ticket Detail form/page
    ↓
Chat later
```

The right-side surface is a **Chat / Conversation workspace for the selected ticket**.

The compact header of the Chat already summarizes the ticket. Therefore the product must **not** show a long user-facing detail inspector containing separate Request Information / Assignment / Timing cards.

The screenshots supplied by the user for Sara are the binding visual/interaction reference for this SPEC.

Latest explicit visual direction overrides the earlier Ticket Detail concept.

---

# 2. Visual reference interpretation

The supplied Sara references establish this hierarchy:

```text
Chat                                             X
┌─────────────────────────────────────────────────┐
│ ticket icon  Title                    [STATUS]   │
│ Requester                         by Owner       │
│ Location                         Date / Time     │
├─────────────────────────────────────────────────┤
│               Chats     Checklist               │
├─────────────────────────────────────────────────┤
│                                                 │
│ conversation / system activity                  │
│                                                 │
│ only this region scrolls                        │
│                                                 │
├─────────────────────────────────────────────────┤
│ 📎  🎤   Type a message                    ⋮    │
├─────────────────────────────────────────────────┤
│       Accept        Assign        Close          │
└─────────────────────────────────────────────────┘
```

Desktop:

```text
Ticket List | Chat panel
```

Mobile / narrow tablet:

```text
Chat panel becomes the main full-width content view.
```

The panel is compact, utility-oriented, and persistent in height.

It must **not** become a vertically growing stack of large cards.

---

# 3. User-visible result after SPEC-07

After SPEC-07:

```text
1. User opens Open or Closed tickets.
2. User selects a ticket.
3. Chat shell opens for that ticket.
4. Chat header shows a compact real ticket summary.
5. Chats tab shows real system activity derived from the persisted ticket.
6. If the ticket was accepted, the accepted activity appears.
7. Composer/action controls are visible as the future interaction shell.
8. User closes Chat or goes Back on mobile.
```

SPEC-07 does **not** yet send chat messages or mutate ticket state.

---

# 4. Scope

## Included

```text
retain GET /api/v1/tickets/:id as the selected-ticket read source
replace user-facing Ticket Detail UI with Ticket Chat shell
rename/refactor detail UI state to chat-panel semantics
desktop docked right-side Chat
mobile/narrow-tablet full-width Chat view
compact ticket summary header
Chats / Checklist tab bar
system activity conversation foundation
created-request system activity
accepted system activity when persisted fields support it
composer visual shell
Accept / Assign / Close visual shell
loading / retry / 404 / 401 handling
stale request protection
Open / Closed interaction
New Request interaction
responsive layout
Light / Dark theme
accessibility
tests
```

## Explicitly excluded

```text
sending chat messages
chat persistence tables
chat message API
chat polling
WebSocket
realtime sync
attachments upload
voice recording
Checklist data/backend
Accept mutation
Assign mutation
Close mutation
notifications
ticket activity timeline API
Report
Settings
Admin flows
migration 0021
schema changes
new master data
Redis / queue / cache
```

Do not fake excluded functionality.

---

# 5. Existing backend work is retained

The already implemented authenticated endpoint remains valid infrastructure:

```http
GET /api/v1/tickets/:id
```

It is now used to hydrate:

```text
Chat summary
+
real system activity
```

Do **not** remove a correct endpoint merely because the previous frontend concept was wrong.

Existing behavior remains:

```text
authentication required
positive signed int64 ID
400 invalid_request
404 ticket_not_found
real PostgreSQL read
historical inactive references readable
Owner = accepted_by
original request Department preserved
stable assignment ordering
[] assignment arrays
```

No backend Chat mutation endpoint is added in SPEC-07.

---

# 6. Remove the old user-facing Ticket Detail concept

The current UI introduced by the superseded SPEC-07 contains concepts such as:

```text
Ticket Detail
Request information
Assignment
Timing
Created
Due
Accepted
Closed
Updated
```

as a long inspector.

That presentation is not part of the product direction.

Remove it from the selected-ticket user experience.

The selected-ticket UI must not look like:

```text
large stacked cards
metadata inspector
admin record viewer
form-like detail page
```

Backend fields can remain available in the API even if the Chat shell does not display all of them.

---

# 7. Canonical frontend naming

The corrected UI must use Chat-oriented names.

Preferred:

```text
TicketChat.svelte
ticket-chat-state.ts
ticket-chat.test.ts
ticket-chat / chat-panel CSS classes
```

Avoid active frontend names such as:

```text
TicketDetail.svelte
detail-state.ts
ticket-detail-* CSS
```

where they describe the obsolete user-facing concept.

The existing backend `ticket_detail_integration_test.go` may remain because it genuinely tests the `GET /tickets/:id` read endpoint.

Do not rename unrelated historical files.

---

# 8. Desktop Chat placement

On desktop the selected ticket opens a Chat column on the right.

Concept:

```text
Tickets workspace
├── Ticket List region
└── Ticket Chat region
```

Hard requirements:

```text
Chat is a sibling layout column
Chat does not overlay the table
no modal backdrop
table remains visible
Chat does not expand document height based on conversation content
Chat remains visually compact like the Sara reference
```

The selected ticket has visible but subtle selection feedback without changing table geometry. The feedback must clear when Chat closes and move when another ticket is selected.

Do not add a new Action column just for opening Chat.

---

# 9. Desktop Chat width

The Sara reference uses a narrow utility panel, not a large detail inspector.

Target behavior:

```text
approximately 19rem–22rem on normal desktop
bounded maximum width
list keeps the majority of workspace width
```

A reasonable CSS strategy is equivalent to:

```text
list: minmax(0, 1fr)
chat: clamp(19rem, ~20–22vw, 22rem)
```

Exact values may adapt to existing spacing, but the panel must remain noticeably narrower than the old Ticket Detail inspector.

At `1366×768`:

```text
Chat must not overlap list
body-level horizontal scrolling must not be introduced
the ticket table may retain its existing internal horizontal scroll
```

---

# 10. Desktop Chat height

The Chat must behave like an application panel, not content that keeps growing downward.

Required internal layout:

```text
row 1: Chat window header
row 2: compact ticket summary
row 3: Chats / Checklist tabs
row 4: conversation minmax(0, 1fr)
row 5: composer
row 6: action row
```

Only the conversation region should normally scroll.

The following remain fixed inside the panel while conversation scrolls:

```text
Chat title + close
ticket summary
tabs
composer
Accept / Assign / Close row
```

Use CSS grid/flex with:

```text
min-height: 0
overflow: hidden on shell
overflow-y: auto on conversation
```

or an equivalent robust layout.

Do not place `overflow: auto` on the whole Chat panel as the primary behavior.

---

# 11. Tickets workspace vertical behavior

The Chat should fit the available Tickets workspace/viewport height.

Do not let conversation content determine page height.

A valid implementation may make the authenticated Tickets content a viewport-aware layout where:

```text
page header = auto
Open / Closed tabs = auto
workspace = remaining usable height
```

Do not perform a broad application-shell rewrite.

Target only what is necessary to give the Chat a bounded application-panel height.

---

# 12. Chat window header

Top row:

```text
Chat                                      X
```

Requirements:

```text
thin compact header
"Chat" text on the left
real Close button / X on the right
no oversized card heading
```

Desktop Close:

```text
closes selected Chat
preserves loaded ticket list
does not refetch list
```

---

# 13. Compact ticket summary

Immediately below the Chat window header, show a compact summary.

Required data:

```text
Title
Status
Requester
Location
Owner when available
Created date/time
```

Suggested reference-aligned layout:

```text
[ticket icon] Title                    [STATUS]
Requester                         by Owner
Location                         Created time
```

Rules:

```text
Title appears once
Location displays location.name only
do not display internal BWP-ROOM-* code
Requester may include department code as existing identity label
Owner = accepted_by
pending with no Owner must not invent one
```

Sara-fidelity constraint:

```text
render the summary as three compact content rows:
title + status
requester + optional by Owner
location + created date/time

do not render visible Requester, Owner, Location, or Created field labels in the summary
pending without accepted_by omits the owner text
```

Do not show large sections for:

```text
Assigned Departments
Assigned Users
Due
Updated
Accepted At
Closed At
```

in the persistent summary.

Those fields remain API data, not primary Chat UI.

---

# 14. Status presentation

Reuse existing ticket status semantics:

```text
pending
accepted
closed
```

Compact status badge in summary.

Do not invent:

```text
assigned status
on hold status
```

Do not copy Sara statuses that do not exist in BWP SonaSea.

---

# 15. Priority presentation in Chat

Priority remains the BWP SonaSea rule:

```text
normal text color
thin red/danger content-sized title border when priority=true
no Priority badge
no red Title text
```

Do not let the compact summary become oversized because of priority.

---

# 16. Chats / Checklist tab bar

Reference contract:

```text
Chats      Checklist
```

For SPEC-07:

```text
Chats = active
Checklist = visible shell only
```

Checklist backend/functionality is not implemented here.

Checklist must not trigger a fake API call.

Safe behavior:

```text
Checklist control is disabled/non-operational in SPEC-07
```

while retaining the visual position from the Sara reference.

This latest screenshot-based direction supersedes earlier assumptions about hiding Checklist elsewhere for this Chat shell.

Future Checklist SPEC may make it interactive.

---

# 17. Conversation foundation

The Chats view uses the large middle region.

It contains **real system activity derived from real ticket data**.

No fabricated user chat messages.

Initial supported conversation items:

```text
1. request-created activity
2. accepted activity when persisted accepted fields are present
```

Do not create artificial messages simply to make the panel look populated.

---

# 18. Request-created activity

Every existing ticket can produce a request-created activity from persisted data.

Reference style:

```text
Requester Name
has created a new request
Location: <location name or —>
Title: <ticket title>

                               Created date/time
```

This is a **presentation of persisted ticket creation data**, not a fake chat record.

The created activity must remain conversation-like rather than becoming a metadata grid:

```text
Location: <location name or —>
Title: <ticket title>
<non-blank description, when present>
```

Do not use a two-column LOCATION/TITLE metadata layout, visible field-label cells, or
unnecessary internal metadata dividers inside this activity.

Do not write a new database row just to display this item.

The item must use:

```text
requester
location.name
title
created_at
```

---

# 19. Accepted activity

Show an accepted activity only when sufficient persisted data exists:

```text
accepted_by != null
accepted_at != null
```

Reference style:

```text
✓ Accepted by <accepted_by identity>

                               Accepted date/time
```

Owner semantics remain:

```text
accepted_by
```

Do not use `assigned_users` as Owner.

Pending ticket:

```text
no accepted event
```

---

# 20. Closed activity

SPEC-07 does **not** invent a "Closed by ..." activity unless the currently exposed real model contains sufficient canonical actor data for that presentation.

`closed_at` alone is not permission to invent a closer identity.

If the necessary data is not in the current read model:

```text
do not render a fake closed actor event
```

Close mutation/activity can be handled in its later SPEC.

---

# 21. No user chat messages yet

SPEC-07 does not implement message history.

Conversation must not contain:

```text
hardcoded greetings
sample messages
fake users
frontend demo message arrays
fake timestamps
```

Empty conversation beyond available system activities is valid.

Runtime data remains real.

---

# 22. Composer shell

At the bottom of Chat, visually match the reference structure:

```text
📎   🎤   Type a message                         ⋮
```

SPEC-07 does not send messages.

Therefore:

```text
composer is visual shell only
no POST request
no optimistic message
no local fake message append
```

The text field/control must be non-submittable in this SPEC.

Recommended:

```text
disabled or readonly with clear aria-disabled semantics
```

while maintaining the intended visual layout.

Do not create deceptive fake success behavior.

---

# 23. Composer icons

Reference-aligned placeholders:

```text
attachment
voice/microphone
overflow/more
```

They are shell controls only in SPEC-07.

They must not:

```text
open upload workflow
record audio
open fake feature menus
change ticket state
```

Future SPECs can activate them.

Use existing icon strategy; do not add a heavy icon dependency solely for this panel.

---

# 24. Accept / Assign / Close shell

Bottom action row:

```text
Accept       Assign       Close
```

These are visible because they are part of the final Chat workspace design.

But SPEC-07 does not implement their backend mutations.

Therefore:

```text
no POST accept
no POST assign
no POST close
no fake state update
```

Controls must be non-operational/disabled in SPEC-07.

They may visually preserve the final layout.

Future action SPECs will activate them individually.

---

# 25. Selected-ticket read API

Opening Chat must still call:

```http
GET /api/v1/tickets/:id
```

The list object is not the authoritative Chat data source.

Do not use only the list summary and skip the selected-ticket read.

Existing endpoint response remains strict real PostgreSQL data.

---

# 26. Error behavior

Keep existing selected-ticket error semantics.

```text
401 → existing authenticated app redirect to /login
404 ticket_not_found → stable Chat not-found state
retryable/network/5xx/malformed → inline Chat error + Retry
```

Retry:

```text
retries only GET /tickets/:id
does not reload entire page
does not refetch the whole ticket list
```

Do not display raw backend error text.

---

# 27. Loading state

When a ticket is selected:

Desktop:

```text
Chat shell opens
conversation/body shows loading
ticket table remains visible
```

Mobile:

```text
Chat screen shows loading
Back remains available
```

Do not show old ticket content while the new ticket is loading.

---

# 28. Separate list state and Chat state

Preserve the correct architecture introduced during the previous implementation.

List state and selected-ticket Chat state must remain separate.

Do not reuse:

```text
ticketRequestSequence
ticketRequestInFlight
```

for Chat read requests.

Chat requires its own request state/sequence or AbortController.

This prevents detail/chat loading from breaking:

```text
Load More
Open / Closed
New Request
list Retry
```

---

# 29. Latest selected ticket wins

Required race behavior:

```text
select A
select B
B response arrives
A response arrives late
→ Chat still displays B
```

Close Chat during request:

```text
late response cannot reopen Chat
```

Switch tab during request:

```text
late response ignored
```

Open New Request during request:

```text
Chat cleared
late response ignored
```

---

# 30. Open / Closed interaction

Switching:

```text
Open ↔ Closed
```

must:

```text
close Chat
invalidate selected-ticket request
continue normal list loading
```

Do not leave an Open ticket Chat beside the Closed list or vice versa.

---

# 31. New Request interaction

Opening New Request:

```text
closes Chat
invalidates selected-ticket request
opens existing New Request modal
```

New Request behavior itself remains unchanged.

Do not let the Chat composer/action shell interact behind the modal.

---

# 32. Load More interaction

Load More remains independent from Chat.

While Chat is open:

```text
Load More can still operate
list cursor remains authoritative
Chat selection remains open unless explicitly closed
```

List request sequencing and Chat sequencing must not invalidate each other.

---

# 33. Ticket selection semantics

Do not add an Action column.

The whole ticket is the pointer target.

Desktop:

```text
clicking anywhere on the ticket row MUST open Chat for that ticket
```

Mobile:

```text
tapping anywhere on the visible ticket card MUST open Chat for that ticket
```

The required interaction is: **pointer click/tap anywhere on the ticket row/card MUST open Chat**.

The implementation must provide both:

```text
whole row/card pointer target
keyboard-operable semantic activation control
```

The title remains a real semantic button/link-like control so keyboard and screen-reader users have an explicit activation path.

```text
Title acts as semantic button/link-like control
```

If the pointer target and semantic title control both exist:

```text
one user activation = one GET request
```

Prevent event-bubbling double fetches.

The selected row/card must use `chatState.selectedTicketID` as its source of truth and show subtle feedback in both Light and Dark themes.

---

# 34. Mobile / narrow-tablet behavior

On narrow layout:

```text
ticket list
→ select ticket
→ Chat becomes full-width main content
```

Reference contract:

```text
Chat header
compact summary
Chats / Checklist
conversation
composer
action row
```

Back:

```text
returns to the already-loaded list
preserves Open / Closed selection
preserves loaded pagination state
does not refetch solely because Back was pressed
```

No horizontal scrolling.

Do not squeeze list + right panel side-by-side on mobile.

---

# 35. Mobile height behavior

The mobile Chat should feel like a messaging screen.

Required:

```text
top header/summary/tabs remain above conversation
conversation fills remaining height
composer/action row stay at bottom
conversation scrolls internally
```

Avoid:

```text
whole Chat page becoming an extremely tall document
composer disappearing far below conversation history
```

Use `100svh`/available-shell-height behavior carefully so mobile browser chrome does not make controls inaccessible.

---

# 36. Light / Dark theme

Chat shell must support both existing themes.

Use project variables such as:

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

Do not hard-code a disconnected Sara blue/white palette into Dark mode.

Sara is the **layout/interaction reference**.

BWP SonaSea keeps its own theme identity.

---

# 37. Accessibility

Required:

```text
semantic ticket-open control
visible focus
Chat region has accessible heading
Close is a real button
Back is a real button on narrow screens
Chats tab has proper selected semantics
Checklist shell communicates disabled state
loading announced with status/live semantics
errors announced appropriately
composer/action shell communicates disabled state
```

Chat is a docked panel, not a modal:

```text
no modal focus trap
no aria-modal on desktop Chat
no backdrop
```

---

# 38. Frontend cleanup from superseded implementation

Remove/replace obsolete user-facing code.

Expected cleanup may include:

```text
remove TicketDetail.svelte
replace with TicketChat.svelte

rename/refactor:
detail-state.ts
→ ticket-chat-state.ts

rename/refactor:
ticket-detail.test.ts
→ ticket-chat.test.ts

replace ticket-detail-* CSS
→ ticket-chat-* CSS
```

Do not preserve dead duplicate components.

Do not keep the old inspector hidden "just in case".

Git history already preserves it.

---

# 39. Frontend state reuse

The previous `TicketDetailStateMachine` race-control logic is conceptually useful.

Reuse the logic if sound, but rename/refactor it to Chat semantics.

For example:

```text
TicketChatStateMachine
```

It should manage:

```text
closed
loading
ready
not_found
error
selectedTicketID
ticket
errorMessage
request sequence
```

Do not rewrite correct concurrency logic unnecessarily.

---

# 40. Backend scope

The backend selected-ticket read endpoint is already implemented.

SPEC-07 correction should not rewrite it unless an actual bug is discovered.

No Chat backend endpoints yet.

Do not add:

```text
GET /tickets/:id/messages
POST /tickets/:id/messages
POST /tickets/:id/accept
POST /tickets/:id/assign
POST /tickets/:id/close
```

in this SPEC.

---

# 41. No migration

No database changes.

Expected:

```text
0001..0017
0019
0020
```

`0018` remains retired.

No `0021`.

Do not alter:

```text
14 requestable departments
141 96 Villas
155 BWP areas
564 BWP rooms
legacy ROOM-8020 / ROOM-7309 cleanup state
```

---

# 42. No runtime mock data

Forbidden:

```text
fake chat messages
sample chat arrays
fake activity timestamps
fake accepted user
fake ticket summary
fake API responses in runtime
persistent SQL demo ticket inserted silently
```

Allowed:

```text
deterministic test-only fixtures in isolated tests
```

System activity is derived from real persisted ticket fields.

---

# 43. Automated backend tests

Keep the existing selected-ticket read integration coverage.

It should continue to verify:

```text
401
400 invalid IDs
404 ticket_not_found
pending
accepted
closed
assignments
historical inactive references
cross-department read
safe response envelope
```

Do not delete good backend coverage because the frontend design changed.

No new backend Chat-message tests are expected because message backend is not implemented.

---

# 44. Frontend tests

Use semantic file naming:

```text
ticket-chat.test.ts
```

Required coverage:

```text
ticket selection calls getTicket
loading → ready
A→B stale response protection
close invalidates late response
tab switch closes Chat
New Request closes Chat
Retry calls selected-ticket GET only
404 state
401 behavior

Chat structure:
Chat heading + Close
compact summary
title only once
location.name only
status
requester
Owner from accepted_by
created timestamp
Chats active
Checklist shell disabled
conversation region
composer shell visible but non-submittable
Accept / Assign / Close visible but non-operational

system activity:
created-request item uses real ticket data
accepted item only when accepted_by + accepted_at exist
pending ticket does not fabricate accepted item

obsolete inspector:
no "Request information" section
no "Assignment" inspector section
no "Timing" inspector section
no Ticket Detail heading
```

Do not rely only on brittle string tests when behavior can be unit-tested directly.

---

# 45. Manual verification

Manual verification requires at least one real development ticket.

If persistent dev DB has zero tickets:

```text
do not silently insert a fake SQL ticket
```

Ask the user before creating one through the normal New Request flow.

When a real ticket is available, verify:

```text
desktop wide
1366×768
mobile/narrow
Light
Dark
```

Desktop:

```text
select ticket → Chat opens
Chat narrow on right
table remains visible
Chat does not cover table
Chat does not lengthen page
only conversation region scrolls
Close works
```

Mobile:

```text
select ticket → full-width Chat
Back works
composer/action row remains at bottom
conversation scrolls
no horizontal overflow
```

---

# 46. Binding UI comparison checklist

Before declaring implementation review-ready, compare against the supplied Sara reference.

The result must feel like:

```text
a compact chat utility
```

not:

```text
a record details page
```

Check:

```text
[ ] "Chat" header is small
[ ] summary is compact
[ ] status sits near title
[ ] requester/location are concise
[ ] owner/time align compactly
[ ] Chats/Checklist strip is directly below summary
[ ] conversation gets most vertical space
[ ] composer is pinned near bottom
[ ] Accept/Assign/Close row is pinned at bottom
[ ] no tall metadata cards
[ ] no full-panel vertical metadata scroll
```

---

# 47. Known failure modes to prevent

Any of these is a SPEC-07 defect:

```text
1. Keeping Ticket Detail inspector and merely renaming its heading "Chat".
2. Rendering Request Information / Assignment / Timing cards inside Chat.
3. Whole panel scrolls while composer disappears below the fold.
4. Chat panel grows page height based on conversation.
5. Chat panel overlays the ticket table.
6. Chat is too wide and leaves little room for list.
7. Mobile squeezes list + panel side-by-side.
8. Fake chat messages are hardcoded.
9. Composer visually sends but nothing is persisted.
10. Accept/Assign/Close pretends success without backend mutations.
11. Checklist pretends to save data.
12. Owner uses assigned_users instead of accepted_by.
13. Location exposes BWP-ROOM-* internal code.
14. Title appears twice.
15. Old ticket remains visible while new selection loads.
16. Late A response overwrites B.
17. Close/tab/New Request allows late response to reopen Chat.
18. Back refetches and loses list pagination.
19. Detail/chat request state is mixed with list request state.
20. Old TicketDetail/detail-state dead code remains active.
21. Old SPEC-07 Ticket Detail document remains the active canonical file.
22. Migration 0021 or Chat schema is introduced prematurely.
23. Runtime mock ticket/message data is added.
24. Unrelated user working-tree changes are staged.
25. Manual UI verification is claimed without actually testing a real ticket.
```

---

# 48. Existing working-tree safety

The user may have unrelated local changes.

Before editing:

```bash
git status
git diff --name-only
git diff
```

Preserve pre-existing changes.

Never run:

```bash
git reset --hard
git clean -fd
git clean -fdx
git restore .
git checkout -- .
git add .
```

If target files contain unrelated user hunks:

```text
edit surgically
stage only SPEC-07 correction hunks
```

If isolation is unsafe:

```text
STOP before commit and report overlap.
```

---

# 49. Canonical SPEC file replacement

The old active file:

```text
Context-Spec-BWP-SonaSea/Spec/SPEC-07-ticket-detail-read.md
```

must no longer remain as the canonical active SPEC.

Replace it with:

```text
Context-Spec-BWP-SonaSea/Spec/SPEC-07-ticket-chat-shell-conversation-foundation.md
```

Prefer a Git rename/delete+add so history remains understandable.

Do not create:

```text
SPEC-07.1
SPEC-07-old
SPEC-07-ticket-detail-read-v2
```

There is only one active SPEC-07 contract.

---

# 50. PROJECT_CONTEXT during implementation

Do not mark SPEC-07 CLOSED in `PROJECT_CONTEXT.md`.

The current context may still contain the stale phrase:

```text
SPEC-07 Ticket Detail Read
```

For implementation priority:

```text
latest explicit user instruction
+
this corrected SPEC-07
```

supersede that stale next-phase label.

After independent review and closure, update PROJECT_CONTEXT in a separate documentation alignment step.

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

On this Windows environment, race verification uses the already-installed MSYS2 UCRT64 GCC if needed:

```text
C:\msys64\ucrt64\bin\gcc.exe
```

Frontend:

```bash
cd frontend
bun install --frozen-lockfile
bun run check
bun test
bun run build
```

Do not upgrade dependencies.

---

# 52. Acceptance criteria

SPEC-07 correction is implementation-ready for review only when:

```text
old Ticket Detail inspector removed
corrected SPEC-07 file is canonical
click/select ticket opens Chat shell
GET /tickets/:id remains real data source

desktop:
compact docked right Chat
table remains visible
panel bounded in width/height
only conversation area scrolls
composer/actions pinned

mobile:
full-width Chat
Back preserves list
conversation scrolls
bottom controls remain reachable

summary:
title once
status
requester
location.name only
Owner = accepted_by
created timestamp

conversation:
real created-request system activity
real accepted activity only when persisted
no fake user messages

shell:
Chats active
Checklist visible but disabled
composer visible but non-submittable
Accept/Assign/Close visible but non-operational

regression:
Open/Closed
Load More
New Request
Location picker
auth
Light/Dark
stale-response protection

quality:
backend tests pass
race test passes
frontend check/tests/build pass
no migration
no runtime mock data
no unrelated changes staged
```

---

# 53. Commit discipline

Review staged diff carefully.

Implementation commit may contain:

```text
corrected SPEC-07
frontend Chat-shell correction
frontend tests
minimal selected-ticket naming/state cleanup
```

Backend should change only if a real defect is found.

Do not update PROJECT_CONTEXT yet.

Suggested commit:

```text
fix: replace ticket detail with chat shell
```

Do not mark SPEC-07 CLOSED until independent review and real-browser verification.

---

# 54. Final report

Return:

```text
Canonical SPEC replacement
Old Ticket Detail removal
Chat component/state naming
Desktop Chat layout
Mobile Chat layout
Ticket summary
System activities
Composer/action shell
Checklist shell
Selected-ticket API preservation
Stale-response behavior
Tests
Backend verification
Frontend verification
Manual verification
Migration/database check
Files changed
Pre-existing changes preserved
Commit SHA
```

Explicitly confirm:

```text
The old Ticket Detail user-facing inspector is removed.
Clicking/selecting a ticket now opens the Chat shell.
GET /api/v1/tickets/:id still reads real PostgreSQL data.
No chat message persistence was implemented.
Accept/Assign/Close remain non-mutating shells.
Checklist remains non-persistent.
No migration 0021 was created.
No runtime mock chat/ticket data was added.
Unrelated working-tree changes were not staged.
```

End with:

```text
Corrected SPEC-07 implementation is ready for independent review; it is not marked CLOSED yet.
```
