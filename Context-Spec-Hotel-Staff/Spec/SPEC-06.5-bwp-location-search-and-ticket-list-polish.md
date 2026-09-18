# SPEC-06.5 — BWP Location Master Data, Searchable Location Picker, and Ticket List Date/Priority Polish

> Project: **BWP SonaSea**
>
> Repository: `thienty1207/BWP-Staff`
>
> Dependency satisfied: **SPEC-06.4 ✅ CLOSED**
>
> This SPEC supersedes the earlier SPEC-06.5 draft. The earlier 148-row BWP dataset is preserved and extended with 7 additional rows, for **155 BWP fixture rows total**.
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

> **Supersession note — current behavior:** SPEC-06.6 supersedes the earlier
> default Location limit of 10 and the red/bold Priority Title treatment.
>
> Current canonical behavior:
> - omitted Location limit → no artificial cap;
> - explicit Location limit → 1..30;
> - literal `%` / `_` search semantics;
> - normal Title text;
> - red content-sized priority border;
> - no visible Priority badge.

---

## 1. Goal

SPEC-06.5 has three bounded goals:

1. complete the **BWP main-property location master data** in PostgreSQL;
2. replace the slow long Location dropdown with a **searchable, bounded, scrollable location picker** backed by the real lookup API;
3. align Ticket List **Created On / Due Date / Priority presentation** with the supplied reference screenshots.

This SPEC intentionally touches:

```text
Database reference data / seed
Backend location lookup
Frontend New Request Location UI
Frontend desktop/mobile Ticket List presentation
Backend + frontend tests
```

It does **not** implement Ticket Detail or later ticket actions.

---

# PART A — BWP LOCATION MASTER DATA

## 2. Dataset ownership

SPEC-06.3 owns the **96 Villas** dataset:

```text
42 named 96 Villas areas
99 room rows 1001..1099
141 total 96 Villas fixture rows
```

SPEC-06.5 owns a separate **BWP main-property** dataset.

After SPEC-06.5:

```text
BWP fixture rows       = 155
96 Villas fixture rows = 141
```

Do not merge these fixture ownership domains.

---

## 3. Source-of-truth rule

The user-supplied screenshots are the source of truth.

Preserve every BWP display name **exactly as shown**, including:

- capitalization;
- punctuation;
- spaces;
- apostrophes;
- hyphen spacing;
- apparent spelling mistakes.

Do not silently correct the source.

Examples that must remain exactly:

```text
BWP - Basemant cold storage
BWP - Capentry work
BWP - HK Pantry Wing 2-2nd Floorr
BWP - HK Pantry Wing 2-3nd Floor
BWP - HK Pantry Wing 3-3th Floor
BWP - Pastry kitchen.
BWP - Toilet Staff leve 1
BWP - Wastewater treatmant room
BWP- MSB room
BWP-Kitchen Office
ENG-OFFICE
FIN DOCUMENT STORE
MasterKey Audit
Server Room- 96 Villas
```

Similar-looking source labels remain distinct unless the user explicitly requests normalization later.

---

## 4. Exact BWP dataset

Use these stable code/name pairs exactly:

