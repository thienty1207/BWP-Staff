# SPEC-06.7 — BWP Room Location Master Data

> Project: **BWP SonaSea**
>
> Repository: `thienty1207/BWP-Staff`
>
> Baseline: `main` at or after `363d4a7cabc7e8b65a316c18210b3c79db13c1c2`
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
> - SPEC-06.5 ✅ CLOSED
> - SPEC-06.6 ✅ CLOSED

> **SPEC-06.7 ✅ CLOSED — post-implementation cleanup:**
> - legacy generic `ROOM-8020` remains stored but inactive;
> - legacy generic `ROOM-7309` remains stored but inactive;
> - canonical `BWP-ROOM-8020` remains active;
> - canonical `BWP-ROOM-7309` remains active;
> - active lookup exposes canonical numeric room names only;
> - the integration test file is `bwp_room_locations_integration_test.go`.

---

## 1. Goal

Add the approved **BWP hotel room numbers** to the existing PostgreSQL `locations` master data.

This SPEC is a bounded reference-data task.

The user supplied:

- the 2xxx room set as explicit text;
- the 3xxx, 4xxx, and 5xxx room sets from reference screenshots.

Those source values are authoritative.

Do not infer additional room numbers from numerical gaps.

Keep the existing real flow:

```text
PostgreSQL locations
→ GET /api/v1/locations
→ searchable New Request Location picker
```

No frontend hardcoding.

---

## 2. Dataset size

The BWP room dataset in this SPEC contains exactly:

```text
2xxx rooms = 72
3xxx rooms = 75
4xxx rooms = 75
5xxx rooms = 65
6xxx rooms = 60
7xxx rooms = 75
8xxx rooms = 76
9xxx rooms = 66
------------------
total      = 564
```

These 564 rows are **additional** to the existing:

```text
BWP area/location fixture = 155
96 Villas fixture         = 141
```

Do not merge fixture ownership markers.

---

## 3. Exact room-number source set

Only the following room numbers are valid for SPEC-06.7.

**Any room number not listed here must NOT be generated.**

### 2xxx rooms — 72 rows

```text
2001, 2002, 2003, 2004, 2005, 2006, 2007, 2008, 2009, 2010, 2011, 2012, 2014, 2015, 2016, 2017, 2018, 2020, 2022, 2024, 2026, 2028, 2100, 2101, 2102, 2103, 2104, 2105, 2106, 2107, 2108, 2109, 2110, 2111, 2112, 2114, 2115, 2116, 2117, 2118, 2120, 2122, 2124, 2126, 2200, 2201, 2202, 2203, 2204, 2205, 2206, 2207, 2208, 2209, 2211, 2215, 2217, 2300, 2301, 2302, 2303, 2304, 2305, 2306, 2307, 2308, 2309, 2310, 2311, 2315, 2317, 2319
```

### 3xxx rooms — 75 rows

```text
3001, 3002, 3003, 3004, 3005, 3006, 3007, 3008, 3009, 3010, 3011, 3012, 3014, 3015, 3016, 3017, 3018, 3020, 3022, 3024, 3026, 3028, 3100, 3101, 3102, 3103, 3104, 3105, 3106, 3107, 3108, 3109, 3110, 3111, 3112, 3114, 3115, 3116, 3117, 3118, 3120, 3122, 3124, 3126, 3200, 3201, 3202, 3203, 3204, 3205, 3206, 3207, 3208, 3209, 3211, 3215, 3217, 3219, 3221, 3300, 3301, 3302, 3303, 3304, 3305, 3306, 3307, 3308, 3309, 3310, 3311, 3315, 3317, 3319, 3321
```

### 4xxx rooms — 75 rows

