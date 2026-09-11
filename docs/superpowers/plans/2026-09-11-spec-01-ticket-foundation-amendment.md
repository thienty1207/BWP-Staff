# SPEC-01 Ticket Foundation Amendment Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Bring the existing SPEC-01 PostgreSQL foundation into conformance with the clarified ticket request, attachment, and multi-assignment model through forward migration `0020`.

**Architecture:** Keep the generic Go migration runner unchanged and execute every discovered SQL migration in order. Add one forward-only PostgreSQL migration that expands the ticket schema, backfills legacy single-user assignments, creates normalized assignment/attachment tables and indexes, then removes the retired legacy columns. Keep development fixture data in the explicit seed command.

**Tech Stack:** Go, Fiber v3, PostgreSQL, pgx v5, pgxpool, explicit SQL, no ORM.

**Spec:** `Context-Spec-BWP-SonaSea/Spec/SPEC-01-database-foundation.md`, with repository rules in `Context-Spec-BWP-SonaSea/PROJECT_CONTEXT.md`.

## Global Constraints

- Do not modify or rewrite migrations `0001` through `0019`.
- Migration version `0018` remains retired and must not be reused.
- Use `0020_ticket_assignment_and_request_fields.sql` as the only new schema migration.
- Do not add login, authentication endpoints, ticket APIs, upload handlers, frontend features, or any future SPEC.
- Preserve existing ticket assignments during upgrade before dropping `tickets.assigned_to` and `tickets.assigned_at`.
- Keep current assignment truth in `ticket_assigned_departments` and `ticket_assigned_users`.
- Store only New Request attachment metadata in PostgreSQL; keep chat attachment metadata in `message_attachments`.
- Use real PostgreSQL integration tests with an isolated schema and never use production data.
- Do not commit `.env`, `.dump`, generated build output, or any secret.

---

### Task 1: Establish failing SPEC-01 contract tests

**Files:**
- Modify: `backend/shared/migrations_test.go`
- Modify: `backend/shared/foundation_test.go`

**Interfaces:**
- Consumes: the existing isolated-schema PostgreSQL test helper, `shared.RunMigrations`, and the existing development seed/security test helpers.
- Produces: failing tests that describe migration `0020`, the revised ticket schema, assignment cardinality, attachment metadata, and legacy-assignment backfill.

- [x] **Step 1: Extend the migration directory contract**

Add `0020_ticket_assignment_and_request_fields.sql` to the expected active schema migration list in `TestSpec01MigrationDirectoryContainsSchemaMigrationsOnly`, while retaining the assertion that `0018_seed_development.sql` is absent.

- [x] **Step 2: Extend the clean-schema foundation assertions**

In `TestSPEC01PostgreSQLFoundation`:

1. Expect versions `1..17`, `19`, and `20` after two migration runs.
2. Require `ticket_assigned_departments`, `ticket_assigned_users`, and `ticket_attachments`.
3. Require the two assignment membership indexes for each assignment table and `idx_ticket_attachments_ticket_created`.
4. Stop treating `idx_tickets_assigned_status` as required and assert it is absent after migration.
5. Insert a ticket with `priority = TRUE` and a concrete `due_at`, then assert the values round-trip; insert another ticket omitting both fields and assert `priority = FALSE` and `due_at IS NULL`.
6. Insert two department assignments and two user assignments for one ticket, assert both collections coexist, and assert duplicate pairs return PostgreSQL SQLSTATE `23505`.
7. Insert valid ticket-level attachment metadata and assert it is stored without requiring a chat message.
8. Insert a real `ticket_messages` row and a separate `message_attachments` row to prove chat attachments remain separate.

Use literal expected values and existing parameterized SQL helpers. Do not add mocks or production-only test APIs.

- [x] **Step 3: Add a real upgrade-path integration test**

Add `TestSPEC01MigrationBackfillsLegacyTicketAssignment` in `backend/shared/foundation_test.go`:

1. Create an isolated PostgreSQL schema with `openSPEC01Pool`.
2. Copy migrations `0001` through `0017` and `0019` into a temporary migration directory and run them with the real migration runner.
3. Insert a real department, requester, assigned user, and legacy ticket with a hand-picked UTC `assigned_at` timestamp using parameterized SQL.
4. Copy only `0020_ticket_assignment_and_request_fields.sql` into that temporary directory and run migrations again.
5. Assert the exact legacy user and timestamp exist in `ticket_assigned_users`.
6. Assert `tickets.assigned_to`, `tickets.assigned_at`, and `idx_tickets_assigned_status` are gone.
7. Assert the new priority/due-time defaults exist on the upgraded ticket.

The test must fail before migration `0020` exists because the active migration contract and upgrade path are incomplete.

- [x] **Step 4: Run the focused tests and capture the expected red result**

From `backend/`, run:

```powershell
go test ./shared -run 'TestSpec01MigrationDirectoryContainsSchemaMigrationsOnly|TestSPEC01PostgreSQLFoundation|TestSPEC01MigrationBackfillsLegacyTicketAssignment' -count=1 -v
```

Expected result before implementation: failure because `0020` is missing and the revised schema/upgrade contract is not yet present. If PostgreSQL test configuration is absent, record the skip and still verify the migration-directory test fails for the missing file.

---

### Task 2: Implement the forward-only ticket foundation migration

**Files:**
- Create: `backend/migrations/0020_ticket_assignment_and_request_fields.sql`