| Code | Display Name |
|---|---|
| `BWP-AREA-001` | BOD Office |
| `BWP-AREA-002` | BWP - Asian kitchen |
| `BWP-AREA-003` | BWP - Back Office/FO |
| `BWP-AREA-004` | BWP - Bakery kitchen |
| `BWP-AREA-005` | BWP - Ballroom |
| `BWP-AREA-006` | BWP - Basemant cold storage |
| `BWP-AREA-007` | BWP - Beach Bar |
| `BWP-AREA-008` | BWP - Bell Desk |
| `BWP-AREA-009` | BWP - BOH Essence |
| `BWP-AREA-010` | BWP - Boiler room |
| `BWP-AREA-011` | BWP - BTS room |
| `BWP-AREA-012` | BWP - Buffet counter |
| `BWP-AREA-013` | BWP - Canteen |
| `BWP-AREA-014` | BWP - Capentry work |
| `BWP-AREA-015` | BWP - cold kitchen |
| `BWP-AREA-016` | BWP - Cooling tower |
| `BWP-AREA-017` | BWP - Cview Bar |
| `BWP-AREA-018` | BWP - Cview Kitchen |
| `BWP-AREA-019` | BWP - Eng store 2B |
| `BWP-AREA-020` | BWP - Eng store 2C |
| `BWP-AREA-021` | BWP - Eng store 3B |
| `BWP-AREA-022` | BWP - Eng store 3C |
| `BWP-AREA-023` | BWP - Eng store 4B |
| `BWP-AREA-024` | BWP - Eng store 4C |
| `BWP-AREA-025` | BWP - Eng store 5B |
| `BWP-AREA-026` | BWP - Eng store 5C |
| `BWP-AREA-027` | BWP - Eng store 6B |
| `BWP-AREA-028` | BWP - Eng store 6C |
| `BWP-AREA-029` | BWP - Eng store 7B |
| `BWP-AREA-030` | BWP - Eng store 7C |
| `BWP-AREA-031` | BWP - Eng store 8B |
| `BWP-AREA-032` | BWP - Eng store 8C |
| `BWP-AREA-033` | BWP - Eng store 9B |
| `BWP-AREA-034` | BWP - Eng store 9C |
| `BWP-AREA-035` | BWP - EPS room |
| `BWP-AREA-036` | BWP - Essence Kitchen |
| `BWP-AREA-037` | BWP - Essence restaurant |
| `BWP-AREA-038` | BWP - Essences Bar |
| `BWP-AREA-039` | BWP - European kitchen |
| `BWP-AREA-040` | BWP - Fan room |
| `BWP-AREA-041` | BWP - FB Office |
| `BWP-AREA-042` | BWP - Female Locker |
| `BWP-AREA-043` | BWP - FO Pantry |
| `BWP-AREA-044` | BWP - FO Reception |
| `BWP-AREA-045` | BWP - Generator room |
| `BWP-AREA-046` | BWP - GRO Counter |
| `BWP-AREA-047` | BWP - Guest elevator wing 1 |
| `BWP-AREA-048` | BWP - Guest elevator wing 3 |
| `BWP-AREA-049` | BWP - Gym |
| `BWP-AREA-050` | BWP - HK OFFICE |
| `BWP-AREA-051` | BWP - HK Pantry Wing 2-2nd Floor |
| `BWP-AREA-052` | BWP - HK Pantry Wing 2-2nd Floorr |
| `BWP-AREA-053` | BWP - HK Pantry Wing 2-3nd Floor |
| `BWP-AREA-054` | BWP - HK Pantry Wing 2-4th Floor |
| `BWP-AREA-055` | BWP - HK Pantry Wing 2-5th Floor |
| `BWP-AREA-056` | BWP - HK Pantry Wing 2-6th Floor |
| `BWP-AREA-057` | BWP - HK Pantry Wing 2-7th Floor |
| `BWP-AREA-058` | BWP - HK Pantry Wing 2-8th Floor |
| `BWP-AREA-059` | BWP - HK Pantry Wing 2-9th Floor |
| `BWP-AREA-060` | BWP - HK Pantry Wing 3-3th Floor |
| `BWP-AREA-061` | BWP - HK Pantry Wing 3-4th Floor |
| `BWP-AREA-062` | BWP - HK Pantry Wing 3-5th Floor |
| `BWP-AREA-063` | BWP - HK Pantry Wing 3-6th Floor |
| `BWP-AREA-064` | BWP - HK Pantry Wing 3-7th Floor |
| `BWP-AREA-065` | BWP - HK Pantry Wing 3-8th Floor |
| `BWP-AREA-066` | BWP - HK Pantry Wing 3-9th Floor |
| `BWP-AREA-067` | BWP - HK Store 2A1 |
| `BWP-AREA-068` | BWP - HK Store 2A3 |
| `BWP-AREA-069` | BWP - HK Store 3A1 |
| `BWP-AREA-070` | BWP - HK Store 3A3 |
| `BWP-AREA-071` | BWP - HK Store 4A1 |
| `BWP-AREA-072` | BWP - HK Store 4A3 |
| `BWP-AREA-073` | BWP - HK Store 5A1 |
| `BWP-AREA-074` | BWP - HK Store 5A3 |
| `BWP-AREA-075` | BWP - HK Store 6A1 |
| `BWP-AREA-076` | BWP - HK Store 6A3 |
| `BWP-AREA-077` | BWP - HK Store 7A1 |
| `BWP-AREA-078` | BWP - HK Store 7A3 |
| `BWP-AREA-079` | BWP - HK Store 8A1 |
| `BWP-AREA-080` | BWP - HK Store 8A3 |
| `BWP-AREA-081` | BWP - HK Store 9A1 |
| `BWP-AREA-082` | BWP - HK Store 9A3 |
| `BWP-AREA-083` | BWP - Ice machine room |
| `BWP-AREA-084` | BWP - In front of Ballroom |
| `BWP-AREA-085` | BWP - Kid's Club |
| `BWP-AREA-086` | BWP - Kid's Playground |
| `BWP-AREA-087` | BWP - Kitchen office |
| `BWP-AREA-088` | BWP - Lagoon |
| `BWP-AREA-089` | BWP - Lagoon pump room |
| `BWP-AREA-090` | BWP - Lagoon Swimming Pool |
| `BWP-AREA-091` | BWP - Laundry room |
| `BWP-AREA-092` | BWP - Lobby |
| `BWP-AREA-093` | BWP - Main kitchen |
| `BWP-AREA-094` | BWP - Main Swimming Pool |
| `BWP-AREA-095` | BWP - Mainpool MEP Room |
| `BWP-AREA-096` | BWP - Male Locker |
| `BWP-AREA-097` | BWP - Medium voltage room |
| `BWP-AREA-098` | BWP - Meeting room |
| `BWP-AREA-099` | BWP - Oasis bar |
| `BWP-AREA-100` | BWP - Oasis Bathroom |
| `BWP-AREA-101` | BWP - Oasis pool |
| `BWP-AREA-102` | BWP - Oasis pool bar |
| `BWP-AREA-103` | BWP - Oasis Swimming Pool |
| `BWP-AREA-104` | BWP - Operator Room |
| `BWP-AREA-105` | BWP - Outside Lobby |
| `BWP-AREA-106` | BWP - PA Store-1st Floor |
| `BWP-AREA-107` | BWP - PA Store-M Floor |
| `BWP-AREA-108` | BWP - Pastry kitchen. |
| `BWP-AREA-109` | BWP - PS/DS room |
| `BWP-AREA-110` | BWP - Pump filter room |
| `BWP-AREA-111` | BWP - Recreation Store |
| `BWP-AREA-112` | BWP - RES Office |
| `BWP-AREA-113` | BWP - Romantic Dinner |
| `BWP-AREA-114` | BWP - Rooftop Fan |
| `BWP-AREA-115` | BWP - Spa |
| `BWP-AREA-116` | BWP - Staff restroom |
| `BWP-AREA-117` | BWP - Staffhouse 1 |
| `BWP-AREA-118` | BWP - Staffhouse 2 |
| `BWP-AREA-119` | BWP - Toilet Ballroom |
| `BWP-AREA-120` | BWP - Toilet Basement |
| `BWP-AREA-121` | BWP - Toilet C view |
| `BWP-AREA-122` | BWP - Toilet Essence |
| `BWP-AREA-123` | BWP - Toilet Lobby |
| `BWP-AREA-124` | BWP - Toilet Oasis |
| `BWP-AREA-125` | BWP - Toilet Staff leve 1 |
| `BWP-AREA-126` | BWP - Toilet Staff leve M |
| `BWP-AREA-127` | BWP - Transformer room |
| `BWP-AREA-128` | BWP - Uniform room |
| `BWP-AREA-129` | BWP - Wastewater treatmant room |
| `BWP-AREA-130` | BWP - Workshop |
| `BWP-AREA-131` | BWP (16 Villas & B- 1st floor) |
| `BWP-AREA-132` | BWP Codotel (M- Rooftop floor) |
| `BWP-AREA-133` | BWP- MSB room |
| `BWP-AREA-134` | BWP- Oasis Kitchen |
| `BWP-AREA-135` | BWP-Butchery Area |
| `BWP-AREA-136` | BWP-Kitchen Office |
| `BWP-AREA-137` | BWP-Receiving Area |
| `BWP-AREA-138` | BWP-Recreation Office |
| `BWP-AREA-139` | BWP-Steward Area |
| `BWP-AREA-140` | CCTV Room |
| `BWP-AREA-141` | ENG-OFFICE |
| `BWP-AREA-142` | Executive Office |
| `BWP-AREA-143` | FIN DOCUMENT STORE |
| `BWP-AREA-144` | FIN OFFICE |
| `BWP-AREA-145` | GENERAL STORE |
| `BWP-AREA-146` | HR Office |
| `BWP-AREA-147` | IT Office |
| `BWP-AREA-148` | MasterKey Audit |
| `BWP-AREA-149` | Nurse Office |
| `BWP-AREA-150` | RECEIVING OFFICE |
| `BWP-AREA-151` | Sales Room |
| `BWP-AREA-152` | Security Office |
| `BWP-AREA-153` | Server Room |
| `BWP-AREA-154` | Server Room- 96 Villas |
| `BWP-AREA-155` | Staffhouse |

