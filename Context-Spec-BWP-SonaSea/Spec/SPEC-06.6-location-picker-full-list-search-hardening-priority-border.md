# SPEC-06.6 — Location Picker Full-List Fidelity, Literal Search Hardening, and Priority Title Border

> Project: **BWP SonaSea**
>
> Repository: `thienty1207/BWP-Staff`
>
> Baseline commit:
>
> `eac284b24c07f9c3849c68b3b8085ee6d13d1964`
>
> SPEC-06.6 is a bounded correction/hardening pass over the SPEC-06.5 implementation.
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
> - SPEC-05.2 ✅ CLOSED
> - SPEC-06 ✅ CLOSED
> - SPEC-06.1 ✅ CLOSED
> - SPEC-06.2 ✅ CLOSED
> - SPEC-06.3 ✅ CLOSED
> - SPEC-06.4 ✅ CLOSED
> - SPEC-06.5 implementation present; closure pending the corrections in this SPEC
>
> No SPEC-07 work is allowed.

---

## 1. Purpose

SPEC-06.6 corrects four concrete issues found after review and visual testing of SPEC-06.5:

1. The Location picker currently loads only a small bounded subset instead of exposing the full active location list like the SARA reference.
2. Location search currently uses SQL `LIKE`, so user-entered `%` and `_` act as wildcard operators instead of literal characters.
3. The picker can theoretically select a stale previously-loaded result from the keyboard while a newer search is still loading.
4. Priority ticket Title styling is wrong: the Title is currently red text. The user wants normal Title text with a red content-sized border around it.

This is a targeted correction. Do not redesign unrelated UI.

---

# PART A — FULL LOCATION LIST FIDELITY

## 2. Visual contract

The SARA reference is binding:

```text
full location catalogue is available
→ only a manageable vertical slice is visible
→ user scrolls inside the result panel
→ user may type to search/filter
```

Important:

```text
FULL DATASET AVAILABLE
≠
SHOW THE FULL DATASET HEIGHT AT ONCE
```

The result popup remains short and scrollable.

---

## 3. Empty-query behavior

When the Location picker opens with no search text:

```text
all active locations must be available in the list
```

The frontend must not artificially request only 10 rows.

Required request:

```http
GET /api/v1/locations
```

Do not add:

- Load More;
- pagination buttons;
- infinite scrolling;
- virtualization.

At the current catalogue size, one complete scrollable list is intentional.

---

## 4. Search behavior

When the user types, search must expose the full matching result set, ranked by relevance.

Required request shape:

```http
GET /api/v1/locations?q=<query>
```

The New Request picker must not inject `limit=10`.

All matching active locations must remain browseable in the bounded result panel.

---

## 5. Backend limit compatibility

Keep optional backend parameter:

```text
limit
```

for explicit callers.

Rules:

```text
limit absent  → no artificial result cap
limit present → validate 1..30 and apply it
```

This applies to both:

```http
GET /api/v1/locations
GET /api/v1/locations?q=oasis
```

If the caller explicitly sends `limit=10`, return at most 10.

Do not silently turn an omitted limit into 10.

---

## 6. Ordering

For empty query:

```sql
ORDER BY name ASC, id ASC
```

For non-empty query preserve:

```text
0 exact name
1 full-name prefix
2 token/word prefix
3 contains
4 name ASC
5 id ASC
```

No fuzzy/AI search.
No pg_trgm.
No migration.

---

## 7. Popup containment

Full-list availability must not recreate the old giant dropdown.

Desktop:

```text
bounded max-height
overflow-y: auto
```

Mobile:

```text
shorter viewport-aware max-height
overflow-y: auto
inside the New Request modal
```

At 360/390/430 widths:

- only several rows should be visible at once;
- the user scrolls inside the result list;
- no horizontal overflow;
- the rest of the form remains reachable.

---

# PART B — LITERAL SEARCH HARDENING

## 8. Current hidden bug

Current search uses logic equivalent to:

```sql
LOWER(name) LIKE '%' || LOWER($1) || '%'
```

Therefore `%` and `_` have wildcard meaning.

This is not SQL injection, but it is incorrect product behavior.

