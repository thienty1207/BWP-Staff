# SPEC-06.4 — Location Dropdown Visual Fidelity

> Project: **BWP SonaSea**
>
> Repository: `thienty1207/BWP-Staff`
>
> Baseline: `main` at or after `0edd9a84d17d16ef1a2fef5e030c197f5eb035a0`
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

> **Supersession note — Location control:** SPEC-06.5 intentionally introduces
> the searchable Location picker and supersedes the native `<select>`
> restriction. The name-only visible label contract remains canonical.

---

## 1. Goal

Fix the **New Request → Location** dropdown so its visible options match the supplied SARA reference screenshots:

- simple;
- compact;
- one readable location name per row;
- no internal database/code metadata shown to the user.

This is a **visual fidelity correction**.

The screenshots supplied by the user are the source of truth for the visible option content.

Do not reinterpret the screenshots into a different UI pattern.

---

## 2. Current Defect

The current frontend renders location options as:

```svelte
<option value={String(location.id)}>
    {location.name}{location.code ? ` (${location.code})` : ''}
</option>
```

This produces cluttered user-facing values such as:

```text
1024 (96BWV-ROOM-1024)
1025 (96BWV-ROOM-1025)

96-BWV - Generator Room (96BWV-AREA-017)
96-BWV - Gym (96BWV-AREA-018)
96-BWV - HK Store (96BWV-AREA-019)
```

This does **not** match the reference.

---

## 3. Required Visible Result

The dropdown must display **only the location name**.

Correct examples:

```text
1012
1014
1015
1016
1017
1018
1019
1020
1021
1022
```

and:

```text
96-BWV - Main Pool
96-BWV - Male Locker
96-BWV - Outside
96-BWV - PA Store
96-BWV - Pastry kitchen
96-BWV - Spa
96-BWV - Toilet Gym
96-BWV - Toilet Hồ bơi phụ
96-BWV - Toilet Lobby
96-BWV - Toilet Spa
```

For the currently reported bad examples:

```text
WRONG:
1024 (96BWV-ROOM-1024)

RIGHT:
1024
```

```text
WRONG:
96-BWV - Generator Room (96BWV-AREA-017)

RIGHT:
96-BWV - Generator Room
```

```text
WRONG:
96-BWV - Kid club (96BWV-AREA-020)

RIGHT:
96-BWV - Kid club
```

---

## 4. Exact Rendering Contract

For every location option:

```text
visible text = location.name
option value = location.id
```

Nothing else.

Required Svelte shape:

```svelte
<option value={String(location.id)}>{location.name}</option>
```

Do not render any of the following in the visible option text:

```text
location.code
location.id
database IDs
parentheses
internal prefixes generated from code
metadata
debug labels
descriptions
category labels
```

Codes remain backend/internal data and may continue to exist in the API/model/tests.

They are simply **not user-facing in this dropdown**.

---

## 5. Reference Images Are Binding

For this SPEC, the supplied reference screenshots are not loose inspiration.

They define the visible behavior:

```text
one line
plain readable text
location name only
compact native-list feel
no code suffix
no second line
no badges
no extra metadata
```

Do not invent a richer select UI.

Do not add chips, cards, grouped sections, icons, search boxes, code labels, or custom metadata.

If a future implementation cannot match a supplied screenshot because of a concrete browser/platform limitation, report the limitation before intentionally deviating.

---

## 6. Dropdown Component

Keep the existing New Request Location control.

Do not replace the current `<select>` with a custom combobox in this SPEC.

Keep:

```text
label = Location
placeholder/default = No location
real PostgreSQL location IDs as option values
```

Only simplify visible option labels.

---

## 7. Ordering

Do not add frontend sorting.

Keep backend lookup order:

```sql
ORDER BY name ASC, id ASC
```

The frontend must render options in the exact order returned by the API.

No client-side regrouping.

No separate `Rooms` / `Areas` headings in this SPEC.

---

## 8. Data Contract

Do not change the real data flow:

```text
PostgreSQL
→ GET /api/v1/locations
→ LookupLocation[]
→ New Request <select>
```

Keep `LookupLocation.code` in the model/API if it is already part of the backend contract.

Do not remove backend code fields merely because they are no longer rendered.

---

## 9. Department Dropdown

Do not alter the Department dropdown.

Department already renders `department.name` only.

SPEC-06.4 is specifically a Location dropdown display correction.

---

## 10. Create Ticket Behavior

Selection must still submit:

```text
location_id = selected option's real database ID
```

Do not submit location name or location code.

Do not change create-ticket semantics.

---

## 11. Responsive Behavior

Verify on:

```text
360
390
430
768
1280
1920
```

Requirements:

- each location appears as one clean option row;
- no code suffix;
- no horizontal overflow introduced by frontend styling;
- mobile and desktop use the same visible naming rule.

Do not redesign the entire New Request modal.

---

## 12. Theme

Keep existing BWP SonaSea Light/Dark theme behavior.

The SARA screenshots define the **content simplicity and hierarchy** of the location options.

Do not force a global theme change.

Do not create a separate custom dropdown solely to force popup colors.

---

## 13. Tests

Add/update focused frontend tests that verify:

- location option renders `location.name`;
- location code is not concatenated into visible option text;
- room example appears as `1024`, not `1024 (96BWV-ROOM-1024)`;
- named area appears as `96-BWV - Generator Room`, not with `(96BWV-AREA-017)`;
- option value still uses the real `location.id`;
- Department dropdown behavior is unchanged;
- no hardcoded location list is introduced.

Use the current lightweight frontend test style.

Do not add a new UI test framework.

---

## 14. Manual Verification

Using real backend/PostgreSQL data, open:

```text
New Request
→ Location
```

Expected room rows:

```text
1012
1014
1015
1016
...
```

Expected named rows:

```text
96-BWV - Main Pool
96-BWV - Male Locker
96-BWV - Outside
96-BWV - PA Store
96-BWV - Pastry kitchen
...
```

Forbidden visible forms:

```text
1012 (96BWV-ROOM-1012)
1024 (96BWV-ROOM-1024)
96-BWV - Main Pool (96BWV-AREA-026)
96-BWV - Generator Room (96BWV-AREA-017)
```

Also verify selecting a row still submits the correct ID.

---

## 15. Scope Guard

Do not implement:

- backend location-data changes;
- new locations;
- location migrations;
- location grouping;
- location hierarchy;
- location search/autocomplete;
- custom combobox;
- Ticket Detail;
- Accept;
- Assign;
- Close;
- Chat;
- Checklist;
- SPEC-07;
- Redis/cache;
- PROJECT_CONTEXT update.

No migration `0021`.

No mock data.

---

## 16. Definition of Done

SPEC-06.4 is complete only when:

- Location dropdown visible labels contain names only;
- no `location.code` is visible;
- room options are plain numbers such as `1024`;
- named location options are plain names such as `96-BWV - Generator Room`;
- real database IDs remain option values;
- backend/API/schema are unchanged;
- no custom over-engineered dropdown is introduced;
- Department behavior is unchanged;
- responsive verification passes;
- frontend tests/check/build pass;
- no runtime mock data is added.

Suggested next phase remains:

```text
SPEC-07 — Ticket Detail Read
```