The final 7 rows are the additional locations supplied after the first BWP screenshot set:

```text
Nurse Office
RECEIVING OFFICE
Sales Room
Security Office
Server Room
Server Room- 96 Villas
Staffhouse
```

---

## 5. Fixture marker

Use exactly:

```text
Development location seed data: BWP
```

Do not reuse:

```text
Development location seed data
Development location seed data: 96 Villas
```

The BWP dataset must be independently identifiable.

---

## 6. Stable IDs and idempotent seed

Extend the existing development fixture workflow.

For an existing BWP fixture row whose code is owned by the BWP marker:

```text
update name to exact source string if needed
set is_active = TRUE
preserve existing primary-key ID
```

For a code collision with a row that is **not** owned by the BWP marker:

```text
return a clear error
rollback the transaction
do not overwrite the non-fixture row
do not leave a partial BWP seed
```

Repeated runs must produce the same final data.

Do not:

- truncate `locations`;
- reset sequences;
- delete locations;
- delete tickets;
- rewrite 96 Villas rows;
- mutate unrelated real/non-fixture locations.

---

## 7. No migration

The existing `locations` schema is sufficient.

Do not create migration `0021`.

Do not edit `backend/migrations/`.

This is reference-data work in the explicit development fixture workflow.

---

# PART B — SEARCHABLE LOCATION LOOKUP BACKEND

