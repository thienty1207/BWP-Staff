# SPEC-08.2 — Post-Rebrand Repository Hardening & Tooling Alignment

**Project:** Hotel Staff
**Repository:** `thienty1207/Hotel_Staff`
**Baseline:** `main` at `7ceb9dc493244350a2408bd572925176c95756f0`
**Status:** **✅ CLOSED**
**Predecessors:** SPEC-07 ✅ CLOSED, SPEC-08 ✅ CLOSED, SPEC-08.1 ✅ CLOSED
**Next product feature after closure:** SPEC-09 — Assign Ticket
**Schema migration required:** **No**
**Migration 0021:** **MUST NOT be created by this SPEC**

---

## Implementation and closure status — 2026-09-19

The bounded application hardening is implemented and locally verified. This includes runtime static
asset cleanup, the Hotel Staff favicon, a disallow-all robots policy, frontend environment-file
ignore alignment, the `hotel_staff` backup-target guard and regression test, a synthetic-PostgreSQL
GitHub Actions workflow, and foreign-Origin rejection for unsafe backend methods with the existing
request-ID error envelope.

`baron automation reconcile` succeeded, and supported plan/Harness/continuity operations identify
SPEC-08.2. The repository capability contract classifies `graphify-local` code-map generation as
optional. Native Baron/Graphify refresh and query were checked but cannot complete with the
installed CLI contract; the managed `docs/baron/platform/STACK_MAP.md` remains at its last
supported generated state and was not hand-edited. The temporary compatibility probe used for
diagnosis was removed, and no wrapper, fake receipt, or fake Stack Map output was committed. Under
the approved optional-tooling exception, this does not block closure because the canonical context,
application source, hosted CI, local verification, runtime smoke, and database-preservation
evidence are complete. The supported `baron plan complete` command was attempted but refused because
the current environment has no trusted execution receipt; no receipt was fabricated.

