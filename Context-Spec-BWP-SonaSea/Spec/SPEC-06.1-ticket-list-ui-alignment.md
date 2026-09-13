# SPEC-06.1 — Ticket List UI Alignment

> Project: **BWP SonaSea**
> Repository: `thienty1207/BWP-Staff`
> Baseline: `main` at or after `3008c3b970d2fb529d745bc07cb5d564150f52c9`
>
> Closed before this SPEC: SPEC-01, SPEC-02, SPEC-03, SPEC-04, SPEC-04.1, SPEC-05, SPEC-05.1, SPEC-05.2, SPEC-06.

## 1. Goal

Align the Ticket List UI with the approved product direction. The current list is functionally correct, but its information hierarchy is wrong for the intended product and the mobile cards waste too much vertical space.

SPEC-06.1 must:

- redesign the desktop ticket list to follow the approved compact table hierarchy;
- redesign the mobile ticket list into compact summary cards;
- remove Ticket ID from visible ticket-list UI;
- expose the real ticket `description` in the list API;
- represent **Requester** and **Owner** with the correct business meaning;
- move Priority away from Title and place it under Due Date on desktop;
- keep Action controls out of scope until their backend operations exist;
- keep Ticket Detail / Chat out of scope until SPEC-07+;
- use only real PostgreSQL-backed data.

No runtime mock data.

---

## 2. Locked business meaning

### 2.1 Requester

**Requester** is the user who originally created/submitted the ticket.

Source of truth:

```text
tickets.requester_id
```

Requester does not change after Accept/Assign/Close.

### 2.2 Owner

**Owner** is the **first person who accepts the ticket**.

Source of truth:

```text
tickets.accepted_by
```

Rules:

```text
pending  → Owner = —
accepted → Owner = accepted_by
closed   → Owner remains accepted_by
```

Later assignment must not redefine Owner. Assigned users/departments are separate concepts.

---

# PART A — DESKTOP TICKET LIST

## 3. Desktop columns

Replace the current visible desktop columns with this exact order:

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

Do **not** show these columns in SPEC-06.1:

```text
Ticket ID
Department
Assignment
Action
```

The table should follow the provided Sara reference for information hierarchy and compactness while retaining BWP SonaSea's own Light/Dark themes. Do not clone Sara's legacy colors/styles.

## 4. Requester display

Preferred compact label:

```text
Full Name (DepartmentCode)
```

Example:

```text
Lưu Văn Trường (REC)
```

If the ticket response does not expose the requester's department code, extend the response with the smallest explicit safe field required. Do not reconstruct this on the client with extra lookups. Do not mock the code.

## 5. Location

Display the real location name.

```text
BWP-Recreation Office
```

If no location:

```text
—
```

Do not show location IDs.

## 6. Title

Show the real ticket title as the main request title.

- No Ticket ID prefix.
- No Priority badge below the title.
- Do not merge Title and Description.

## 7. Description

The desktop list must show the real ticket description. Extend:

```http
GET /api/v1/tickets
```

to return `tickets.description`.

Rules:

```text
NULL / blank → —
non-null     → compact preview
```

Keep Description bounded to roughly 1–2 visual lines using line-clamp/ellipsis or equivalent. Do not let one ticket expand the whole table row excessively. No fabricated text.

## 8. Status

Status remains exactly:

```text
Pending
Accepted
Closed
```

Backed by existing values:

```text
pending
accepted
closed
```

Keep a compact status badge.

## 9. Owner display

Owner uses only `accepted_by`.

Preferred label:

```text
Full Name (DepartmentCode)
```

Example:

```text
Hồ Thiên Tỷ (IT)
```

For pending tickets:

```text
—
```

If the accepted user's department code is not currently exposed, extend the response with the smallest explicit safe field required.

Never derive Owner from:

- assigned users;
- assigned departments;
- requester;
- current logged-in user.

## 10. Created On

Display the real creation date/time compactly. Example:

```text
Sep 12, 2026
4:43 PM
```

Equivalent locale-aware formatting consistent with the current app is acceptable. Desktop should not use relative-only time as the sole display.

## 11. Due Date + Priority

Priority must move out of the Title area.

