# SPEC-08.1 — Hotel Staff Rebrand Transition

> **Project:** Hotel Staff
>
> **Repository:** `thienty1207/Hotel_Staff`
>
> **Baseline:** `main` at `95cf693e905dd970a5388db27a7346659745b76a`
>
> **Database:** local PostgreSQL database has already been manually renamed to `hotel_staff`
>
> **Prerequisite:** SPEC-08 ✅ CLOSED
>
> **Status:** ✅ CLOSED
>
> This SPEC is a bounded project-identity migration. It must finish before SPEC-09 Assign Ticket begins.

---

# 1. Goal

Change the active product/repository identity from:

```text
BWP SonaSea
BWP-SonaSea
BWP Staff
BWP-Staff
bwp-sonasea
```

to the canonical Hotel Staff identity:

```text
UI/product: Hotel Staff
project slug: Hotel-Staff
GitHub repo: thienty1207/Hotel_Staff
Go module: github.com/thienty1207/Hotel_Staff/backend
local PostgreSQL DB: hotel_staff
```

The GitHub repository and local PostgreSQL database were already renamed manually before this SPEC.

This SPEC aligns the repository internals with those changes.

---

# 2. Non-goals

This SPEC must NOT implement:

```text
Assign Ticket
Close Ticket
chat message persistence
Checklist persistence
attachments
notifications/realtime
Report
Settings
Admin UI
```

Do not start SPEC-09.

Do not change ticket lifecycle behavior.

Do not add infrastructure.

---

# 3. Critical data safety boundary

The product brand is changing.

Persisted master-data identifiers are not automatically changing.

Do NOT rename or rewrite:

```text
BWP-AREA-*
BWP-ROOM-*
Development location seed data: BWP
Development location seed data: BWP Rooms
96 Villas fixture ownership
```

Do NOT update existing ticket/location foreign keys merely for branding.

Do NOT create migration `0021`.

Any future anonymization of those persisted identifiers requires a separate explicitly approved data-migration SPEC.

---

# 4. Git remote

The local clone may still have the pre-rename remote URL.

Inspect:

```bash
git remote -v
```

Canonical target:

```text
https://github.com/thienty1207/Hotel_Staff.git
```

If `origin` still references `thienty1207/BWP-Staff`, update only the remote URL:

```bash
git remote set-url origin https://github.com/thienty1207/Hotel_Staff.git
```

Then verify:

```bash
git remote -v
git fetch origin
git branch -vv
```

Do not rewrite commit history.

---

# 5. Canonical documentation folder

Current historical directory:

```text
Context-Spec-BWP-SonaSea/
```

Canonical target:

```text
Context-Spec-Hotel-Staff/
```

Use Git-aware move:

```bash
git mv Context-Spec-BWP-SonaSea Context-Spec-Hotel-Staff
```

Preserve all SPEC files and history.

Do not delete/recreate closed SPECs.

The new canonical context is the Hotel Staff `PROJECT_CONTEXT.md` supplied with this SPEC.

---

# 6. Historical closed SPECs

SPEC-01 through SPEC-08 remain historical contracts.

Do not bulk-rewrite every old occurrence of `BWP` / `SonaSea`.

Historical project-name text is allowed inside closed SPEC bodies.

Legacy dataset terminology must remain where it identifies actual persisted fixture contracts.

Future agents should use:

```text
PROJECT_CONTEXT.md
+
latest SPEC/amendment
```

for current identity.

SPEC-07 and SPEC-08 remain ✅ CLOSED.

---

# 7. README/current documentation

Update active/current documentation to use Hotel Staff.

At minimum audit:

```text
README.md
database/README.md
database/scripts/README.md
docs/baron/platform/PROJECT_PROFILE.md
docs/baron/architecture/CURRENT_ARCHITECTURE.md
docs/baron/harness/DOMAIN_LANGUAGE.md
```

Only change files containing stale product identity.

Do not rewrite unrelated Baron-managed content.

---

# 8. Go module / imports

Current stale module:

```text
github.com/thienty1207/BWP-Staff/backend
```

Canonical module:

```text
github.com/thienty1207/Hotel_Staff/backend
```

Update `backend/go.mod`.

Update all internal Go imports from the old repository prefix to the new one.