---

## 9. Required literal semantics

User input must be treated literally.

Examples:

```text
q=%
```

must search for a literal `%`.

```text
q=_
```

must search for a literal `_`.

Neither may behave as a SQL wildcard.

Backslash/punctuation must not accidentally alter search meaning.

---

## 10. Preferred SQL approach

Prefer avoiding `LIKE` wildcard interpretation:

```sql
LOWER(name) = LOWER($1)
starts_with(LOWER(name), LOWER($1))
starts_with(token, LOWER($1))
strpos(LOWER(name), LOWER($1)) > 0
```

Equivalent safe SQL is acceptable.

If `LIKE` remains, correctly escape `%`, `_`, and `\` with an explicit escape character.

Keep:

- parameterized SQL;
- active-only;
- one SQL query;
- deterministic ranking;
- no N+1;
- no filtering/ranking in Go.

---

# PART C — STALE KEYBOARD-SELECTION HARDENING

## 11. Current race

Possible sequence:

```text
old results loaded
→ old result highlighted
→ user types new query
→ UI displays Searching…
→ old results/highlight still exist in state
→ user presses Enter
```

An old hidden result must never be selected.

---

## 12. Required invariant

As soon as a new search is scheduled:

```text
highlightedLocationIndex = -1
```

While:

```text
locationSearchLoading === true
```

result keyboard navigation/selection is disabled.

At minimum:

```text
ArrowDown ignored
ArrowUp ignored
Enter cannot select a result
```

Clearing old results while loading is also acceptable.

Key invariant:

> A result from query N must never be selectable while query N+1 is pending.

---

## 13. Latest response wins

Keep the existing request-sequence protection.

A stale HTTP response must never replace a newer result set.

Closing/clearing the picker must continue invalidating pending responses.

No polling.
No automatic retry.

---

# PART D — PRIORITY TITLE BORDER

## 14. Current UI defect

Wrong:

```text
priority = red Title text
```

Required:

```text
priority = normal Title text + red border around the Title content box
```

The user reference screenshots are binding.

---

## 15. Priority visual contract

For `ticket.priority === true`:

```text
normal title text color
1px red/danger border
small border radius
tight padding
content-sized box
max-width: 100%
safe wrapping
```

Conceptually:

```css
display: inline-block;
width: fit-content;
max-width: 100%;
box-sizing: border-box;
border: 1px solid var(--danger);
border-radius: small;
padding: tight;
```

Do not stretch the border across the full table cell.

Do not use fixed width or fixed height.

---

## 16. Content-sized behavior

Short Title:

```text
máy tính lễ tân mất mạng
```

→ compact short border.

Long Title:

```text
Promotion Update Support / Hỗ Trợ Cập Nhật Quảng Cáo
```

→ wider border; if the available width is reached, text wraps and the border grows vertically around the wrapped text.

The border follows the rendered Title box.

---

## 17. Text color

Do not make priority text red.

Light theme:

```text
normal dark/black Title text
```

Dark theme:

```text
normal high-contrast theme Title text
```

Do not force literal black on a dark theme.

The only priority-specific color is the border.

---

## 18. Priority badge

Do not reintroduce visible text:

```text
Priority
```

on the ticket list.

Priority indication is the red Title border only.

Apply on:

- desktop Title;
- mobile compact-card Title.

---

# PART E — KEEP EXISTING 06.5 WORK

## 19. Date/time formatting

Do not change the SPEC-06.5 stacked timestamp design:

```text
Created On:
date
time