## 8. Existing endpoint

Continue using:

```http
GET /api/v1/locations
```

Do not create a second location-search endpoint.

The endpoint remains:

```text
authenticated
PostgreSQL-backed
active-only
real IDs / codes / names
```

---

## 9. Query parameters

Add optional search parameters:

```text
q
limit
```

Examples:

```http
GET /api/v1/locations?q=oasis&limit=10
GET /api/v1/locations?q=96%20v&limit=10
GET /api/v1/locations?q=server&limit=10
GET /api/v1/locations?q=bwp%20-%20lobby&limit=10
```

### q

Rules:

- optional;
- trim leading/trailing whitespace;
- case-insensitive;
- blank after trimming behaves as no query;
- cap accepted query length at a small explicit safe value such as 100 Unicode code points / reasonable equivalent;
- invalid oversized query returns a bounded client error, not an unbounded DB operation.

### limit

Rules:

```text
default = 10
minimum = 1
maximum = 30
```

Invalid limit must use the project's existing lookup validation/error style.

Do not silently allow arbitrary result sizes.

---

## 10. Search meaning

The user wants the list to shrink live toward the closest textual match.

This SPEC requires deterministic textual relevance, not semantic AI search and not typo-correction infrastructure.

Search must be case-insensitive and use the actual PostgreSQL `locations.name`.

