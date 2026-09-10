# SPEC-01 Database Foundation Implementation Plan

> For agentic workers: REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox syntax for tracking.

Goal: Build the complete SQLx/PostgreSQL foundation described by SPEC-01 and provision the development administrator hothienty with an Argon2id password hash without exposing the local credential in the repository.

Architecture: Keep the backend as a small Rust modular monolith. Config reads the existing split DATABASE_* variables, database owns PgPool creation and migrations, and a narrowly scoped development seed module creates static reference data plus the admin user. SQLx migrations remain explicit and sequential under backend/migrations. No repositories, auth API, frontend, or generic abstraction layer is introduced.

Tech Stack: Rust 2024, Axum, Tokio, SQLx PostgreSQL migrations, dotenvy, url, argon2, thiserror, PostgreSQL pgcrypto extension, and the existing SvelteKit/Bun frontend left unchanged.

Spec: Context-Spec-BWP-SonaSea/Spec/SPEC-01-database-foundation.md; product auth and performance overrides are recorded in Context-Spec-BWP-SonaSea/PROJECT_CONTEXT.md.

Global Constraints

- PostgreSQL, Rust 2024, Axum, Tokio, and SQLx.
- BIGSERIAL/BIGINT identifiers; snake_case and plural table names.
- Ticket statuses exactly pending, accepted, and closed; assignment remains separate.
- Foreign keys default to ON DELETE RESTRICT; history-bearing data is not normally physically deleted.
- TIMESTAMPTZ, explicit SQL, bounded result sets, stable pagination, and no N+1.
- username is the only login identifier; email is optional contact data.
- Accounts are admin-created; no public signup, forgot-password, or email-reset flow.
- hothienty is seeded only in development; its local password is hashed with Argon2id.
- Never commit local credentials, plaintext passwords, or reusable public seed credentials.
- Extremely high database-performance bar: indexes must match WHERE/JOIN/ORDER BY paths, no redundant indexes, short transactions, bounded pool, explicit columns, and EXPLAIN (ANALYZE, BUFFERS) evidence.
- No frontend changes in SPEC-01; SvelteKit remains Bun-managed.

Execution: inline in the current task after the user approved the design.

---

### Task 1: Establish package modules and failing tests

Files:
- Create: backend/src/lib.rs
- Create: backend/src/config.rs with tests first
- Create: backend/src/seed.rs with tests first
- Modify: backend/Cargo.toml

Interfaces:
- DatabaseSettings::from_parts(host, port, name, user, password) returns encoded database settings.
- hash_password(password) returns a PHC Argon2id hash.

Steps:

- [ ] Write a config test for split variables, a hyphenated database name, and special-character credential encoding.
- [ ] Write a config test rejecting min_connections greater than max_connections.
- [ ] Write a password test that verifies the original password and asserts the Argon2id algorithm.
- [ ] Run cargo test config::tests seed::tests from backend and confirm the expected missing-implementation failure.
- [ ] Add only argon2, dotenvy, sqlx with runtime-tokio-rustls/postgres/migrate, thiserror, and url.
- [ ] Add lib.rs module declarations, then rerun the focused tests and keep the failure attributable to missing implementation.

---

### Task 2: Implement configuration and PgPool

Files:
- Modify: backend/src/config.rs
- Create: backend/src/shared/mod.rs
- Create: backend/src/shared/database.rs
- Modify: backend/src/lib.rs

Interfaces:
- Config::from_env() returns Config or a non-secret ConfigError.
- database::connect(&DatabaseSettings) returns PgPool.
- database::run_migrations(&PgPool) runs embedded migrations.

Steps:

- [ ] Implement DatabaseSettings with URL, max connections, min connections, and acquire timeout.
- [ ] Compose an encoded URL from DATABASE_HOST, DATABASE_PORT, DATABASE_NAME, DATABASE_USER, and DATABASE_PASSWORD when DATABASE_URL is absent.
- [ ] Support DATABASE_URL as an optional override.
- [ ] Reject invalid numeric values, zero max connections, and min greater than max without putting secrets in errors.
- [ ] Read APP_ENV, SEED_DEVELOPMENT_DATA, and development admin settings without logging them.
- [ ] Run config tests and confirm green.
- [ ] Implement PgPoolOptions with bounded min/max connections and acquire timeout.
- [ ] Embed migrations from ./migrations and run cargo fmt --check plus cargo check --all-targets --all-features.

---

### Task 3: Write the red database contract and create schema migrations

Files:
- Create: backend/tests/database_foundation.rs with assertions first
- Create: backend/migrations/0001_extensions.sql through 0016_audit_logs.sql

Interfaces:
- Migrations create every required table and enum used by later tests and seed code.

Steps:

- [ ] Write the real PostgreSQL integration test that loads local config, connects, runs migrations, and asserts all 15 required tables and all four enum types.
- [ ] Add checks for username, employee_code, nullable email, Argon2id password_hash storage, FK behavior, ticket lifecycle checks, and key column types.
- [ ] Run cargo test --test database_foundation -- --nocapture before migrations exist and confirm the expected red state.
- [ ] Add 0001_extensions.sql with only pgcrypto.
- [ ] Add 0002_enums.sql with exact user_role, ticket_status, notification_type, and audit_action values.
- [ ] Add tables in dependency order: departments, users, user_preferences, auth_sessions, locations, tickets, ticket_activity, ticket_messages, message_attachments, checklist_items, announcements, staff_meals, notifications, and audit_logs.
- [ ] Add username as a required unique login identity, keep employee_code required and unique, and keep email nullable and non-login.
- [ ] Add explicit foreign keys and RESTRICT defaults, with SET NULL only for nullable history actors.
- [ ] Add ticket status/acceptance/assignment/closure consistency checks, checklist completion checks, published announcement checks, and notification read checks.
- [ ] Rerun the database contract test and correct schema failures until green.

---

### Task 4: Add high-performance indexes, triggers, and static seed

Files:
- Create: backend/migrations/0017_indexes.sql
- Create: backend/migrations/0018_seed_development.sql
- Modify: backend/tests/database_foundation.rs

Interfaces:
- Every index has a documented access path.
- Static seed is safe to rerun.
- updated_at is maintained by one simple trigger function.

Steps:

- [ ] Extend the integration test with exact required index names, updated_at trigger behavior, and reference seed assertions.
- [ ] Run the focused database test and confirm new assertions fail before 0017/0018 exist.
- [ ] Create composite indexes aligned to ticket status/date, requester/date, assigned/status, department/status/date, location/date, and ticket-message/date access paths.
- [ ] Add indexes for activity, attachments, checklist ordering/completion, announcements, active meals, notifications, sessions, and audit history.
- [ ] Add partial indexes for unread notifications, published announcements, and non-revoked sessions.
- [ ] Add stable id tie-breakers only where they materially support deterministic pagination; do not create redundant indexes.
- [ ] Add set_updated_at() and triggers only to tables with updated_at.
- [ ] Seed departments IT, HK, FO, ENG, HR, FB and coded locations Lobby, Ballroom, Back Office, Room 8020, Room 7309, Villa using safe conflict handling.
- [ ] Keep admin credentials out of SQL migrations.
- [ ] Rerun the focused database test and confirm green.

---

### Task 5: Add Argon2id development admin provisioning

Files:
- Modify: backend/src/seed.rs
- Modify: backend/src/config.rs
- Modify: backend/src/main.rs
- Create: backend/src/bin/seed_development.rs
- Modify: backend/tests/database_foundation.rs

Interfaces:
- seed::hash_password(password) returns an Argon2id PHC string.
- seed::seed_development_admin(&PgPool, &DevelopmentAdmin) returns Created or AlreadyPresent.
- main runs dotenv loading, config, migrations, safe development seeding, then the health server.
- seed_development performs the same initialization once and exits.

Steps:

- [ ] Add a failing integration assertion for username hothienty, role admin, IT department, NULL email, Argon2id hash, and configured password verification.
- [ ] Call the seed twice and assert the second call does not replace the first hash.
- [ ] Run the focused test and confirm it is red because the seed implementation is missing.
- [ ] Implement explicit Argon2id version 0x13 hashing with a random salt; do not log password, salt, hash, or URL.
- [ ] Insert username hothienty, employee_code IT-ADMIN-001, full_name Ho Thien Ty, email NULL, role admin, and IT department.
- [ ] Use INSERT ... ON CONFLICT (username) DO NOTHING and never reset an existing password during startup.
- [ ] Insert default user_preferences only for the newly created admin.
- [ ] Gate automatic seeding on APP_ENV=development and SEED_DEVELOPMENT_DATA=true.
- [ ] Run cargo test --test database_foundation -- --nocapture and cargo run --bin seed_development; verify no credentials are printed.

---

### Task 6: Add environment example and documentation

Files:
- Modify: backend/.env locally only
- Modify: database/README.md
- Modify: README.md only if it contradicts backend/migrations

Steps:

- [ ] Add local-only seed settings to backend/.env while preserving the supplied database settings.
- [ ] Keep development settings only in the existing local backend/.env; never generate an example env file.
- [ ] Document the one-shot seed command and backend/migrations location.
- [ ] Document username-only login, admin-created accounts, and absence of signup/forgot-password.
- [ ] Verify with git status and a secret scan that the real password occurs only in ignored backend/.env.

---

### Task 7: Full verification and Baron proof

Files:
- Review all changed backend files and context/spec updates.
- Update Baron plan/proof/trace through Baron commands only.

Steps:

- [ ] Run migrations from zero against a fresh non-production PostgreSQL database or disposable schema.
- [ ] Run the migration path against the existing bwp-sonasea database without dropping/resetting it.
- [ ] Run representative read queries with EXPLAIN (ANALYZE, BUFFERS), record plans/timing, and verify indexes are used or explain any small-table planner choice.
- [ ] Run cargo fmt --check.
- [ ] Run cargo clippy --all-targets --all-features -- -D warnings.
- [ ] Run cargo test --all-targets --all-features -- --nocapture.
- [ ] Verify all schema, enum, FK, uniqueness, lifecycle, index, trigger, seed, and Argon2id assertions.
- [ ] Review final diff for password leakage, raw tokens, unnecessary indexes, frontend scope drift, and auth business logic scope drift.
- [ ] Record proof, mandatory gate evidence, autopilot review, and trace score; do not claim completion if any required gate fails.