This is a source identity change only.

Do not alter package architecture or business behavior.

After update:

```bash
go list ./...
go vet ./...
go test ./... -count=1
go test -race ./... -count=1
go build ./...
```

must resolve the new module path cleanly.

---

# 9. Frontend product copy

Remove old product-facing BWP SonaSea branding from active UI.

Known baseline references include:

```text
frontend/src/routes/+page.svelte
- Tickets | BWP SonaSea
- BWP SonaSea staff tickets
- /images/bwp-logo.png
- BWP SonaSea alt text

frontend/src/routes/login/+page.svelte
- BWP SonaSea — Login
- BWP SonaSea sign-in
- BWP SONASEA STAFF
- Access your BWP SonaSea account.
```

Canonical UI copy:

```text
Hotel Staff
Hotel Staff Login
Hotel Staff staff tickets
Access your Hotel Staff account.
```

Use natural wording; do not mechanically produce awkward duplicated text.

Do not change ticket workflow/UI hierarchy.

---

# 10. Frontend brand assets

Audit:

```text
frontend/static/images/bwp-logo.png
img/Logo/BWP-logo.png
login background image
BWP-named UI demo/reference images
```

Rules:

1. Product-facing runtime UI must not show the old BWP/SonaSea brand after SPEC-08.1.
2. Historical screenshots/reference images under `img/UI-DEMO` may remain as historical reference assets if they are not shipped as active runtime branding.
3. Do not delete historical design/reference material merely to eliminate filenames.
4. Do not fabricate a new corporate logo.

Prefer an existing generic hotel icon/asset where suitable.

If `login-background.png` contains baked-in BWP/SonaSea text that cannot be safely corrected without a new image asset, report it explicitly as a blocker instead of claiming full visual rebrand completion.

---

# 11. Frontend localStorage keys

Current legacy keys include:

```text
bwp-theme
bwp-remembered-username
```

Canonical target:

```text
hotel-staff-theme
hotel-staff-remembered-username
```

Because this is development state, losing old local preference values is acceptable unless a compatibility migration is trivial.

Do not store passwords or session tokens in localStorage.

---

# 12. Session cookie

Current cookie:

```text
bwp_session
```

Canonical target for a clean development rebrand:

```text
hotel_staff_session
```

Renaming the cookie will invalidate existing browser sessions. This is acceptable for the current development environment.

Requirements:

```text
server login sets hotel_staff_session
auth middleware reads hotel_staff_session
logout clears hotel_staff_session
auth/integration tests use hotel_staff_session
no code path relies on bwp_session after migration
```

Preserve:

```text
HttpOnly
SameSite=Lax
Secure outside development
server-side session model
SHA-256 token hash storage
```

Do not change authentication architecture.

After the rename, manual login again is expected.

---

# 13. PostgreSQL database configuration

The user already renamed the local database in PostgreSQL:

```text
bwp-sonasea
→
hotel_staff
```

Update the local, uncommitted:

```text
backend/.env
```

so `DATABASE_URL` targets:

```text
.../hotel_staff
```

Do NOT expose or commit the password.

Do NOT add `.env.example`.

Do NOT create another database.

Do NOT copy data into a new database.

Verify existing data remains available in `hotel_staff`.

---

# 14. Database backup naming

Current active documentation/script naming still references:

```text
bwp-sonasea.dump
```

Canonical target:

```text
hotel_staff.dump
```

Audit/update:

```text
README.md
database/README.md
database/scripts/backup_database.ps1
```

If the existing dump file is tracked:

- use a Git-aware rename where appropriate;
- do not regenerate or overwrite a backup with current database content unless explicitly necessary and safe;
- do not treat backup-file naming as a schema migration.

---

# 15. No schema/master-data mutation

Rebranding should require:

```text
0 schema migrations
0 master-data rewrites
0 ticket rewrites
```

Expected migration state remains:

```text
0001..0017
0019
0020
0018 retired
no 0021
```

If any source change appears to require a DB migration purely for branding, stop and report before proceeding.

---

# 16. Preserve SPEC-07 / SPEC-08 behavior

Do not regress Ticket Chat.

Preserve:

```text
desktop sibling Chat panel
mobile full-width Chat
compact X close
conversation-only mobile scroll
selected-ticket stale-response protection
GET /api/v1/tickets/:id
```