Only active locations are eligible.

A query should match when its normalized text is contained in the location name.

---

## 11. Relevance ordering

For non-empty `q`, order results by the following ranking:

```text
0. exact name match
1. full-name prefix match
2. word/token prefix match
3. contains match
4. name ASC
5. id ASC
```

Equivalent explicit SQL is acceptable if it preserves this behavior.

No fuzzy-search dependency or PostgreSQL extension is required for this SPEC.

Do not add pg_trgm solely for this task.

Examples:

### q = `96 v`

Expected close results should surface entries such as:

```text
96 Villas
Server Room- 96 Villas
```

with the stronger prefix match first.

### q = `oasis`

Expected matching results include:

```text
BWP - Oasis bar
BWP - Oasis Bathroom
BWP - Oasis pool
BWP - Oasis pool bar
BWP - Oasis Swimming Pool
BWP- Oasis Kitchen
BWP - Toilet Oasis
```

The exact order among same-rank values must remain deterministic.

### q = `server`

Expected matching results include:

```text
Server Room
Server Room- 96 Villas
```

plus any other active server-named locations if present.

---

## 12. No-query behavior

`GET /api/v1/locations` without `q` remains supported for compatibility.

Do not break existing consumers/tests.

For the New Request searchable picker, frontend should prefer bounded queries instead of loading and rendering the entire location catalogue as one giant open list.

---

## 13. Query performance

Keep the implementation explicit and bounded.

Requirements:

- parameterized SQL;
- no `SELECT *`;
- no per-location queries;
- no N+1;
- active-only filtering in SQL;
- `LIMIT` in SQL;
- deterministic ordering in SQL;
- no loading all rows into Go and filtering there.

Do not add Redis/cache.

---

# PART C — SEARCHABLE LOCATION PICKER FRONTEND

## 14. SPEC-06.4 relationship

SPEC-06.4 established:

```text
visible text = location.name
submitted value = location.id
```

SPEC-06.5 keeps that rule.

However, SPEC-06.5 **supersedes only the native-select component constraint** from SPEC-06.4 because the user now explicitly requires search and a bounded result list.

The new control may be a focused accessible searchable combobox/listbox.

Do not regress to visible internal codes.

---

## 15. Required interaction

New Request → Location must work like this:

```text
user focuses/clicks Location
→ result panel opens
→ user types
→ frontend searches/filter results progressively
→ result list becomes smaller/more relevant
→ user clicks/taps desired location
→ input shows the selected location name
→ create request submits the selected real location_id
```

Do not require the user to scroll through hundreds of rows to find a location.

---

## 16. Search request behavior

Use the real backend search.

Recommended behavior:

```text
input changes
→ debounce ~150–250 ms
→ GET /api/v1/locations?q=<text>&limit=<bounded>
```

Requirements:

- do not request on every keystroke without any debounce;
- cancel or ignore stale responses;
- the newest query result must win;
- no automatic infinite retry;
- no polling;
- no runtime fallback list.

A small debounce is allowed and expected.

Do not introduce a large state-management library.

---

## 17. Empty-input behavior

When the Location field is focused but query text is empty:

- show only a bounded initial result set;
- do not render the entire catalogue;
- preserve a clean `No location` choice because location remains optional.

The initial result count should use the same bounded limit.

---

## 18. Selection semantics

The user sees:

```text
location.name
```

The application stores/submits:

```text
location.id
```

Do not submit:

```text
location.code
location.name
```

When a selected location is cleared:

```text
location_id = null
```

Preserve SPEC-06 optional-location behavior.

---

## 19. Search result visible content

Each result row contains only the readable location name.

Examples:

```text
1001
96 Villas
BWP - Oasis pool
Server Room- 96 Villas
IT Office
```

Forbidden:

```text
1001 (96BWV-ROOM-1001)
BWP - Oasis pool (BWP-AREA-101)
Server Room (BWP-AREA-153)
```

No ID/code/debug metadata.

---

## 20. Desktop dropdown behavior

Desktop result panel must:

- anchor directly under the Location field;
- remain within the New Request UI layer;
- use a bounded max-height;
- scroll internally when results exceed visible height;
- not expand the modal to the height of the entire location dataset;
- not push the rest of the form off-screen;
- preserve keyboard and mouse selection.

---

## 21. Mobile dropdown behavior

The current bad behavior shown by the user is:

```text
location dropdown opens extremely tall
→ extends far outside the modal/form
→ dominates the screen
```

That behavior is explicitly forbidden.

Mobile must follow the supplied SARA reference:

```text
search field
→ short result panel below it
→ only several results visible
→ internal vertical scroll for more
→ modal stays inside viewport
```

Requirements:

- result panel must not escape the modal viewport;
- result panel max-height should be viewport-aware;
- internal `overflow-y: auto`;
- modal/form itself must remain usable;
- user can scroll result rows without scrolling the entire page unexpectedly;
- closing/selecting the result must restore normal form interaction;
- no horizontal overflow at 360/390/430 widths.

Do not solve this by showing all results and shrinking font excessively.

---

## 22. Accessibility / interaction

Use an accessible combobox/listbox pattern appropriate to the existing Svelte codebase.

At minimum support:

```text
focus input
type query
click/tap result
Escape closes result panel
selected value remains clear/readable
```

Keyboard ArrowUp/ArrowDown + Enter support is strongly preferred and should be implemented if it can be done without introducing a large dependency.

Do not add a third-party dropdown library solely for this SPEC unless a concrete blocker is reported first.

---

## 23. Loading / empty / error states

The searchable picker must handle:

```text
loading
no matches
lookup error
selected value
No location
```

Suggested visible states:

```text
Searching…
No locations found.
Unable to load locations. Retry.
```

Do not expose raw backend error messages.

Preserve auth behavior:

```text
401 → login
```

---

# PART D — TICKET LIST CREATED / DUE / PRIORITY UI

## 24. Created On

Desktop and mobile Ticket List must render `created_at` as two visual lines:

```text
DD-MMM-YYYY
hh:mm AM/PM
```

Example:

```text
12-Sep-2026
04:43 PM
```

The date is above the time.

Use the existing timestamp value; do not change the backend timestamp contract.

---

## 25. Due Date

When `due_at` exists, render:

```text
DD-MMM-YYYY
hh:mm AM/PM
```

Example:

```text
12-Sep-2026
06:43 PM
```

Date above time.

When `due_at` is null:

```text
—
```

Do not reserve a second Priority row underneath.

---

## 26. Priority presentation

Remove the visible `Priority` badge from Ticket List.

When:

```text
priority = true
```

render the ticket **Title itself** as:

- bold/emphasized;
- red using the existing BWP SonaSea semantic danger/priority token if one exists;
- readable in Light and Dark theme.

When:

```text
priority = false
```

render Title normally.

Do not display visible text:

```text
Priority
```

in the ticket list.

Do not change the boolean field in backend/database.

This is a presentation change only.

---

## 27. Desktop ticket list

The SPEC-06.1 desktop column hierarchy remains:

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

Do not add Action.

Only change:

```text
Created On formatting
Due Date formatting
Priority visual treatment on Title
```

---

## 28. Mobile ticket list

Keep the compact mobile card direction from SPEC-06.1.

Apply the same priority rule:

```text
priority ticket → red/bold Title
normal ticket → normal Title
```

Keep Created time compact but if the mobile card currently displays the full Created timestamp, date must appear above time rather than as one long line.

Do not make the card tall again.

---

# PART E — TESTS

## 29. BWP seed tests

PostgreSQL-backed tests must prove:

- BWP fixture count = 155;
- all 155 code/name pairs match section 4 exactly;
- all BWP fixture rows active;
- repeated seed is idempotent;
- fixture-owned existing ID is preserved;
- non-fixture collision errors;
- collision rolls back the BWP seed transaction;
- 96 Villas fixture count remains 141;
- 96 Villas rows remain unchanged;
- generic/unrelated locations remain unchanged;
- similar-looking BWP names are not merged.

---

## 30. Location search backend tests

