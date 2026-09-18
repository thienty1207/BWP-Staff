# SPEC-06.3 — 96 Villas Location Master Data

> Project: **BWP SonaSea**
>
> Repository: `thienty1207/BWP-Staff`
>
> Baseline: `main` at or after `7204fd5e2c9b8381c2be33c132763a927bf740d7`
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

> **Later cleanup note:** SPEC-06.7 cleanup supersedes preservation of generic
> `ROOM-8020` and `ROOM-7309` as active rows. Those two legacy generic rows
> remain stored but inactive. This does not alter the 96 Villas fixture.

---

## 1. Goal

Add the approved **96 Villas** location master data to PostgreSQL so the existing **New Request → Location** dropdown can use the real database rows.

This is a bounded reference-data task.

Do not redesign the New Request form.

Do not hardcode locations in Svelte.

Keep the existing flow:

```text
PostgreSQL locations
→ GET /api/v1/locations
→ New Request Location dropdown
```

PostgreSQL remains the source of truth.

---

## 2. Scope of 96 Villas data

The approved 96 Villas dataset has:

```text
42 named area/location rows
99 room rows: 1001 through 1099 inclusive
-------------------------------------------
141 total 96 Villas rows
```

The room range is inclusive:

```text
1001
1002
...
1098
1099
```

Do not create `1000`.
Do not create `1100`.

---

## 3. Exact named locations

Preserve the display names exactly as supplied.

Do not silently normalize capitalization, punctuation, spacing, apostrophes, Vietnamese text, or wording.

| Code | Display Name |
|---|---|
| `96V` | 96 Villas |
| `96BWV-AREA-001` | 96-BWV - Asian kitchen |
| `96BWV-AREA-002` | 96-BWV - Auxiliary Swimming Pool |
| `96BWV-AREA-003` | 96-BWV - Bathroom |
| `96BWV-AREA-004` | 96-BWV - Buffet counter |
| `96BWV-AREA-005` | 96-BWV - Cold kitchen |
| `96BWV-AREA-006` | 96-BWV - Eng Fire Pump Room |
| `96BWV-AREA-007` | 96-BWV - Eng Mainpool MEP Room |
| `96BWV-AREA-008` | 96-BWV - Eng MEP Room |
| `96BWV-AREA-009` | 96-BWV - Eng Subpool MEP Room |
| `96BWV-AREA-010` | 96-BWV - Eng Water Treatment Room |
| `96BWV-AREA-011` | 96-BWV - Eng Well Water treatment Room |
| `96BWV-AREA-012` | 96-BWV - Eng workshop |
| `96BWV-AREA-013` | 96-BWV - European kitchen |
| `96BWV-AREA-014` | 96-BWV - Extra Pool |
| `96BWV-AREA-015` | 96-BWV - Female Locker |
| `96BWV-AREA-016` | 96-BWV - FO Reception |
| `96BWV-AREA-017` | 96-BWV - Generator Room |
| `96BWV-AREA-018` | 96-BWV - Gym |
| `96BWV-AREA-019` | 96-BWV - HK Store |
| `96BWV-AREA-020` | 96-BWV - Kid club |
| `96BWV-AREA-021` | 96-BWV - Kid's Club |
| `96BWV-AREA-022` | 96-BWV - Kid's Playground |
| `96BWV-AREA-023` | 96-BWV - Lobby |
| `96BWV-AREA-024` | 96-BWV - Lobby Lounge |
| `96BWV-AREA-025` | 96-BWV - Main kitchen |
| `96BWV-AREA-026` | 96-BWV - Main Pool |
| `96BWV-AREA-027` | 96-BWV - Male Locker |
| `96BWV-AREA-028` | 96-BWV - Outside |
| `96BWV-AREA-029` | 96-BWV - PA Store |
| `96BWV-AREA-030` | 96-BWV - Pastry kitchen |
| `96BWV-AREA-031` | 96-BWV - Spa |
| `96BWV-AREA-032` | 96-BWV - Toilet Gym |
| `96BWV-AREA-033` | 96-BWV - Toilet Hồ bơi phụ |
| `96BWV-AREA-034` | 96-BWV - Toilet Lobby |
| `96BWV-AREA-035` | 96-BWV - Toilet Spa |
| `96BWV-AREA-036` | 96-BWV - Toilet Tropicana |
| `96BWV-AREA-037` | 96-BWV - Tropicana Bar |
| `96BWV-AREA-038` | 96-BWV - Tropicana Kitchen |
| `96BWV-AREA-039` | 96-BWV - Tropicana Restaurant |
| `96BWV-AREA-040` | 96-BWV-Kitchen Office |
| `96BWV-AREA-041` | 96-BWV-Steward |