```text
4001, 4002, 4003, 4004, 4005, 4006, 4007, 4008, 4009, 4010, 4011, 4012, 4014, 4015, 4016, 4017, 4018, 4020, 4022, 4024, 4026, 4028, 4100, 4101, 4102, 4103, 4104, 4105, 4106, 4107, 4108, 4109, 4110, 4111, 4112, 4114, 4115, 4116, 4117, 4118, 4120, 4122, 4124, 4126, 4200, 4201, 4202, 4203, 4204, 4205, 4206, 4207, 4208, 4209, 4211, 4215, 4217, 4219, 4221, 4300, 4301, 4302, 4303, 4304, 4305, 4306, 4307, 4308, 4309, 4310, 4311, 4315, 4317, 4319, 4321
```

### 5xxx rooms — 65 rows

```text
5001, 5002, 5003, 5004, 5005, 5006, 5007, 5008, 5009, 5010, 5011, 5012, 5014, 5015, 5016, 5017, 5018, 5020, 5022, 5024, 5026, 5028, 5100, 5101, 5102, 5103, 5104, 5105, 5106, 5107, 5108, 5109, 5110, 5111, 5112, 5114, 5115, 5116, 5117, 5118, 5120, 5122, 5124, 5209, 5211, 5215, 5217, 5219, 5221, 5300, 5301, 5302, 5303, 5304, 5305, 5306, 5307, 5308, 5309, 5310, 5311, 5315, 5317, 5319, 5321
```

### 6xxx rooms — 60 rows

```text
6001, 6002, 6003, 6004, 6005, 6006, 6007, 6008, 6009, 6010, 6011, 6012, 6014, 6015, 6016, 6017, 6019, 6100, 6101, 6102, 6103, 6104, 6105, 6106, 6107, 6108, 6109, 6110, 6111, 6112, 6114, 6115, 6117, 6119, 6200, 6201, 6202, 6203, 6205, 6207, 6209, 6211, 6215, 6217, 6219, 6221, 6300, 6301, 6302, 6303, 6304, 6305, 6307, 6309, 6311, 6315, 6317, 6319, 6321, 6666
```

### 7xxx rooms — 75 rows

```text
7001, 7002, 7003, 7004, 7005, 7006, 7007, 7008, 7009, 7010, 7011, 7012, 7014, 7015, 7016, 7017, 7018, 7019, 7020, 7022, 7024, 7026, 7100, 7101, 7102, 7103, 7104, 7105, 7106, 7107, 7108, 7109, 7110, 7111, 7112, 7114, 7115, 7116, 7117, 7118, 7119, 7120, 7122, 7124, 7200, 7201, 7202, 7203, 7204, 7205, 7206, 7207, 7208, 7209, 7211, 7215, 7217, 7219, 7221, 7300, 7301, 7302, 7303, 7304, 7305, 7306, 7307, 7308, 7309, 7310, 7311, 7315, 7317, 7319, 7321
```

### 8xxx rooms — 76 rows

```text
8001, 8002, 8003, 8004, 8005, 8006, 8007, 8008, 8009, 8010, 8011, 8012, 8014, 8015, 8016, 8017, 8018, 8019, 8020, 8022, 8024, 8026, 8100, 8101, 8102, 8103, 8104, 8105, 8106, 8107, 8108, 8109, 8110, 8111, 8112, 8114, 8115, 8116, 8117, 8118, 8119, 8120, 8122, 8124, 8200, 8201, 8202, 8203, 8204, 8205, 8206, 8207, 8208, 8209, 8211, 8215, 8217, 8219, 8221, 8300, 8301, 8302, 8303, 8304, 8305, 8306, 8307, 8308, 8309, 8310, 8311, 8315, 8317, 8319, 8321, 8888
```

### 9xxx rooms — 66 rows

```text
9001, 9002, 9003, 9004, 9005, 9006, 9007, 9008, 9009, 9010, 9011, 9012, 9014, 9015, 9017, 9100, 9101, 9102, 9103, 9104, 9105, 9106, 9107, 9108, 9109, 9110, 9111, 9200, 9201, 9202, 9203, 9205, 9207, 9209, 9211, 9215, 9217, 9219, 9221, 9300, 9301, 9302, 9303, 9304, 9305, 9307, 9309, 9311, 9315, 9317, 9319, 9321, 9501, 9502, 9503, 9504, 9505, 9506, 9507, 9508, 9509, 9510, 9511, 9601, 9602, 9999
```