GitHub Actions CI run [35347035603](https://github.com/thienty1207/Hotel_Staff/actions/runs/35347035603)
for hardening commit `397b097f75a7c6a64d62076ca93668ba86ea6eaa` completed successfully for Frontend
and Backend. The follow-up CI run [35347412623](https://github.com/thienty1207/Hotel_Staff/actions/runs/35347412623)
for the verification documentation commit also completed successfully for Frontend and Backend.
SPEC-08.2 is therefore **✅ CLOSED** by the docs-only closure commit carrying this evidence. No
migration `0021`, schema change, seed rewrite, or persisted BWP master-data/ticket-data change was
made.

---

## 1. Purpose

SPEC-08.2 is a bounded post-rebrand hardening pass.

The Hotel Staff rebrand is already complete. This SPEC does **not** redo the rebrand and does **not** implement a new product feature.

The purpose is to remove technical residue left after the rename, align repository/tooling state with the actual application, harden the public/static surface, add deterministic repository verification, and make the repository safe to continue into SPEC-09 without stale orchestration state.

This SPEC exists because the runtime product is currently correct, while several repository/tooling surfaces still describe older project state or expose unnecessary legacy artifacts.

---

## 2. Current verified product identity

The following identity is canonical:

- Product: **Hotel Staff**
- Repository: `thienty1207/Hotel_Staff`
- Go module: `github.com/thienty1207/Hotel_Staff/backend`
- Documentation root: `Context-Spec-Hotel-Staff/`
- Local PostgreSQL database: `hotel_staff`
- Session cookie: `hotel_staff_session`
- Theme storage key: `hotel-staff-theme`
- Remembered username key: `hotel-staff-remembered-username`

SPEC-08.2 must preserve this identity.

---

## 3. Current known-good runtime state

The rebrand closure already verified:

- `/health`
- `/ready`
- login
- `/api/v1/auth/me`
- logout
- Open tickets
- Closed tickets
- Ticket Chat
- persisted accepted ticket data
- Light theme
- Dark theme
- mobile/narrow behavior
- frontend check/tests/build
- backend gofmt/list/vet/test/race/build

The current rebrand is **not** to be reopened.

Any regression introduced by SPEC-08.2 is a blocker.

---

## 4. Problems this SPEC must address

### 4.1 Baron active state is stale

The application context says the project is ready for SPEC-09, but several Baron "current" documents still describe historical work.

Known stale examples include:

- `docs/baron/continuity/CURRENT.md`
  - still refers to SPEC-01
  - still contains historical Rust/Cargo language
  - still contains the old `Context-Spec-BWP-SonaSea` path

- `docs/baron/plans/CURRENT.md`
  - still says SPEC-05 is `in_progress`

- `docs/baron/harness/CURRENT_INTENT.md`
  - still describes SPEC-05

- `docs/baron/platform/STACK_MAP.md`
  - reports no entrypoints
  - reports no build command
  - reports no test command

This is dangerous because `AGENTS.md` requires future agents to consume Baron continuity/plan/harness state.

A future agent must not receive contradictory project state such as:

- canonical context: SPEC-08.2 complete / SPEC-09 next
- Baron current state: SPEC-05 in progress

### 4.2 Legacy BWP assets remain in runtime public static storage

The active application no longer renders BWP branding, but the following files remain under the frontend runtime static directory:

- `frontend/static/images/bwp-logo.png`
- `frontend/static/images/login-background.png`

The active login currently uses:

- `/images/login-background.jpg`
- `/images/hotel-login-icon.svg`

The BWP logo already has a historical copy under:

- `img/Logo/BWP-logo.png`

Historical UI screenshots also remain under `img/UI-DEMO/`.

Runtime-public static storage must not keep unnecessary old-brand artifacts.

### 4.3 Favicon is still the Svelte starter logo

`frontend/src/lib/assets/favicon.svg` is the default Svelte logo and is currently used as the application favicon.

Hotel Staff should use an existing neutral Hotel Staff/hotel icon instead of framework starter branding.

No new corporate logo is required.

### 4.4 Search-engine crawling is currently allowed

`frontend/static/robots.txt` currently allows crawling.

Hotel Staff is a staff/internal operational application and should not invite search-engine indexing.

### 4.5 Frontend `.gitignore` contradicts the repository environment-file contract

Root `.gitignore` correctly ignores environment files.

However `frontend/.gitignore` currently includes exceptions that may allow:

- `.env.example`
- `.env.test`

The canonical project contract explicitly forbids introducing these alternate environment files.

### 4.6 Database backup script can produce a misleadingly named dump

`database/scripts/backup_database.ps1` reads the real database name from `backend/.env`, but always publishes:

- `database/hotel_staff.dump`

If `DATABASE_URL` is accidentally changed to another database, the script could dump that different database into a file still named `hotel_staff.dump`.

The backup command must fail closed if the configured database is not `hotel_staff`.

### 4.7 Repository verification is local-only

There is currently no `.github/workflows/` CI workflow.

Local agent verification is useful, but a push to `main` is not independently rechecked by GitHub.

The repository should gain a minimal deterministic CI workflow for the current Go + SvelteKit/Bun stack.

### 4.8 Browser mutation requests should reject foreign Origins

Current auth uses:

- HttpOnly cookie
- SameSite=Lax
- Secure outside development

That is a good baseline.

However authenticated mutation endpoints should additionally reject explicit foreign browser `Origin` values.

This SPEC adds a bounded same-origin defense without introducing a full CSRF-token system.

### 4.9 Current canonical context must describe post-hardening reality

After implementation, `PROJECT_CONTEXT.md` must be updated with the final SPEC-08.2 state and exact closure commit.

It must not claim stale "current main" values after closure.

---

## 5. Explicit non-goals

SPEC-08.2 must **not** implement:

- SPEC-09 Assign Ticket
- Close Ticket
- persistent chat messages
- checklist behavior
- attachments
- notifications
- Staff Meal
- Announcements
- Report
- Settings
- Admin UI
- Redis
- queues
- microservices
- Kubernetes
- new business schema
- data anonymization
- historical SPEC rewriting
- master-data migration
- migration 0021

The following are also explicitly deferred:

### 5.1 Login rate limiting

Login rate limiting is desirable before public production deployment.

It is **not** implemented in SPEC-08.2 because the final reverse-proxy/trusted-client-IP topology is not yet locked.

Adding an IP-based limiter before defining trusted proxy behavior risks rate-limiting all users under a proxy/NAT as one client.

Track it as a pre-production security requirement.

### 5.2 Production deployment architecture

The current Vite `/api` proxy is development-only.

A production deployment contract must later define:

- frontend hosting/runtime
- backend hosting/runtime
- reverse proxy / `/api` routing
- HTTPS termination
- trusted proxy/client IP behavior
- rate limiting
- production origin
- health/readiness use

Do not implement that architecture in SPEC-08.2.

### 5.3 BWP master-data anonymization

The following legacy persisted identifiers must remain unchanged:

- `BWP-AREA-*`
- `BWP-ROOM-*`
- `Development location seed data: BWP`
- `Development location seed data: BWP Rooms`
- `96 Villas` fixture data

If the repository later needs to become a fully generic CV/public-demo dataset, that must be a separate explicit anonymization contract.

---

## 6. Hard invariants

### 6.1 Database invariants

SPEC-08.2 must not:

- rewrite applied migrations
- create migration 0021
- alter schema
- reseed business data
- rewrite tickets
- rename persisted BWP codes
- mutate legacy BWP fixture ownership markers

Expected migration path remains:

- 0001–0017 active
- 0018 permanently retired
- 0019 active
- 0020 active
- 0021 unused and available for a future real schema change

### 6.2 Ticket invariants

Current ticket behavior must remain unchanged:

- status values exactly `pending`, `accepted`, `closed`
- Accept remains implemented
- Assign remains disabled/unimplemented
- Close remains disabled/unimplemented
- `tickets.department_id` remains original destination/family
- assignment remains separate
- Accept remains atomic with `ticket_activity`
- replay semantics remain unchanged

### 6.3 Auth invariants

Preserve:

- username/password only
- server-side sessions
- `hotel_staff_session`
- raw token never stored in DB
- SHA-256 session token hash
- Argon2id password verification
- HttpOnly
- SameSite=Lax
- Secure outside development
- inactive users cannot authenticate
- no password storage in frontend localStorage

### 6.4 Runtime data policy

No mock runtime data.

An empty database must produce real empty state.

Test fixtures remain deterministic and isolated.

---

## 7. Baron reconciliation requirements

### 7.1 Baron-owned files must not be repaired manually

`AGENTS.md` states that identity/metadata mismatch must use Baron-supported reconciliation.

Therefore:

- inspect Baron help/available commands
- use official Baron reconciliation/refresh commands only
- do not fabricate unsupported commands
- do not manually rewrite Baron-owned identity metadata merely to make text look current

At minimum investigate the supported equivalents of:

- `baron automation reconcile`
- project/platform refresh
- code map / stack detection refresh
- continuity/plan/harness lifecycle completion or reset

Use actual installed syntax.

### 7.2 Required Baron end-state

After SPEC-08.2, active/current Baron state must not tell future agents that:

- SPEC-01 is current
- SPEC-05 is still in progress
- Rust/Cargo is the active backend
- `Context-Spec-BWP-SonaSea` is the active context path
- build/test commands are unknown

The resulting current tooling state must agree with the repository reality:

- Go/Fiber backend
- SvelteKit/Bun frontend
- current documentation root
- previous feature/rebrand work closed
- no active unfinished SPEC-05 work
- next product feature is SPEC-09, but SPEC-09 is not yet implemented

### 7.3 Legacy Baron project slug exception

`.baron/project.toml` may still contain:

`project_slug = "bwp-sonasea"`

Do not edit this field manually.

If supported tooling cannot rename/rebind it, keep the already accepted exception:

- internal Baron legacy identity only
- not product/runtime branding
- not exposed by Hotel Staff UI/API

This exception does not require reopening SPEC-08.1.

### 7.4 If reconciliation cannot update stale current state

If supported Baron tooling cannot safely reconcile the active CURRENT/PLAN/HARNESS/STACK state:

- do not hand-edit managed state in violation of repository rules
- stop SPEC-08.2 closure
- report the exact blocker
- preserve all application code

SPEC-08.2 must remain OPEN until future agents can no longer be routed by materially stale "current" state.

For the final closure gate, this rule is satisfied when the active CURRENT/PLAN/HARNESS state is
reconciled to SPEC-08.2 and the only remaining limitation is the explicitly optional
`graphify-local` provider. In that case, document the compatibility exception and preserve the last
supported managed Stack Map; do not hand-edit it or fabricate a refresh receipt.

---

## 8. Runtime static asset cleanup

Remove from runtime public static storage:

- `frontend/static/images/bwp-logo.png`
- `frontend/static/images/login-background.png`

Before deletion, verify active source does not reference either file.

Preserve historical artifacts outside runtime static storage, including:

- `img/Logo/BWP-logo.png`
- `img/UI-DEMO/*BWP*.png`

Do not bulk-delete historical evidence.

Add/update tests so runtime product branding requires only current Hotel Staff assets.

---

## 9. Favicon cleanup

Stop using the Svelte starter favicon.

Preferred implementation:

- use existing `/images/hotel-login-icon.svg` as favicon
- remove `frontend/src/lib/assets/favicon.svg` if no longer referenced

Do not create a new corporate logo.

Acceptance:

- browser favicon source is Hotel Staff/hotel-neutral
- no `svelte-logo` remains as active product favicon

---

## 10. robots.txt hardening

Change the frontend robots policy to:

```text
User-agent: *
Disallow: /
```

Purpose:

- discourage indexing of an internal staff application
- avoid search-engine discovery of login/runtime static surfaces

This is not an authentication mechanism and must not be documented as one.

---

## 11. Frontend environment-file ignore cleanup

In `frontend/.gitignore`:

remove exceptions that allow:

- `.env.example`
- `.env.test`

The frontend must follow the same environment-file policy as the root repository.

Do not create replacement example/test env files.

---

## 12. Backup database safety guard

Update:

`database/scripts/backup_database.ps1`

After parsing `DATABASE_URL`, verify the selected database name is exactly:

`hotel_staff`

If not:

- stop before invoking `pg_dump`
- return a clear error that does not contain the database password
- do not write/replace `hotel_staff.dump`
- do not write/replace `full_app_schema.sql`

Example intent:

```text
Refusing backup: DATABASE_URL targets "<name>"; expected "hotel_staff".
```

Do not print credentials or the complete `DATABASE_URL`.

Keep existing temporary-file atomic publication behavior.

Keep `PGPASSWORD` restoration behavior.

---

## 13. GitHub CI foundation

Add a minimal GitHub Actions workflow under:

`.github/workflows/ci.yml`

or an equivalently clear semantic filename.

The workflow must run on at least:

- pushes to `main`
- pull requests targeting `main`

### 13.1 Frontend CI job

Use Bun.

Run:

```text
bun install --frozen-lockfile
bun run check
bun test
bun run build
```

### 13.2 Backend CI job

Use the Go version declared/required by the project.

Run:

```text
gofmt verification
go list ./...
go vet ./...
go test ./... -count=1
go test -race ./... -count=1
go build ./...
```

### 13.3 PostgreSQL-dependent tests

If backend tests require the project database configuration:

- use an ephemeral PostgreSQL service in CI
- create a temporary runtime `backend/.env` during the CI job
- do not commit that file
- do not introduce `.env.test`
- do not introduce `DATABASE_TEST_URL`
- do not introduce `TEST_*` DB URLs

The CI database must be isolated from any real/user database.

Use only synthetic CI credentials.

Do not run development seed against a real database.

### 13.4 CI claims

Do not claim CI passed until the pushed workflow actually runs successfully on GitHub.

If workflow execution is unavailable because of repository/platform restrictions, report that accurately.

Branch protection is useful but is not a code-level closure requirement for this SPEC.

---

## 14. Origin validation for authenticated mutations

Add a small backend middleware or equivalent bounded enforcement for unsafe mutation methods:

- POST
- PUT
- PATCH
- DELETE

Behavior:

1. if the request contains an `Origin` header:
   - it must match the configured `FRONTEND_ORIGIN`
   - otherwise reject with HTTP 403

2. if no `Origin` header is present:
   - non-browser/direct API tooling may continue to work

3. normal CORS preflight behavior must remain valid

4. GET/HEAD/read-only requests must not be blocked by this mutation guard

Use the existing canonical configured frontend origin.

Do not add wildcard trust.

Do not trust arbitrary `X-Forwarded-*` headers as an origin substitute.

Use the existing application error envelope.

Suggested stable error contract:

```json
{
  "error": {
    "code": "forbidden_origin",
    "message": "Request origin is not allowed",
    "request_id": "..."
  }
}
```

Exact internal implementation may differ, but it must be deterministic and tested.

### 14.1 Required tests

Add tests for at least:

- allowed configured Origin on mutation
- rejected foreign Origin on mutation
- mutation with no Origin remains usable for non-browser API tooling
- GET request is unaffected
- CORS preflight is unaffected
- request ID/error envelope is preserved

Do not implement a CSRF-token system in this SPEC.

---

## 15. Active old-brand audit

After implementation, search active/runtime repository areas for:

- `BWP-Staff`
- `BWP SonaSea`
- `BWP-SonaSea`
- `SonaSea`
- `bwp-sonasea`
- `bwp_session`
- `bwp-theme`
- `bwp-remembered-username`

Also search `BWP`.

Classify remaining matches.

Allowed remaining categories:

- historical closed SPEC text
- legacy persisted `BWP-AREA-*`
- legacy persisted `BWP-ROOM-*`
- legacy seed ownership markers
- historical screenshots/assets under archive/reference paths
- Baron legacy internal slug exception
- optional `graphify-local` code-map/managed Stack Map compatibility exception
- explicitly documented non-user-facing security test vector

Not allowed:

- active runtime branding
- active runtime public BWP static asset
- current README product identity
- current Go module/import identity
- current cookie/storage keys
- current operational backup name
- current orchestration state claiming old feature work is active

---

## 16. Documentation alignment

Update:

`Context-Spec-Hotel-Staff/PROJECT_CONTEXT.md`

after implementation.

It must record:

- SPEC-08.2 status
- final closure commit
- Baron current-state reconciliation outcome
- runtime static asset cleanup
- favicon
- robots policy
- frontend env ignore rule
- backup guard
- CI state
- mutation Origin defense
- deferred pre-production rate limiting
- deferred production deployment architecture
- legacy data preservation
- next feature: SPEC-09

Do not rewrite historical closed SPECs.

Update this SPEC itself only for implementation notes and closure evidence.

---

## 17. Required verification

### 17.1 Frontend

From `frontend/`:

```text
bun install --frozen-lockfile
bun run check
bun test
bun run build
```

Required:

- 0 Svelte check errors
- 0 Svelte check warnings
- all tests pass
- build passes

### 17.2 Backend

From `backend/`:

```text
gofmt -d .
go list ./...
go vet ./...
go test ./... -count=1
go test -race ./... -count=1
go build ./...
```

Required:

- no stale `BWP-Staff` module imports
- all tests pass
- race tests pass
- build passes

### 17.3 Runtime smoke

Perform a bounded runtime regression check:

- `/health`
- `/ready`
- login
- `/me`
- Open tickets
- Ticket Chat
- logout

No ticket lifecycle mutation is required for this SPEC.

Do not create/accept/assign/close tickets only to test repository cleanup.

### 17.4 Database preservation

Read-only compare before/after:

- departments
- active departments
- locations
- active locations
- `BWP-AREA-*`
- `BWP-ROOM-*`
- tickets
- migration ledger

Expected business data must remain unchanged.

No seed.

No migration 0021.

### 17.5 Static surface

Verify:

- `/images/bwp-logo.png` no longer exists in runtime static source
- unused `login-background.png` no longer exists in runtime static source
- current login background `.jpg` remains
- hotel icon remains
- favicon no longer points at Svelte starter asset
- `robots.txt` disallows crawling

### 17.6 Git scope

Before commit:

```text
git status
git diff
git diff --check
```

Before push:

```text
git diff --cached
git diff --cached --check
```

Preserve unrelated local work.

Never use:

- `git add .`
- `git reset --hard`
- `git clean -fd`
- `git restore .`

---

## 18. Closure criteria

SPEC-08.2 may be marked **✅ CLOSED** only when all applicable items pass:

- [x] canonical Hotel Staff identity unchanged
- [x] Baron active/current state no longer materially routes agents to stale SPEC-01/SPEC-05/Rust state
- [x] Baron legacy project slug handled only through supported tooling or documented accepted exception
- [x] runtime-public `bwp-logo.png` removed
- [x] unused runtime `login-background.png` removed
- [x] historical BWP artifacts outside runtime static storage preserved
- [x] Svelte starter favicon removed from active product use
- [x] robots policy disallows crawling
- [x] frontend `.gitignore` no longer permits `.env.example` / `.env.test`
- [x] backup script refuses non-`hotel_staff` database targets
- [x] no credentials printed/committed
- [x] GitHub CI workflow added
- [x] frontend local verification passes
- [x] backend local verification passes
- [x] mutation Origin validation implemented and tested
- [x] runtime smoke passes
- [x] DB business data unchanged
- [x] no migration 0021
- [x] SPEC-07 remains CLOSED
- [x] SPEC-08 remains CLOSED
- [x] SPEC-08.1 remains CLOSED
- [x] SPEC-09 remains unimplemented
- [x] PROJECT_CONTEXT updated to final post-08.2 reality

The optional `graphify-local` refresh/query incompatibility is recorded as an accepted tooling
exception. The managed Stack Map was not manually changed and no fresh Stack Map is claimed.

---

## 19. Expected implementation footprint

Likely touched files include:

- Baron-managed current state generated/updated through supported tooling
- `frontend/static/images/bwp-logo.png` — delete
- `frontend/static/images/login-background.png` — delete
- `frontend/src/routes/+layout.svelte`
- `frontend/src/lib/assets/favicon.svg` — likely delete
- `frontend/static/robots.txt`
- `frontend/.gitignore`
- frontend branding/static tests
- `database/scripts/backup_database.ps1`
- backend app middleware/error tests for Origin validation
- `.github/workflows/ci.yml`
- `Context-Spec-Hotel-Staff/PROJECT_CONTEXT.md`
- `Context-Spec-Hotel-Staff/Spec/SPEC-08.2-post-rebrand-repository-hardening.md`

This list is not permission to bulk-stage files.

Only actual scope-related changes may be committed.

---

## 20. Commit policy

Do not amend previous rebrand commits.

Use a new commit.

Suggested implementation commit:

```text
fix: harden post-rebrand repository state
```

If closure documentation is split after independent verification, a second docs-only commit is acceptable:

```text
docs: close SPEC-08.2 repository hardening
```

Push only to:

`thienty1207/Hotel_Staff`

Do not use the old BWP repository identity.

---

## 21. Final state after closure

After successful closure:

```text
SPEC-07    ✅ CLOSED
SPEC-08    ✅ CLOSED
SPEC-08.1  ✅ CLOSED
SPEC-08.2  ✅ CLOSED

Next:
SPEC-09 — Assign Ticket design
```

The repository should then be ready for the next product feature without active stale tooling state or unnecessary old-brand runtime artifacts.