Important source rules:

- Screenshot overlap repeated some rows such as `96-BWV - Eng MEP Room`, `96-BWV - Eng Subpool MEP Room`, `96-BWV - Tropicana Bar`, and `96-BWV - Tropicana Kitchen`. These are continuation duplicates and must result in only one database row each.
- `96-BWV - Kid club` and `96-BWV - Kid's Club` are two distinct source labels and must both be retained.
- `96-BWV-Kitchen Office` and `96-BWV-Steward` intentionally preserve the source spelling without spaces around the hyphen after `96-BWV`.

---

## 4. Room locations

Create exactly 99 room-location rows.

Display names:

```text
1001
1002
...
1099
```

Stable code convention:

```text
96BWV-ROOM-1001
96BWV-ROOM-1002
...
96BWV-ROOM-1099
```

Do not prefix the visible room name with `Room` or `96-BWV` in this SPEC.

The user supplied the room locations separately as numbers, so the visible names remain the exact numeric strings.

---

## 5. 96 Villas fixture ownership marker

Use a dedicated development-fixture marker for this dataset:

```text
Development location seed data: 96 Villas
```

Do not reuse the generic marker:

```text
Development location seed data
```

for the new 96 Villas rows.

The dedicated marker lets the seed safely identify which rows it owns.

---

## 6. Existing generic development locations

The repository currently has generic development locations such as:

```text
Lobby
Ballroom
Back Office
Room 8020
Room 7309
Villa
```

SPEC-06.3 must **not** delete or deactivate these rows.

Reason:

The user has supplied only the complete **96 Villas** dataset, not the complete global hotel location master list.

Therefore this SPEC must only guarantee that the 96 Villas rows exist and are correct.

Unrelated locations may coexist in the dropdown until their own master-data SPECs are provided.

---

## 7. Stable codes and conflict safety

All 96 Villas rows must use non-null stable codes.

Named areas use the explicit codes defined in section 3.

Rooms use:

```text
96BWV-ROOM-<room number>
```

The seed must be idempotent.

For a code that already belongs to a row with the 96 Villas fixture marker:

```text
update name if required
set is_active = TRUE
preserve existing ID
```

For a code collision with a row that does **not** have the 96 Villas fixture marker:

```text
do not overwrite it
do not deactivate it
do not steal the code
return a clear seed error / blocker
```

Do not silently ignore a conflicting non-fixture row and then claim the 96 Villas dataset is complete.

---

## 8. Seed implementation

Extend the existing development fixture workflow.

Prefer keeping the existing generic development locations intact and adding a focused 96 Villas dataset/helper in the same `backend/admin` fixture area.

Do not create a new subsystem.

The seed must:

```text
begin transaction
→ insert/update all 42 named locations
→ generate and insert/update rooms 1001..1099
→ verify/handle conflicts safely
→ commit
```

Repeated runs must produce the same result.

Do not:

- truncate `locations`;
- reset sequences;
- delete ticket references;
- recreate rows unnecessarily;
- change IDs of compatible existing 96 Villas rows.

---

## 9. Generation rule for rooms

Do not manually hardcode 99 room entries if a simple explicit loop is clearer.

A bounded loop is acceptable:

```text
for room := 1001; room <= 1099; room++
```

Each generated row must have:

```text
name = decimal room number
code = 96BWV-ROOM-<room>
description = Development location seed data: 96 Villas
is_active = TRUE
```

No other room numbers.

---

## 10. No runtime mock data

Do not hardcode the 141 location options in:

```text
NewRequestDialog.svelte
frontend lookup API
frontend models
```

The frontend must continue to load:

```http
GET /api/v1/locations
```

from PostgreSQL.

No fallback array.

No fake runtime locations.

---

## 11. No migration

The existing `locations` table already supports:

```text
id
code
name
description
is_active
```

No schema change is required.

Do not create migration `0021`.

Do not put reference-data INSERT statements into `backend/migrations/`.

This belongs to the explicit development fixture workflow.

---

## 12. Location lookup compatibility