Due Date:
date
time
```

Null due date remains:

```text
—
```

---

## 20. Database/master data

Do not alter:

```text
155 BWP fixture rows
141 96 Villas fixture rows
department fixtures
location names/codes
fixture markers
```

No migration `0021`.

No schema change.

---

# PART F — TESTS

## 21. Backend tests

Add/update tests proving:

### No-limit behavior

```http
GET /api/v1/locations
```

returns all active locations.

```http
GET /api/v1/locations?q=<query>
```

returns all active matches when `limit` is absent.

Explicit limits still work:

```text
1
10
30
```

Invalid values remain rejected:

```text
0
31
non-number
```

### Literal wildcard behavior

Create isolated rows containing literal:

```text
%
_
```

Verify:

```text
q=%
q=_
```

match literal characters only.

They must not match arbitrary rows.

### Ranking regression

Preserve exact → prefix → token-prefix → contains → name/id ordering.

Also keep:

- auth required;
- active-only;
- trim;
- case-insensitive;
- 100-rune query cap;
- no duplicate results;
- positive real IDs.

---

## 22. Frontend picker tests

Tests must prove:

- initial picker load no longer sends `limit=10`;
- search request no longer injects `limit=10`;
- shared API may still support an explicit limit;
- new search resets highlighted index;
- Enter/Arrow selection is blocked while loading;
- stale-response protection remains;
- visible rows show `location.name` only;
- selected value remains real `location.id`;
- clearing means `location_id = null`;
- `No location` remains supported.

Do not add a new UI framework.

---

## 23. Priority CSS tests

Tests must verify the actual CSS contract:

- priority class has red/danger border;
- priority class does not set text color to danger/red;
- content-sized display/width behavior exists;
- max-width protects long titles;
- desktop Title gets the priority class;
- mobile Title gets the priority class;
- normal Title has no priority border;
- no visible Priority badge exists.

Do not only test that a class name exists.

---

# PART G — MANUAL VERIFICATION

## 24. Location browsing

Using real backend/PostgreSQL:

```text
New Request → Location
```

With empty search:

- confirm the list is not capped at 10;
- confirm the full active catalogue is reachable by internal scrolling;
- confirm the popup remains bounded.

---

## 25. Search verification

Test:

```text
oasis
server
96 v
%
_
```

Confirm `%` and `_` behave literally.

---

## 26. Stale keyboard verification

Test:

```text
search query A
highlight an A result
quickly type query B
press Enter before B finishes
```

Expected:

```text
old A result is NOT selected
```

After B finishes, only B results may be selected.

---

## 27. Priority visual verification

Verify:

### short priority Title
- normal text color;
- compact red border sized around text.

### long priority Title
- border expands naturally;
- wraps when needed;
- border encloses wrapped text.

Check desktop/mobile and Light/Dark when tooling allows.

No Priority badge.

---

## 28. Responsive verification

Verify representative sizes, and exact viewports when available:

```text
360
390
430
768
1280
1920
```

If exact viewport automation is unavailable, report that honestly.

Do not claim unperformed checks.

---

# PART H — ENGINEERING GUARDRAILS

## 29. Full list is intentional

A few hundred location rows is an acceptable full browse list for this product stage.

Do not prematurely add:

- Redis;
- pagination;
- infinite scroll;
- virtualization;
- Elasticsearch;
- fuzzy-search infrastructure.

Future high-concurrency optimization remains a separate later performance phase.

---

## 30. Scope guard

Do NOT implement:

- new location data;
- department changes;
- Ticket Detail;
- Accept;
- Assign;
- Close;
- Chat;
- Checklist;
- attachments;
- Report;
- Settings;
- Admin UI;
- location hierarchy/grouping;
- fuzzy AI search;
- pg_trgm;
- Redis/cache;
- WebSocket;
- polling;
- migrations;
- SPEC-07;
- PROJECT_CONTEXT update;
- runtime mock data.

Keep the patch narrow.

---

## 31. Definition of Done

SPEC-06.6 is complete only when:

- empty Location picker exposes the full active catalogue;
- searched lookup without explicit limit is not capped at 10;
- popup remains bounded and internally scrollable;
- mobile does not overflow;
- `%` and `_` search literally;
- ranking remains correct;
- stale old results cannot be keyboard-selected while a newer search is pending;
- latest-response-wins remains;
- priority Title text is normal color;
- priority Title has a red content-sized border;
- short and long Title borders size naturally;
- no Priority badge;
- no fixture/schema/migration changes;
- backend/frontend verification passes;
- no SPEC-07 work.

After this passes review:

```text
SPEC-06.5 ✅ CLOSED
SPEC-06.6 ✅ CLOSED
```

Then update `PROJECT_CONTEXT.md` before SPEC-07.