---

## 4. No gap inference

The room numbering is intentionally sparse.

Examples of numbers that are absent from the supplied source must remain absent.

Do not assume complete numeric ranges such as:

```text
2001..2319
3001..3321
4001..4321
5001..5321
6001..6666
7001..7321
8001..8888
9001..9999
```

That would create invalid rooms.

Implementation must use the exact approved set from section 3.

---

## 5. Visible room names

Each room is stored as a Location whose visible name is the plain room number.

Examples:

```text
2001
2126
2319
3001
3321
4001
4321
5001
5321
6001
6666
7001
7321
8001
8888
9001
9321
9999
```

Do not display:

```text
Room 2001
BWP - 2001
BWP Room 2001
```

unless the user explicitly requests that later.

The searchable Location picker must continue displaying only:

```text
location.name
```

as defined by SPEC-06.4 through SPEC-06.6.

---

## 6. Stable internal codes

Use:

```text
BWP-ROOM-<room number>
```

Examples:

```text
BWP-ROOM-2001
BWP-ROOM-2319
BWP-ROOM-3001
BWP-ROOM-4321
BWP-ROOM-5321
BWP-ROOM-6666
BWP-ROOM-7321
BWP-ROOM-8888
BWP-ROOM-9321
BWP-ROOM-9999
```

Codes are internal metadata.

They must not appear in the user-facing Location picker.

---

## 7. Dedicated fixture marker

Use exactly:

```text
Development location seed data: BWP Rooms
```

Do not reuse:

```text
Development location seed data
Development location seed data: BWP
Development location seed data: 96 Villas
```

The BWP room dataset must be independently identifiable.

---

## 8. Seed implementation

Extend the existing development fixture workflow.

Do not create a new subsystem.

Prefer a clear explicit room-number slice/set grouped by floor range over broad range generation, because the source contains many deliberate gaps.

The seed must run inside the existing development fixture transaction.

For every approved room:

```text
code        = BWP-ROOM-<number>
name        = <plain numeric string>
description = Development location seed data: BWP Rooms
is_active   = TRUE
```

---

## 9. Idempotency and ID preservation

Repeated seed runs must be safe.

For an existing row whose code is owned by the BWP Rooms marker:

```text
refresh exact name if necessary
set is_active = TRUE
preserve primary-key ID
```

Second and later seed runs must:

```text
create no duplicate rows
retain the same 564-room set
retain IDs for existing owned rows
```

---

## 10. Collision safety

If a `BWP-ROOM-*` code already exists on a row that is NOT owned by:

```text
Development location seed data: BWP Rooms
```

then:

```text
return a clear seed error
rollback the transaction
do not overwrite the existing row
do not leave partial room data
```

Do not steal or silently reuse non-fixture rows.

---

## 11. Preserve existing location datasets

SPEC-06.7 must not change:

```text
155 BWP area/location fixture rows
141 96 Villas fixture rows
generic development locations
unrelated real/non-fixture locations
```

Do not deactivate, rename, delete, or re-code them.

---

## 12. No migration

No schema change is required.

Do not:

- create migration `0021`;
- modify an existing migration;
- add reference-data INSERT statements to migrations.

Use the existing development fixture path.

---

## 13. Lookup/search compatibility

Do not redesign:

```http
GET /api/v1/locations
```

Keep all SPEC-06.6 behavior:

```text
authenticated
PostgreSQL-backed
active-only
no artificial limit when limit is omitted
literal search for % and _
exact → prefix → token-prefix → contains ranking
name ASC / id ASC deterministic tie-break
```

With blank Location search, the 265 BWP rooms must be available in the full scrollable catalogue.

Searching a room number must find the corresponding room.

Examples:

```text
q=2001
→ 2001

q=2319
→ 2319

q=3001
→ 3001

q=4321
→ 4321

q=5124
→ 5124
```

---

## 14. Create Ticket compatibility

Do not change create-ticket semantics.

When a BWP room is selected:

```text
user sees: 2301
frontend stores: real PostgreSQL location.id
POST sends: location_id
```

Do not submit the room number or internal code as a replacement for `location_id`.

---

## 15. Frontend

No frontend implementation change is expected.

The searchable picker from SPEC-06.5/06.6 should automatically display the new real PostgreSQL rows.

Do not hardcode the 265 rooms in Svelte/TypeScript.

If frontend source appears to require changes, stop and report why before broadening scope.

---

## 16. Backend tests

Add focused PostgreSQL-backed tests proving:

- BWP Rooms fixture count is exactly `564`;
- exact room-number set equality with section 3;
- exactly 72 2xxx rooms;
- exactly 75 3xxx rooms;
- exactly 75 4xxx rooms;
- exactly 65 5xxx rooms;
- exactly 60 6xxx rooms;
- exactly 75 7xxx rooms;
- exactly 76 8xxx rooms;
- exactly 66 9xxx rooms;
- all room fixture rows are active;
- code format is exactly `BWP-ROOM-<number>`;
- visible name is exactly the numeric room string;
- no duplicate codes;
- no duplicate fixture rows;
- repeated seed is idempotent;
- an existing owned room keeps its ID;
- non-fixture code collision errors;
- collision rolls back all partial changes;
- 155 BWP area rows remain unchanged;
- 141 96 Villas rows remain unchanged;
- unrelated locations remain unchanged.

The exact-set test must catch both:

```text
missing approved room
unexpected inferred room
```

---

## 17. Search integration tests

Using authenticated real PostgreSQL integration tests, verify:

```http
GET /api/v1/locations?q=2001
GET /api/v1/locations?q=2319
GET /api/v1/locations?q=3001
GET /api/v1/locations?q=4321
GET /api/v1/locations?q=5321
GET /api/v1/locations?q=6666
GET /api/v1/locations?q=7321
GET /api/v1/locations?q=8888
GET /api/v1/locations?q=9321
GET /api/v1/locations?q=9999
```

Each expected room must be present with:

```text
positive database ID
correct internal code
exact plain numeric name
```

Also verify representative absent numbers do not appear merely because they fall inside a surrounding range.

---

## 18. Manual verification

Run the normal development seed twice.

Verify in PostgreSQL:

```text
BWP Rooms = 564
BWP areas = 155
96 Villas = 141
```

Then open:

```text
New Request → Location
```

Verify representative searches:

```text
2001
2126
2319
3001
3321
4001
4321
5001
5321
6001
6666
7001
7321
8001
8888
9001
9321
9999
```

Visible result must be the plain room number.

Do not claim manual verification unless actually performed.

---

## 19. Performance note

This SPEC intentionally adds 564 rows to the existing Location catalogue.

Do not add Redis, pagination, Elasticsearch, virtualization, or another search service.

SPEC-06.6 already defines the intended full-list + bounded-scroll behavior for the current catalogue size.

Future load/performance tuning remains a separate later phase.

---

## 20. Scope guard

Do NOT implement:

- new BWP area locations;
- new 96 Villas locations;
- room-number normalization;
- guessed missing rooms;
- frontend hardcoded rooms;
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
- migrations;
- Redis/cache;
- SPEC-07;
- PROJECT_CONTEXT update;
- runtime mock data.

---

## 21. Definition of Done

SPEC-06.7 is complete only when:

- exactly 564 BWP room rows exist;
- exact room set matches section 3;
- no unlisted room is generated;
- room names are plain numbers;
- stable `BWP-ROOM-*` codes are used;
- dedicated BWP Rooms marker is used;
- seed is transactional/idempotent/collision-safe;
- existing IDs are preserved for owned rows;
- BWP areas remain 155 and unchanged;
- 96 Villas remain 141 and unchanged;
- searchable Location picker finds the new rooms using real PostgreSQL data;
- no frontend hardcoding;
- no schema/migration changes;
- no SPEC-07 work;
- required tests and verification pass.

After closure, update `PROJECT_CONTEXT.md` before moving to SPEC-07.