Do not redesign `GET /api/v1/locations`.

It must continue to:

- require authentication as currently implemented;
- return active PostgreSQL rows;
- expose real `id`, `code`, and `name`;
- preserve deterministic ordering by `name`, then `id`.

No new endpoint.

No client-side fake grouping.

No search/autocomplete redesign in this SPEC.

---

## 13. Create Ticket compatibility

`POST /api/v1/tickets` must continue to validate `location_id` as an active real PostgreSQL location.

Do not change:

```text
requester
department
title
description
priority
due_at
status
ticket_activity
transaction behavior
error mapping
```

A 96 Villas location selected in the dropdown must submit its real database ID.

---

## 14. Backend tests

Add focused PostgreSQL-backed tests proving:

- exactly 42 named 96 Villas rows exist;
- exactly 99 room rows exist;
- total 96 Villas fixture count is exactly 141;
- all 141 rows are active after seed;
- room names are exactly `1001` through `1099`;
- room codes are exactly `96BWV-ROOM-1001` through `96BWV-ROOM-1099`;
- no `1000`;
- no `1100`;
- no duplicate 96 Villas codes;
- no duplicate rows caused by repeated screenshot overlap;
- both `96-BWV - Kid club` and `96-BWV - Kid's Club` exist separately;
- Unicode name `96-BWV - Toilet Hồ bơi phụ` is preserved exactly;
- `96-BWV-Kitchen Office` is preserved exactly;
- `96-BWV-Steward` is preserved exactly;
- repeated seed is idempotent;
- existing compatible 96 Villas IDs are preserved;
- unrelated generic development locations are not changed;
- unrelated real/non-fixture locations are not changed;
- a non-fixture code collision is rejected safely rather than overwritten.

---

## 15. Lookup integration test

Use an authenticated integration test for:

```http
GET /api/v1/locations
```

After seeding, verify the response contains all 141 96 Villas rows with:

```text
real positive database IDs
correct codes
correct names
```

The endpoint may contain additional unrelated active locations.

Therefore:

**Do not assert that the whole endpoint has exactly 141 rows.**

Filter/assert the 96 Villas code set.

Preserve the existing deterministic lookup ordering.

---

## 16. Manual verification

Run the normal development seed against the local PostgreSQL database twice.

Verify the 96 Villas dataset in PostgreSQL:

```text
42 named areas
99 rooms
141 total
```

Then open **New Request → Location** using the real backend.

Verify representative values are present:

```text
96 Villas
96-BWV - Asian kitchen
96-BWV - Eng MEP Room
96-BWV - Kid club
96-BWV - Kid's Club
96-BWV - Lobby
96-BWV - Toilet Hồ bơi phụ
96-BWV - Tropicana Restaurant
96-BWV-Kitchen Office
96-BWV-Steward
1001
1050
1099
```

Verify there are no duplicate 96 Villas rows.

Do not create fake tickets just for manual verification if existing integration tests already prove `location_id` create-ticket behavior.

---

## 17. Frontend

No frontend implementation change should be necessary if the current Location dropdown already uses `GET /api/v1/locations`.

Do not modify frontend merely to inject the new values.

Run frontend regression verification only.

If a frontend code change appears necessary, stop and explain why before broadening scope.

---

## 18. Scope guard

Do not implement:

- Ticket Detail;
- Accept;
- Assign;
- Close;
- Chat;
- Checklist;
- attachment work;
- department changes;
- location hierarchy;
- location grouping;
- location search/autocomplete;
- Admin location-management UI;
- SPEC-07;
- migration `0021`;
- Redis/cache;
- frontend hardcoded locations;
- PROJECT_CONTEXT update.

Only implement 96 Villas location reference-data alignment and its tests.

---

## 19. Definition of Done

SPEC-06.3 closes only when:

- all 42 named 96 Villas locations exist exactly as specified;
- all rooms 1001–1099 exist;
- total 96 Villas dataset is 141 rows;
- stable codes are used;
- all 96 Villas rows are active;
- seed is safe and idempotent;
- non-fixture code collisions are not overwritten;
- unrelated locations are preserved;
- dropdown still uses the real PostgreSQL lookup endpoint;
- no frontend hardcoded locations exist;
- no migration is added;
- create-ticket behavior is unchanged;
- backend and frontend regression verification passes.

Suggested next phase remains:

```text
SPEC-07 — Ticket Detail Read
```
