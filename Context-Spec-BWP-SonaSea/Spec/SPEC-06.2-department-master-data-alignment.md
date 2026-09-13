# SPEC-06.2 — Department Master Data Alignment

> Project: **BWP SonaSea**
>
> Repository: `thienty1207/BWP-Staff`
>
> Baseline: `main` at or after `cb3959bf41cd84a3b81c51907445e01e8ec9e7aa`
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

---

## 1. Goal

Align the **New Request → Department** options with the approved hotel request-destination list.

The current dropdown is incomplete because the local development department fixtures contain only a small subset.

This SPEC is a bounded reference-data alignment task.

Do not redesign the New Request form and do not hardcode department options in Svelte.

The existing flow must remain:

```text
PostgreSQL departments
→ GET /api/v1/departments
→ New Request Department dropdown
```

PostgreSQL remains the source of truth.

---

## 2. Approved Department List

The active requestable department list must display these names exactly:

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

The dropdown ordering should remain alphabetical, using the existing backend lookup ordering.

Do not add a frontend-only custom order unless the existing backend no longer produces the order above.

---

## 3. Stable Codes

Use these stable codes for development reference rows:

| Code | Display Name |
|---|---|
| `CON` | Concierge |
| `DA` | Damaged Asset |
| `FB` | F&B |
| `FIN` | Finance Request |
| `FO` | Front Office |
| `HK` | Housekeeping |
| `HKPPM` | Housekeeping PPM |
| `IT` | IT |
| `KIT` | Kitchen |
| `LDRY` | Laundry |
| `LF` | Lost & Found |
| `MAINT` | Maintenance |
| `REC` | REC |
| `SEC` | Security |

Codes are internal stable identifiers. The dropdown displays `name`, not `code`.

Existing compatible codes such as `IT`, `HK`, `FO`, and `FB` must retain their IDs where possible.

---

## 4. Existing Development Fixtures

Current development fixtures include legacy/incomplete rows such as:

```text
IT Department
Housekeeping
Front Office
Engineering
Human Resources
Food & Beverage
```

Update the development fixture definition so that it represents the approved list above.

For compatible existing rows:

```text
FB → rename display name to F&B
IT → rename display name to IT
FO → keep Front Office
HK → keep Housekeeping
```

The seed must update existing development fixture rows by code instead of silently leaving stale display names behind.

Do not change primary keys merely to rename display names.

---

## 5. Legacy Development-Only Rows

`ENG / Engineering` and `HR / Human Resources` are not part of the approved New Request destination list.

Do not delete referenced rows.

For local development data only, make legacy development-only rows non-requestable by setting them inactive **only when they are clearly rows created by the development fixture mechanism**.

A safe implementation may identify those rows using the existing development-fixture description/known fixture codes.

Do not deactivate arbitrary real production data.

If the current seed architecture cannot safely distinguish development fixture rows from real data, stop and report the blocker instead of performing destructive updates.

---

## 6. Seed Idempotency

`cmd/seed_development` must remain safe to run repeatedly.

Required behavior:

```text
first run
→ inserts missing approved departments
→ updates compatible development fixture names
→ safely hides obsolete development-only request destinations

second/subsequent run
→ no duplicates
→ same final data
→ no destructive reset
```

Use a transaction.

Do not truncate `departments`.

Do not reset IDs/sequences.

Do not delete tickets/users.

Do not modify unrelated departments that are not known development fixtures.

---

## 7. No Runtime Mock Data

Do not hardcode the 14 department options in:

```text
NewRequestDialog.svelte
frontend ticket API
frontend ticket model
```

The frontend must continue to call the real endpoint:

```http
GET /api/v1/departments
```

The endpoint must continue to return active rows from PostgreSQL.

No fallback array.

No sample runtime data.

---

## 8. No Migration

This SPEC changes reference data, not schema.

Do not create migration `0021`.

Do not put fixture/data INSERT statements into `backend/migrations/`.

The existing project contract keeps schema migrations separate from development fixture data.

---

## 9. Backend Lookup Compatibility

Do not change lookup semantics unless necessary.

`GET /api/v1/departments` must continue to:

- require authentication as currently implemented;
- return active departments only;
- return real PostgreSQL rows;
- preserve its existing compact response;
- order departments deterministically by name/id.

No new endpoint.

No pagination required for this small master list.

---

## 10. Create Ticket Compatibility

`POST /api/v1/tickets` must continue to validate the chosen `department_id` against an active department.

Do not change:

```text
requester
status
priority
due_at
description
ticket_activity
transaction behavior
error mapping
```

This SPEC is only about making the approved destinations available.

---

## 11. Tests

Add focused backend tests for the development seed/reference data:

- all 14 approved departments exist after seeding;
- exact display names are correct;
- codes are unique;
- seed is idempotent;
- `FB` becomes `F&B`;
- `IT` becomes `IT`;
- legacy development `ENG` and `HR` do not appear as active request destinations;
- unrelated non-development department rows are not modified/deactivated;
- no duplicate rows are created.

Also verify lookup behavior:

```text
GET /api/v1/departments
```

returns the approved active list from PostgreSQL in deterministic order.

Do not add frontend tests merely to assert hardcoded names, because names must come from the API.

Existing frontend New Request tests must still pass.

---

## 12. Manual Verification

Run the real development seed against the local PostgreSQL database.

Then open New Request and verify the Department dropdown contains:

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

Verify:

- no duplicate entries;
- no `Engineering`;
- no `Human Resources`;
- no `Food & Beverage`;
- no `IT Department`;
- selecting an item still creates a ticket using its real database ID.

Do not create unnecessary sample tickets purely for this verification if existing integration tests already cover ticket creation.

---

## 13. Scope Guard

Do not implement:

- Ticket Detail;
- Accept;
- Assign;
- Close;
- Chat;
- Checklist;
- attachments;
- Staff Meal;
- Announcements;
- Report;
- Settings;
- Admin UI;
- new department-management UI;
- new schema;
- migration `0021`;
- Redis/cache;
- hardcoded frontend dropdown data.

Do not update `PROJECT_CONTEXT.md` in this SPEC unless explicitly requested later.

---

## 14. Definition of Done

SPEC-06.2 is complete only when:

- the real PostgreSQL-backed Department dropdown contains the approved 14 entries;
- the visible names match exactly;
- the seed remains idempotent and safe;
- legacy development-only `Engineering` and `Human Resources` no longer appear as request destinations;
- existing compatible department IDs are preserved where possible;
- no frontend hardcoded department list exists;
- no schema migration is added;
- no runtime mock data is added;
- existing login, ticket list, and create-ticket flows remain working;
- required Go/Bun verification passes.

Suggested next phase remains:

```text
SPEC-07 — Ticket Detail Read
```