Test authenticated:

```http
GET /api/v1/locations
GET /api/v1/locations?q=...
```

Cover:

- unauthenticated → 401;
- no-query compatibility;
- query trim;
- case-insensitive search;
- exact match ranking;
- prefix ranking;
- contains ranking;
- stable tie ordering;
- active-only filtering;
- default limit;
- explicit limit;
- max limit cap / validation;
- empty result;
- query `96 v`;
- query `oasis`;
- query `server`;
- real positive IDs;
- no duplicate rows.

Do not use a separate test database environment variable.

Use the project's existing temporary-schema test contract.

---

## 31. Frontend searchable picker tests

Add focused tests for:

- location name-only display;
- internal code not rendered;
- query sent through lookup API;
- bounded `limit` sent;
- debounce/stale-response protection contract;
- selected result stores real ID;
- clearing selection produces null location;
- `No location` remains possible;
- no giant hardcoded location array;
- loading / no-results / error state;
- no auto retry;
- mobile result container has bounded scroll contract;
- Department selection remains unaffected.

Do not add a large frontend testing framework.

---

## 32. Ticket list presentation tests

Cover:

- Created On date and time rendered separately;
- Due Date date and time rendered separately;
- null Due Date → `—`;
- visible Priority badge removed;
- priority Title receives the dedicated priority class/style;
- normal Title does not receive priority styling;
- both desktop and mobile contracts.

Do not change ticket backend models just for presentation tests.

---

# PART F — MANUAL VERIFICATION

## 33. Database verification

Run development seed twice using the normal local configuration.

Verify:

```text
BWP fixture rows       = 155
96 Villas fixture rows = 141
```

No duplicate BWP fixture codes.

---

## 34. Search verification

Using real frontend + backend, verify:

```text
q = 96 v
q = oasis
q = server
q = bwp - lobby
```

Check that results become progressively smaller/relevant as the user types.

Selecting a result must create request state with the real `location_id`.

---

## 35. Responsive verification

Verify:

```text
360
390
430
768
1280
1920
```

Especially 360/390/430:

- location result list stays inside New Request modal;
- only a bounded number of rows are visible;
- result area scrolls internally;
- no horizontal overflow;
- modal actions remain reachable;
- search remains usable with mobile keyboard.

Verify Light and Dark themes.

---

## 36. Ticket-list visual verification

Verify real tickets:

### normal ticket

```text
Title = normal style
Created date above time
Due date above time or —
```

### priority ticket

```text
Title = bold red
no visible Priority badge
Created date above time
Due date above time or —
```

Verify desktop and mobile.

---

# PART G — SCOPE GUARD

## 37. Do not implement

Do not implement:

- Ticket Detail;
- Accept;
- Assign;
- Close;
- Chat;
- Checklist;
- attachments;
- Report;
- Settings;
- Admin location-management UI;
- semantic/AI search;
- pg_trgm migration;
- Redis/cache;
- WebSocket;
- polling;
- location hierarchy/group headings;
- SPEC-07;
- migration `0021`;
- runtime mock data.

Do not update `PROJECT_CONTEXT.md` in this SPEC unless explicitly requested after closure.

---

# PART H — DEFINITION OF DONE

## 38. SPEC-06.5 closes only when

- BWP fixture count is exactly 155;
- 7 newly supplied locations are included exactly;
- all source labels remain exact;
- 96 Villas remains 141 and unchanged;
- seed is transactional/idempotent/collision-safe;
- `GET /api/v1/locations` supports bounded real search;
- search relevance follows the locked ranking;
- New Request uses a searchable bounded picker;
- mobile result list does not overflow the modal;
- user sees names only, never internal codes;
- selected location still submits real `location_id`;
- Created On = date above time;
- Due Date = date above time or `—`;
- visible Priority badge is removed;
- priority Title is bold/red;
- desktop/mobile regression passes;
- no migration;
- no runtime mock data;
- no SPEC-07 work.

Suggested next phase:

```text
Update PROJECT_CONTEXT.md
→ then SPEC-07 — Ticket Detail Read
```