Required desktop structure:

```text
Sep 12, 2026, 6:43 PM
        Priority
```

The Priority badge must be:

- below the due date/time;
- centered inside the Due Date cell.

If `priority=false`, do not show the badge.

If `due_at=NULL` and `priority=true`, show:

```text
—
Priority
```

with Priority still centered below the due-date area.

## 12. No Action column yet

Do not render:

```text
Accept
Assign
Close
```

Do not render disabled/fake action buttons just to imitate the reference. Action controls will be introduced only when the corresponding backend operations exist.

---

# PART B — MOBILE TICKET LIST

## 13. Remove the current tall mobile layout

The existing mobile ticket card is too tall because it stacks developer-style fields such as:

```text
DEPARTMENT
LOCATION
REQUESTER
ASSIGNMENT
CREATED
DUE
```

Remove this structure.

Do not show Ticket ID on mobile.

Do not reproduce the desktop table by stacking every column vertically.

## 14. Compact mobile card

Use a compact summary-card layout based on the provided Sara mobile reference.

Target hierarchy:

```text
[icon] Title                         Status

       Location                     Owner
       Requester                    Created time
       Description preview...
```

The exact CSS may adapt to BWP SonaSea, but the card must remain compact and readable.

Required visible information:

- Title;
- Location;
- Requester;
- short Description preview when present;
- Status;
- Owner when accepted;
- Created time;
- compact Priority indicator.

Do not show:

- Ticket ID;
- Department as a separate stacked field;
- Assignment;
- Action buttons;
- a large Due block;
- full Description.

The objective is that multiple tickets can be seen on one mobile screen without excessive scrolling.

## 15. Mobile identity labels

Requester:

```text
Full Name (DepartmentCode)
```

Owner:

```text
Full Name (DepartmentCode)
```

For pending tickets Owner is logically `—`. The Owner area may be visually omitted while Pending if that improves density, provided Pending status is clear.

## 16. Mobile Description preview

- maximum roughly 1–2 lines;
- use line-clamp/ellipsis;
- if description is absent, do not reserve blank vertical space;
- do not show full Description in the list.

Full ticket content belongs to Ticket Detail later.

## 17. Mobile timestamp

Use a compact created-time display. A relative format such as `1 day ago` is acceptable if implemented cleanly without adding a dependency solely for this purpose. Otherwise keep a compact date/time formatter.

## 18. Mobile Priority

Priority must stay compact. Do not add a large dedicated row. It must not appear under Title as in the current UI.

Desktop remains stricter: Priority must be below Due Date and centered.

---

# PART C — FUTURE TICKET DETAIL CONTRACT

## 19. Mobile card → Ticket Detail is locked for the next phase

Approved future flow:

```text
Compact Ticket List
→ user selects a ticket
→ dedicated Ticket Detail / Chat screen
```

This direction is based on the provided mobile detail reference.

However, SPEC-06.1 must **not** implement:

- Ticket Detail API;
- Ticket Detail route/navigation;
- fake detail page;
- chat;
- checklist;
- placeholder ticket content;
- dead links.

Do not pretend the card opens detail until the real feature exists. This becomes the design baseline for **SPEC-07 — Ticket Detail Read**.

---

# PART D — BACKEND RESPONSE ALIGNMENT

## 20. Extend the existing ticket-list response only as required

Required new data:

```text
description
```

Also expose the minimum identity metadata needed to render:

```text
Requester: Full Name (DepartmentCode)
Owner:     Full Name (DepartmentCode)
```

Preferred explicit shape:

```json
{
  "requester": {
    "id": 10,
    "full_name": "Lưu Văn Trường",
    "department_code": "REC"
  },
  "accepted_by": {
    "id": 20,
    "full_name": "Hồ Thiên Tỷ",
    "department_code": "IT"
  }
}
```

`accepted_by` remains nullable. An equivalent small explicit shape is acceptable if cleaner in the existing code.

Do not expose unnecessary user/profile fields. Never expose password/session/token material.

## 21. SQL/performance rules

Preserve the current list architecture and performance behavior.