Do not regress Accept.

Preserve:

```text
POST /api/v1/tickets/:id/accept
pending → accepted
server-session actor
atomic accepted activity
first accepter wins
same-user idempotency
different-user conflict
closed-ticket conflict
no assignment side effect
```

---

# 17. Preserve current action UI

Desktop and Chat:

```text
Accept = green
Assign = blue
Close = red
```

Current behavior:

```text
Accept = implemented
Assign = disabled/non-mutating
Close = disabled/non-mutating
```

Mobile cards still have no desktop Action column.

---

# 18. Repository-wide old-brand audit

After intended changes, perform a repository search for:

```text
BWP SonaSea
BWP-SonaSea
BWP Staff
BWP-Staff
SonaSea
bwp-sonasea
bwp_session
bwp-theme
bwp-remembered-username
```

Classify every remaining occurrence.

Allowed remaining occurrences may include only:

```text
historical closed SPEC wording
legacy persisted dataset identifiers/descriptions
historical screenshots/reference assets
historical notes explicitly marked as historical
```

Not allowed:

```text
active UI product copy
current README product identity
current Go module/import path
current auth cookie name
current localStorage keys
current backup script target
current canonical context/repo identity
```

Do not blindly replace bare `BWP` because that would corrupt legacy master data.

---

# 19. Tests

Update tests only where technical identity changed.

Expected examples:

```text
Go imports/module path
auth cookie tests
config/database URL test examples
frontend brand/copy tests if present
frontend storage-key tests if present
```

Do not rewrite fixture expectations for:

```text
BWP-AREA-*
BWP-ROOM-*
```

Those remain intentional.

---

# 20. Frontend verification

Run:

```bash
cd frontend

bun install --frozen-lockfile
bun run check
bun test
bun run build
```

Required:

```text
0 check errors
0 check warnings
all tests pass
build pass
```

---

# 21. Backend verification

Run from backend with the existing platform/toolchain setup:

```bash
gofmt -d .
go vet ./...
go test ./... -count=1
go test -race ./... -count=1
go build ./...
```

Also verify:

```bash
go list ./...
```

resolves packages under:

```text
github.com/thienty1207/Hotel_Staff/backend/...
```

not the old repository module path.

---

# 22. Database/manual verification

Using the updated local `.env`:

1. Start backend.
2. Confirm it connects to `hotel_staff`.
3. Confirm existing tickets still load.
4. Confirm Login works after cookie rename.
5. Confirm Open/Closed list loads.
6. Open Ticket Chat.
7. Confirm previously accepted ticket still reads as accepted.
8. Do not mutate a ticket merely for the rebrand test unless necessary.

Read-only SQL verification may confirm existing data is intact.

Do not print database credentials.

---

# 23. Visual verification

Check Light and Dark themes.

Check desktop and mobile/narrow viewport.

Required:

```text
no active BWP/SonaSea brand text
Hotel Staff product identity is visible where appropriate
no broken logo image
Login remains usable
Tickets shell remains usable
Ticket Chat unchanged functionally
action colors/layout unchanged
```

If login background contains baked-in old brand text, this must be reported, not hidden in the final report.

---

# 24. Git safety

Before editing:

```bash
git status
git log -5 --oneline
git remote -v
git diff --name-only
git diff
```

Preserve unrelated working-tree changes.

Never use:

```bash
git reset --hard
git clean -fd
git clean -fdx
git restore .
git checkout -- .
git add .
```

Use explicit staging.

---

# 25. Expected scope

Likely touched areas:

```text
Context-Spec-Hotel-Staff/
README.md
backend/go.mod
Go imports using old repo path
auth cookie implementation/tests
frontend current product copy
frontend localStorage keys
runtime branding asset references
database README / backup script
possibly tracked backup filename
```

`backend/.env` is local-only and must remain uncommitted.

Do not edit unrelated feature logic.

---

# 26. Commit discipline

Prefer one bounded rebrand commit after verification.

Suggested commit:

```text
refactor: rebrand project as Hotel Staff
```

Before commit:

```bash
git diff --check
git diff --cached
git diff --cached --check
git status
```

Push to:

```text
origin/main
```

where `origin` is:

```text
https://github.com/thienty1207/Hotel_Staff.git
```

---

# 27. Closure criteria

SPEC-08.1 may be marked ✅ CLOSED only when all are true:

```text
repo remote points to Hotel_Staff
canonical docs directory is Context-Spec-Hotel-Staff
PROJECT_CONTEXT uses Hotel Staff
current README uses Hotel Staff
Go module/import path uses Hotel_Staff
runtime UI no longer presents BWP SonaSea branding
frontend storage keys are rebranded
auth cookie is rebranded to hotel_staff_session
local backend connects to hotel_staff
backup script/docs use hotel_staff naming
frontend verification passes
backend verification passes
existing ticket data remains intact
no migration 0021
no BWP master-data codes were rewritten
old-brand occurrence audit is classified
```

If a baked-in image still visibly contains the old brand, SPEC-08.1 is not visually closed until that asset is replaced or the user explicitly accepts it.

---

# 28. Final report

Return:

```text
Git remote
Docs move
PROJECT_CONTEXT
README/current docs
Go module/import path
Frontend branding
Brand assets
localStorage keys
Session cookie
Database connection
Backup naming
Legacy BWP dataset preservation
Frontend verification
Backend verification
Manual verification
Old-brand occurrence audit
Files changed
Unrelated changes preserved
Commit SHA
Remaining blockers
```

Explicitly confirm:

```text
Product is now Hotel Staff.
Repository is thienty1207/Hotel_Staff.
Local DB target is hotel_staff.
No BWP-AREA-* / BWP-ROOM-* persisted identifiers were renamed.
No migration 0021 was created.
SPEC-07 remains CLOSED.
SPEC-08 remains CLOSED.
SPEC-09 was not implemented.
```

Do not mark SPEC-08.1 CLOSED if any required rebrand gate is still unverified.

End with either:

```text
SPEC-08.1 Hotel Staff rebrand is ready for independent closure review.
```

or:

```text
SPEC-08.1 Hotel Staff rebrand has remaining blockers.
```

---

# 29. Closure record — 2026-09-18

This final closure record supersedes the pre-closure response alternatives above. SPEC-08.1 is
closed after the rebrand gates and the additional user-reported mobile/narrow checks were verified.
The following closure evidence is recorded:

- GitHub repository identity verified as `thienty1207/Hotel_Staff`.
- Go module/import identity verified as `github.com/thienty1207/Hotel_Staff/backend`.
- Active runtime presents Hotel Staff branding; the login background is neutral.
- Authentication uses `hotel_staff_session`; theme and remembered-username storage use
  `hotel-staff-theme` and `hotel-staff-remembered-username`.
- Local backend connects to `hotel_staff`; login, `/me`, logout, and unauthenticated post-logout
  behavior passed.
- Open and Closed ticket lists and a persisted accepted-ticket Chat passed; created and accepted
  activity were visible.
- Light and Dark themes worked and persisted through reload.
- The user manually verified mobile/narrow behavior: no body-level horizontal overflow, full-card
  Chat activation, reachable close control, contained conversation scrolling, reachable
  composer/actions, and correct Accept / Assign / Close layout.
- Existing local database data remained intact: 16 departments (14 active), 866 locations
  (864 active), 155 `BWP-AREA-*` rows, 564 `BWP-ROOM-*` rows, and 4 tickets. Tickets 15 and 16
  retained their persisted acceptance fields.
- `BWP-AREA-*` and `BWP-ROOM-*` identifiers were preserved; no seed or business-data rewrite was
  run. Backup naming uses `hotel_staff.dump`.
- No migration `0021` was created. Frontend and backend automated verification passed locally;
  GitHub CI is not claimed.
- Baron retains internal `project_slug = "bwp-sonasea"`. It is not runtime/product branding and
  is not exposed by the application UI or API. Repository policy prohibits manual edits to
  Baron-owned identity metadata, and no supported rename/rebind path is available in the current
  trusted-receipt environment. No fake trusted execution/review receipt was created; this is an
  accepted internal tooling exception, not a product closure blocker.

SPEC-07 remains CLOSED. SPEC-08 remains CLOSED. SPEC-08.1 is CLOSED. SPEC-09 was not created or
implemented; its design is the next authorized planning step.