**Interfaces:**
- Consumes: the legacy schema created by `0008_tickets.sql` and `0017_indexes.sql`, including `tickets.assigned_to`, `tickets.assigned_at`, and `idx_tickets_assigned_status`.
- Produces: the revised ticket columns, normalized assignment tables, ticket-level attachment metadata table, required indexes, and a migrated schema with no legacy single-assignee columns.

- [x] **Step 1: Add request fields**

Add:

```sql
ALTER TABLE tickets
    ADD COLUMN priority BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN due_at TIMESTAMPTZ NULL;
```

- [x] **Step 2: Add normalized assignment and ticket-attachment tables**

Create `ticket_assigned_departments`, `ticket_assigned_users`, and `ticket_attachments` with the exact SPEC-01 columns, explicit `ON UPDATE RESTRICT ON DELETE RESTRICT` foreign keys, non-blank attachment string checks, non-negative size/dimension checks, and the required unique/non-unique indexes.

- [x] **Step 3: Validate and backfill legacy assignments**

Before dropping the legacy columns, fail the migration if a legacy ticket has only one side of the old `(assigned_to, assigned_at)` pair. Then insert every non-null legacy assignment into `ticket_assigned_users` preserving `ticket_id`, `assigned_to` as `user_id`, and the original `assigned_at`. Do not fabricate an actor or timestamp.

- [x] **Step 4: Contract the retired legacy model**

Drop `idx_tickets_assigned_status`, drop `tickets_assignment_pair_check`, and remove `tickets.assigned_to` and `tickets.assigned_at` after the backfill. Keep the whole migration inside the existing migration runner transaction so a failure rolls back the expansion, backfill, and contract together.

- [x] **Step 5: Run the focused tests green**

Run the focused command from Task 1. The expected result is that the revised schema, assignment behavior, attachment behavior, and legacy backfill pass against the isolated real PostgreSQL database.

---

### Task 3: Align checked-in database scripts and schema export

**Files:**
- Modify: `database/scripts/verify_foundation.sql`
- Modify: `database/README.md` only where the active migration list or `0020` behavior needs clarification
- Modify: `database/full_app_schema.sql` so the checked-in schema export reflects the revised ticket model after the local backup refresh

**Interfaces:**
- Consumes: the active migration directory and revised PostgreSQL schema.
- Produces: reviewable database documentation and verification scripts that no longer require the retired legacy ticket index/columns and that include all three new tables and indexes.

- [x] **Step 1: Update read-only foundation verification**

Add the three new tables and five new assignment/attachment indexes to `verify_foundation.sql`, remove `idx_tickets_assigned_status` from its required list, and include checks for the active ledger ending at version `0020` without requiring a contiguous version `0018`.

- [x] **Step 2: Apply the migration to the existing local development database safely**

Before the local upgrade, run the existing `database/scripts/backup_database.ps1` and confirm it produces the ignored `database/bwp-sonasea.dump`. Run the application/seed migration path against the existing local database, verify the legacy assignment count before and after, and confirm the old version-18 ledger/data is preserved rather than reset.

- [x] **Step 3: Refresh the schema export**

Run the backup script after the local migration so `database/full_app_schema.sql` is generated from the actual revised database. Review the diff to ensure it contains `priority`, `due_at`, the three new tables and indexes, and no `tickets.assigned_to`, `tickets.assigned_at`, or `idx_tickets_assigned_status`.

---

### Task 4: Full verification, audit, and delivery

**Files:**
- Modify: only files already listed above, if verification exposes a concrete SPEC-01 defect
- Do not create: any future SPEC or feature implementation

**Interfaces:**
- Consumes: the revised migration, tests, scripts, and current context/spec documents.
- Produces: reproducible command evidence and a pushed Git commit containing only the in-scope implementation plus the already-present context edits.

- [x] **Step 1: Run required Go checks**

From `backend/`, run exactly:

```powershell
gofmt -d .
go mod tidy
go vet ./...
go test ./... -count=1
go build ./...
```

If supported, also run `go test -race ./... -count=1` with the configured MSYS2 GCC environment.

- [x] **Step 2: Verify fresh and existing PostgreSQL behavior**

Against a dedicated fresh test database, run migrations twice and check the new tables/constraints, zero development fixtures after normal migrations, and ledger versions `1..17,19,20`. Run the explicit development seed twice and verify departments, locations, admin, and Argon2id password behavior. Against the existing local database, verify data preservation and no ledger reset.

- [x] **Step 3: Run repository hygiene checks**

Run `git diff --check` and inspect `git status --short`. Ensure no `.env`, `.dump`, target/bin output, temporary migration directory, or unrelated file is staged.

- [ ] **Step 4: Commit and push**

Stage the new migration, strengthened tests, aligned database scripts/export, and the already-present context/spec edits without modifying their content. Commit with:

```text
feat: amend SPEC-01 ticket foundation
```

Push `main` to `origin`, then verify the local commit equals `origin/main` and the worktree is clean.

---

## Self-review checklist

- `0020` is forward-only and `0001`–`0019` remain byte-for-byte unchanged.
- Version `0018` is not reused and remains absent from the active migration directory.
- Every discovered migration is executed by the existing generic runner; no filename special case is added.
- Legacy assignment data is copied before the old columns are removed.
- Assignment tables support multiple departments, multiple users, and both at once.
- `priority` is boolean with default `FALSE`; `due_at` is nullable.
- New Request attachments are ticket-level metadata; chat attachments remain message-level metadata.
- All required indexes have documented access patterns and deterministic `id` tie-breakers where list ordering requires them.
- Tests use real PostgreSQL and isolated schemas; no production data is used.
- No future application feature or future SPEC is implemented.