- No N+1.
- Add `description` to the existing base ticket query.
- If requester/owner department codes require joins, join them in the same base query.
- Do not issue per-ticket user/department lookup queries.
- Keep keyset pagination unchanged.
- Keep `LIMIT + 1` behavior.
- Keep the existing bulk assignment queries even though Assignment is no longer displayed; do not broadly refactor them in this SPEC unless a direct compatibility issue requires it.
- No `SELECT *`.
- No OFFSET pagination.
- No Redis/cache.

## 22. Create Ticket compatibility

`POST /api/v1/tickets` must continue to work.

Do not change:

```text
requester from authenticated session
status = pending
ticket_activity = created
transaction atomic
no automatic POST retry
```

If a shared Ticket response model is extended, keep Create Ticket response compatible with frontend parsing.

---

# PART E — FRONTEND STRUCTURE

## 23. Keep components focused

Do not turn root `+page.svelte` into a giant component.

A small focused split for desktop/mobile ticket list rendering is allowed if it improves readability. Avoid component explosion and generic abstractions.

## 24. Theme

The new list must work correctly in both:

```text
Light
Dark
```

Use current BWP SonaSea theme tokens. No login redesign in this SPEC.

---

# PART F — TESTS

## 25. Backend tests

Extend PostgreSQL-backed tests to cover:

- list response contains real description;
- null description remains null/safe;
- requester department code is correct;
- accepted owner department code is correct;
- pending accepted_by remains null;
- accepted owner comes from accepted_by;
- assignment does not become Owner;
- pagination/auth behavior remains correct;
- no per-ticket identity lookup pattern is introduced.

Use only official `backend/.env` and `DATABASE_URL`. Use isolated temporary PostgreSQL schemas. Do not introduce `.env.test`, `DATABASE_TEST_URL`, `TEST_DATABASE_URL`, `TEST_*`, or `APP_ENV=test`.

## 26. Frontend tests

Add focused tests for:

- ticket parser accepts description;
- requester identity department code;
- accepted_by identity department code;
- null accepted_by;
- visible UI does not show Ticket ID;
- desktop column order is exact;
- desktop does not show Department / Assignment / Action columns;
- Priority is no longer under Title;
- mobile list uses compact summary hierarchy;
- old stacked mobile field layout is absent;
- Description preview exists;
- no fake Ticket Detail navigation;
- no runtime mock data.

Do not add a large UI test framework solely for SPEC-06.1.

---

# PART G — MANUAL RESPONSIVE VERIFICATION

## 27. Verify these viewports

```text
360
390
430
768
1280
1920
```

Desktop must verify:

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

and:

```text
no ID
no Department column
no Assignment
no Action
Description compact
Owner correct
Priority below Due Date
```

Mobile must verify:

```text
compact cards
no ID
no old stacked field layout
multiple tickets visible
no horizontal overflow
Title/Location/Requester/Description/Status/Owner/time readable
```

Verify both Light and Dark.

---

# PART H — NON-GOALS

## 28. Out of scope

Do not implement:

- Ticket Detail backend;
- Ticket Detail route/navigation;
- Accept;
- Assign;
- Close;
- Chat;
- Checklist;
- Attachments;
- Staff Meal;
- Announcements;
- Report;
- Settings;
- Admin UI;
- notifications;
- WebSocket;
- polling;
- Redis;
- cache;
- search;
- manual Refresh;
- migration `0021`.

No database migration is expected.

---

# PART I — DEFINITION OF DONE

## 29. SPEC-06.1 closes only when

- desktop list matches the approved business-facing hierarchy;
- mobile list is compact;
- Ticket ID is absent from visible list UI;
- Description comes from PostgreSQL;
- Requester means ticket creator;
- Owner means first accepter / `accepted_by`;
- pending tickets have no Owner;
- assignment does not redefine Owner;
- Priority is below Due Date on desktop;
- Action is absent;
- old mobile stacked-detail card is removed;
- no fake Ticket Detail exists;
- keyset pagination remains unchanged;
- no N+1 is introduced;
- no migration is added;
- no runtime mock data exists;
- Light/Dark both work;
- backend/frontend verification passes.

Next expected phase:

```text
SPEC-07 — Ticket Detail Read
```
