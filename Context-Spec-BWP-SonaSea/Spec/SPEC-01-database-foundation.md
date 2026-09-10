# SPEC-01 — Database Foundation

## 1. Goal

Create the complete initial PostgreSQL database foundation for **BWP SonaSea**.

This specification covers the database schema required for the planned full application, including:

- Authentication and sessions
- Users and departments
- Locations
- Tickets
- Ticket lifecycle
- Ticket assignment
- Ticket activity history
- Ticket chat
- Chat attachments
- Ticket checklist
- Staff Meal
- Announcements
- Notifications
- User preferences
- Reporting support
- Audit logging

The database must be production-oriented from the beginning, but the implementation must remain simple, explicit, maintainable, and easy to evolve through sequential PostgreSQL SQL migrations.

This spec does **not** implement business APIs or frontend UI.

---

## 2. Technology

- Database: **PostgreSQL**
- Backend language: **Go (current supported version)**
- Backend framework: **Fiber v3**
- Request/database context: **standard `context.Context`**
- Database library: **pgx v5 and pgxpool**
- Migration system: **explicit Go runner over sequential SQL files**

---

## 3. Engineering Principles

### 3.1 Boring and explicit code

Prefer simple, explicit database code.

Do not introduce abstractions without a concrete requirement.

Avoid:

- Generic repositories
- Generic database traits
- Repository factories
- Unnecessary trait objects
- `Arc`
- `Mutex`
- `RwLock`
- Complex generics
- Macro-heavy custom database abstractions
- ORM-style magic

pgx and pgxpool should be used directly and transparently. SQL should remain
visible in the database package.

---

### 3.2 Database-first foundation

The initial schema must support the full planned BWP SonaSea website from the beginning.

Future migrations are allowed and expected, but the first schema should already cover all currently known core features.

---

### 3.3 Data integrity first

Use:

- Primary keys
- Foreign keys
- Unique constraints
- Check constraints where appropriate
- Proper timestamp types
- Explicit indexes
- Safe deletion rules

Do not rely only on application code to enforce critical relationships.

---

## 4. Naming Conventions

Use:

- `snake_case` for table names
- `snake_case` for column names
- plural table names
- singular Go type names

Examples:

```text
users
departments
tickets
ticket_messages
checklist_items
```

Primary keys:

```text
id
```

Foreign keys:

```text
user_id
department_id
ticket_id
```

Timestamp columns:

```text
created_at
updated_at
accepted_at
assigned_at
closed_at
published_at
```

---

## 5. Identifier Strategy

Use PostgreSQL `BIGSERIAL` / `BIGINT` identifiers for core relational tables.

Reason:

- Simple
- Fast
- Easy to debug
- Easy to inspect manually
- Suitable for this application scale

Public identifiers may be added later if external exposure requires them.

---

# 6. Database Extensions

Enable only what is actually required.

Initial migration may enable:

```sql
CREATE EXTENSION IF NOT EXISTS pgcrypto;
```

Do not add unnecessary PostgreSQL extensions.

---

# 7. Enum Strategy

Prefer PostgreSQL enums only for values that are stable and tightly controlled.

Required initial enums:

```text
user_role
ticket_status
notification_type
audit_action
```

Recommended values:

## 7.1 `user_role`

```text
admin
staff
```

Do not create a large RBAC system in SPEC-01.

---

## 7.2 `ticket_status`

Exactly:

```text
pending
accepted
closed
```

Important:

**Assigned is not a ticket status.**

A ticket can be:

```text
status = accepted
assigned_to = <user>
```

Assignment is a separate property.

---

## 7.3 `notification_type`

Initial values:

```text
ticket_created
ticket_accepted
ticket_assigned
ticket_closed
new_message
announcement
system
```

---

## 7.4 `audit_action`

Initial values:

```text
create
update
delete
login
logout
accept
assign
close
publish
unpublish
```

---

# 8. Core Tables

The initial schema must contain:

```text
departments
users
user_preferences
auth_sessions
locations

tickets
ticket_activity
ticket_messages
message_attachments
checklist_items

announcements
staff_meals

notifications
audit_logs
```

---

# 9. Table Specifications

## 9.1 `departments`

Purpose:

Store hotel departments such as IT, Housekeeping, Front Office, Engineering, etc.

Columns:

```text
id              BIGSERIAL PRIMARY KEY
code            VARCHAR(32) NOT NULL
name            VARCHAR(100) NOT NULL
description     TEXT NULL
is_active       BOOLEAN NOT NULL DEFAULT TRUE
created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
```

Constraints:

```text
UNIQUE(code)
UNIQUE(name)
```

Examples:

```text
IT
HK
FO
ENG
FB
HR
```

Do not hard-code departments in application source code.

---

## 9.2 `users`

Purpose:

Store all BWP SonaSea users.

Authentication uses `username` as the login identifier. Email is optional
contact/profile data only; it is not a login identifier. Users are created by
an administrator. Public signup and forgot-password/email-reset flows are
outside the product direction.

Columns:

```text
id                  BIGSERIAL PRIMARY KEY
username            VARCHAR(50) NOT NULL
employee_code       VARCHAR(50) NOT NULL
email               VARCHAR(255) NULL
password_hash       TEXT NOT NULL
full_name           VARCHAR(150) NOT NULL
department_id       BIGINT NOT NULL
role                user_role NOT NULL DEFAULT 'staff'
avatar_url          TEXT NULL
is_active           BOOLEAN NOT NULL DEFAULT TRUE
last_login_at       TIMESTAMPTZ NULL
created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
```

Foreign key:

```text
department_id -> departments.id
```

Constraints:

```text
UNIQUE(username)
UNIQUE(employee_code)
UNIQUE(email) when email is not null
```

Rules:

- Never store plaintext passwords.
- `password_hash` contains an Argon2id password hash.
- `username` is the only login identifier; `email` is never used for login.
- Disabled users remain in the database.
- Do not physically delete normal users if historical references exist.

Recommended indexes:

```text
username
employee_code
email
department_id
is_active
```

---

## 9.3 `user_preferences`

Purpose:

Store per-user UI/application preferences.

Columns:

```text
id                  BIGSERIAL PRIMARY KEY
user_id             BIGINT NOT NULL
theme               VARCHAR(20) NOT NULL DEFAULT 'dark'
language            VARCHAR(10) NOT NULL DEFAULT 'en'
notifications_enabled BOOLEAN NOT NULL DEFAULT TRUE
created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
```

Foreign key:

```text
user_id -> users.id
```

Constraint:

```text
UNIQUE(user_id)
```

Theme initial accepted values:

```text
dark
light
system
```

Use a check constraint instead of creating another enum unless needed later.

---

## 9.4 `auth_sessions`

Purpose:

Store server-side authenticated sessions.

Columns:

```text
id                  BIGSERIAL PRIMARY KEY
session_token_hash  TEXT NOT NULL
user_id             BIGINT NOT NULL
expires_at          TIMESTAMPTZ NOT NULL
last_used_at        TIMESTAMPTZ NULL
user_agent          TEXT NULL
ip_address          INET NULL
created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
revoked_at          TIMESTAMPTZ NULL
```

Foreign key:

```text
user_id -> users.id
```

Constraints:

```text
UNIQUE(session_token_hash)
```

Rules:

- Never store the raw session token.
- Store only a secure hash of the session token.
- Logout should revoke the session using `revoked_at`.
- Expired or revoked sessions are invalid.

Indexes:

```text
session_token_hash
user_id
expires_at
```

Recommended partial index:

```text
revoked_at IS NULL
```

---

## 9.5 `locations`

Purpose:

Represent physical hotel locations.

Examples:

```text
Room 8020
Room 7309
Lobby
Restaurant
Ballroom
Office
Villa
Back Office
```

Columns:

```text
id              BIGSERIAL PRIMARY KEY
code            VARCHAR(50) NULL
name            VARCHAR(150) NOT NULL
description     TEXT NULL
is_active       BOOLEAN NOT NULL DEFAULT TRUE
created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
```

Constraints:

```text
UNIQUE(code) when code is not null
```

Index:

```text
name
```

---

# 10. Ticket Domain

## 10.1 `tickets`

Purpose:

Store support requests.

Columns:

```text
id                  BIGSERIAL PRIMARY KEY
requester_id        BIGINT NOT NULL
department_id       BIGINT NOT NULL
location_id         BIGINT NULL

title               VARCHAR(255) NOT NULL
description         TEXT NULL

status              ticket_status NOT NULL DEFAULT 'pending'

accepted_by         BIGINT NULL
accepted_at         TIMESTAMPTZ NULL

assigned_to         BIGINT NULL
assigned_at         TIMESTAMPTZ NULL

closed_by           BIGINT NULL
closed_at           TIMESTAMPTZ NULL

created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
```

Foreign keys:

```text
requester_id  -> users.id
department_id -> departments.id
location_id   -> locations.id
accepted_by   -> users.id
assigned_to   -> users.id
closed_by     -> users.id
```

Ticket rules:

### Pending ticket

```text
status = pending
accepted_by = NULL
accepted_at = NULL
```

### Accepted ticket

```text
status = accepted
accepted_by IS NOT NULL
accepted_at IS NOT NULL
```

### Assigned ticket

Assignment does not change ticket status automatically.

Example:

```text
status = accepted
assigned_to = 42
assigned_at = timestamp
```

### Closed ticket

```text
status = closed
closed_by IS NOT NULL
closed_at IS NOT NULL
```

Do not physically delete tickets.

---

## 10.2 Required ticket indexes

The following query patterns must be optimized:

Open tickets ordered newest first:

```text
status + created_at
```

Requester ticket history:

```text
requester_id + created_at
```

Assigned work:

```text
assigned_to + status
```

Department queue:

```text
department_id + status + created_at
```

Location history:

```text
location_id + created_at
```

Recommended SQL indexes:

```sql
CREATE INDEX idx_tickets_status_created_at
ON tickets (status, created_at DESC);

CREATE INDEX idx_tickets_requester_created_at
ON tickets (requester_id, created_at DESC);

CREATE INDEX idx_tickets_assigned_status
ON tickets (assigned_to, status);

CREATE INDEX idx_tickets_department_status_created
ON tickets (department_id, status, created_at DESC);

CREATE INDEX idx_tickets_location_created_at
ON tickets (location_id, created_at DESC);
```

---

# 11. Ticket Activity History

## 11.1 `ticket_activity`

Purpose:

Provide an immutable ticket event timeline.

Examples:

```text
ticket created
ticket accepted
ticket assigned
ticket reassigned
ticket closed
ticket reopened in future
```

Columns:

```text
id              BIGSERIAL PRIMARY KEY
ticket_id       BIGINT NOT NULL
actor_user_id   BIGINT NULL
action          VARCHAR(50) NOT NULL
metadata        JSONB NULL
created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
```

Foreign keys:

```text
ticket_id     -> tickets.id
actor_user_id -> users.id
```

Indexes:

```text
ticket_id + created_at
actor_user_id + created_at
```

Rules:

- Activity rows are append-only.
- Do not update historical activity records in normal application flow.
- `metadata` should contain only supplemental structured information.

Example metadata:

```json
{
  "previous_assignee_id": 10,
  "new_assignee_id": 18
}
```

---

# 12. Ticket Chat

## 12.1 `ticket_messages`

Purpose:

Store chat messages inside a ticket.

Columns:

```text
id              BIGSERIAL PRIMARY KEY
ticket_id       BIGINT NOT NULL
sender_id       BIGINT NOT NULL
content         TEXT NULL
created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
edited_at       TIMESTAMPTZ NULL
deleted_at      TIMESTAMPTZ NULL
```

Foreign keys:

```text
ticket_id -> tickets.id
sender_id -> users.id
```

Rules:

- Message may contain text, attachment(s), or both.
- Do not physically delete normal chat messages.
- Use `deleted_at` for soft deletion if deletion is later supported.

Indexes:

```text
ticket_id + created_at
sender_id + created_at
```

Required:

```sql
CREATE INDEX idx_ticket_messages_ticket_created
ON ticket_messages (ticket_id, created_at);
```

---

## 12.2 `message_attachments`

Purpose:

Store metadata for images/files sent in ticket chat.

Binary file data must not be stored directly in PostgreSQL.

Columns:

```text
id              BIGSERIAL PRIMARY KEY
message_id      BIGINT NOT NULL
file_name       VARCHAR(255) NOT NULL
storage_key     TEXT NOT NULL
public_url      TEXT NULL
mime_type       VARCHAR(150) NOT NULL
file_size_bytes BIGINT NOT NULL
width           INTEGER NULL
height          INTEGER NULL
created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
```

Foreign key:

```text
message_id -> ticket_messages.id
```

Constraints:

```text
file_size_bytes >= 0
width >= 0
height >= 0
```

Indexes:

```text
message_id
```

Rules:

- Actual files should later live in object storage.
- Database stores only file metadata and object/storage references.

---

# 13. Checklist

## 13.1 `checklist_items`

Purpose:

Store checklist tasks associated with a ticket.

Columns:

```text
id                  BIGSERIAL PRIMARY KEY
ticket_id           BIGINT NOT NULL
title               VARCHAR(255) NOT NULL
sort_order          INTEGER NOT NULL DEFAULT 0

is_completed        BOOLEAN NOT NULL DEFAULT FALSE
completed_by        BIGINT NULL
completed_at        TIMESTAMPTZ NULL

assigned_to         BIGINT NULL

created_by          BIGINT NOT NULL
created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
```

Foreign keys:

```text
ticket_id      -> tickets.id
completed_by   -> users.id
assigned_to    -> users.id
created_by     -> users.id
```

Indexes:

```text
ticket_id + sort_order
ticket_id + is_completed
assigned_to + is_completed
```

Rules:

- Completed item should contain `completed_by` and `completed_at`.
- Checklist order must be deterministic.

---

# 14. Announcements

## 14.1 `announcements`

Purpose:

Store admin-created announcements shown to staff.

Columns:

```text
id              BIGSERIAL PRIMARY KEY
title           VARCHAR(255) NOT NULL
content         TEXT NOT NULL
author_id       BIGINT NOT NULL

is_published    BOOLEAN NOT NULL DEFAULT FALSE
published_at    TIMESTAMPTZ NULL

created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
```

Foreign key:

```text
author_id -> users.id
```

Indexes:

```text
is_published + published_at
author_id + created_at
```

Recommended:

```sql
CREATE INDEX idx_announcements_published_at
ON announcements (published_at DESC)
WHERE is_published = TRUE;
```

Author history is served by `(author_id, created_at DESC, id DESC)`.

Rules:

- Draft announcements are allowed.
- Only published announcements appear in normal staff UI.
- Pagination must be supported.

---

# 15. Staff Meal

## 15.1 `staff_meals`

Purpose:

Store the current or historical Staff Meal menu image uploaded by admin.

The UI uses a single complete menu image, not one image per day.

Columns:

```text
id                  BIGSERIAL PRIMARY KEY
title               VARCHAR(255) NULL
image_storage_key   TEXT NOT NULL
image_url           TEXT NULL

valid_from          DATE NULL
valid_to            DATE NULL

is_active           BOOLEAN NOT NULL DEFAULT TRUE
uploaded_by         BIGINT NOT NULL

created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
```

Foreign key:

```text
uploaded_by -> users.id
```

Indexes:

```text
is_active + created_at
valid_from + valid_to
```

Rules:

- The application may keep historical Staff Meal entries.
- Normal UI should show the currently active menu.
- Image binary data is not stored in PostgreSQL.

---

# 16. Notifications

## 16.1 `notifications`

Purpose:

Store in-app notifications.

Columns:

```text
id                  BIGSERIAL PRIMARY KEY
user_id             BIGINT NOT NULL
type                notification_type NOT NULL
title               VARCHAR(255) NOT NULL
message             TEXT NULL

ticket_id           BIGINT NULL
announcement_id     BIGINT NULL

is_read             BOOLEAN NOT NULL DEFAULT FALSE
read_at             TIMESTAMPTZ NULL

created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
```

Foreign keys:

```text
user_id         -> users.id
ticket_id       -> tickets.id
announcement_id -> announcements.id
```

Indexes:

```text
user_id + is_read + created_at
user_id + created_at
```

Recommended:

```sql
CREATE INDEX idx_notifications_user_unread_created
ON notifications (user_id, created_at DESC)
WHERE is_read = FALSE;
```

---

# 17. Audit Logging

## 17.1 `audit_logs`

Purpose:

Track important administrative and security-sensitive actions.

Columns:

```text
id              BIGSERIAL PRIMARY KEY
actor_user_id   BIGINT NULL
action          audit_action NOT NULL
entity_type     VARCHAR(100) NOT NULL
entity_id       BIGINT NULL
metadata        JSONB NULL
ip_address      INET NULL
created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
```

Foreign key:

```text
actor_user_id -> users.id
```

Indexes:

```text
actor_user_id + created_at
entity_type + entity_id + created_at
action + created_at
```

Examples:

```text
login
logout
password change
admin user creation
admin user disable
ticket close
announcement publish
```

Rules:

- Audit records are append-only in normal application flow.
- Do not place secrets, raw passwords, session tokens, or sensitive credentials in `metadata`.

---

# 18. Reporting Support

No separate reporting database is required in SPEC-01.

Reports should initially query PostgreSQL directly.

The schema must support:

```text
total tickets
open tickets
closed tickets
tickets by department
tickets by category if categories are later added
tickets by status
tickets by requester
tickets by assignee
tickets over time
average acceptance time
average close time
```

Do not create premature materialized views.

Add reporting indexes only after real queries are known and measured.

---

# 19. `created_at` and `updated_at`

Use:

```text
TIMESTAMPTZ
```

Do not use timezone-naive timestamps for application events.

All timestamps should be stored in PostgreSQL with timezone awareness.

Frontend is responsible for rendering local time.

---

# 20. Update Timestamp Strategy

Use one simple shared PostgreSQL trigger function for `updated_at`, or explicitly update timestamps in SQL.

Preferred implementation for consistency:

```text
set_updated_at()
```

Apply only to tables containing `updated_at`.

Do not create unnecessary database trigger complexity beyond this.

---

# 21. Delete Strategy

## Hard delete allowed

Only for disposable development/test data where no production history is required.

## Soft delete / deactivate preferred

Use existing state columns:

```text
users.is_active
departments.is_active
locations.is_active
ticket_messages.deleted_at
```

Tickets must not be physically deleted in normal application behavior.

Audit logs should not be deleted by normal application operations.

---

# 22. Foreign Key Delete Rules

Use conservative deletion rules.

Default:

```text
ON DELETE RESTRICT
```

Where historical relationships must survive.

Use:

```text
ON DELETE SET NULL
```

only where losing the direct reference is acceptable.

Avoid cascading deletion of:

```text
tickets
ticket_messages
ticket_activity
audit_logs
```

Do not allow deleting a user to erase business history.

---

# 23. Search Support

Initial ticket search must be supportable by:

```text
ticket id
title
location
requester
```

Do not add Elasticsearch.

Do not add PostgreSQL full-text search unless simple indexed queries become insufficient.

Start simple.

---

# 24. Pagination Support

All potentially large lists must be designed for pagination.

Examples:

```text
tickets
ticket_messages
announcements
notifications
audit_logs
```

Initial implementation may use offset pagination.

Cursor pagination may be introduced later if measured workload requires it.

---

# 25. Seed Data

Development seed data should include:

## Departments

At minimum:

```text
IT
HK
FO
ENG
HR
FB
```

Example names:

```text
IT Department
Housekeeping
Front Office
Engineering
Human Resources
Food & Beverage
```

## Admin user

Provide a development-only seed admin.

The initial development admin uses:

```text
username: hothienty
role: admin
department: IT
email: NULL
```

The development password is read from local-only environment configuration and
hashed with Argon2id by Go before insertion. Do not place the plaintext
password or a reusable credential in a public SQL migration.

Never commit a real production password.

Password must be hashed.

## Locations

Provide representative development examples:

```text
Lobby
Ballroom
Back Office
Room 8020
Room 7309
Villa
```

Seed scripts must be clearly marked as development data.

The normal server migration path must not insert these fixtures into a
production database. The historical `0018_seed_development.sql` entry remains
in the sequential migration list because an existing database may already
have its checksum recorded in `_sqlx_migrations`. The Go migration runner
therefore records that entry as successful with its original checksum without
executing its development-only SQL. The guarded
`cmd/seed_development` workflow inserts the departments and locations through
real PostgreSQL transactions, then seeds the development admin.

---

# 26. SQL Migration Structure

Use sequential migration files under `backend/migrations/`.

Example:

```text
backend/migrations/
├── 0001_extensions.sql
├── 0002_enums.sql
├── 0003_departments.sql
├── 0004_users.sql
├── 0005_user_preferences.sql
├── 0006_auth_sessions.sql
├── 0007_locations.sql
├── 0008_tickets.sql
├── 0009_ticket_activity.sql
├── 0010_ticket_messages.sql
├── 0011_message_attachments.sql
├── 0012_checklist_items.sql
├── 0013_announcements.sql
├── 0014_staff_meals.sql
├── 0015_notifications.sql
├── 0016_audit_logs.sql
├── 0017_indexes.sql
├── 0018_seed_development.sql
└── 0019_announcement_author_index.sql
```

Exact filenames may differ if there is a clear reason.

Do not create one giant migration containing the entire database if splitting improves readability.

Do not modify or rewrite an already-applied migration solely to change its
behavior. Preserve its ledger checksum and use a forward-compatible runner or
explicit seed workflow when a historical migration contains development-only
data.

---

# 27. Go Database Package

SPEC-01 should establish only the minimal database foundation required to connect and run migrations.

Suggested structure:

```text
cmd/
├── server/main.go
└── seed_development/main.go
config/config.go
shared/database.go
admin/seed.go
admin/fixtures.go
shared/security/password.go
```

Example responsibility:

```text
config/config.go
    ↓
read the single required DATABASE_URL plus bounded pool settings

shared/database.go
    ↓
create pgxpool.Pool and ping PostgreSQL
    ↓
```

Do not implement domain repositories yet unless needed only to validate schema access.

---

# 28. Connection Pool

Use pgx `pgxpool.Pool`.

Do not wrap the pool in:

```text
Arc
Mutex
RwLock
```

pgxpool is already designed for concurrent use.

Initial pool configuration should be simple and configurable through environment variables.

Suggested environment variables:

```text
DATABASE_URL
DATABASE_MAX_CONNECTIONS
DATABASE_MIN_CONNECTIONS
DATABASE_ACQUIRE_TIMEOUT_SECONDS
```

Reasonable development defaults may be used.

Configuration must reject a pool maximum above 100 connections or an acquire
timeout above 60 seconds. These are resource-safety ceilings; the defaults
remain small (`10` maximum connections, `1` minimum connection, and a `5`
second acquire timeout) and should be changed only with measured evidence.

Do not prematurely tune the pool for unrealistic traffic.

---

# 29. Environment Configuration

Use `.env` only for local development.

The project configuration uses one PostgreSQL connection URL:

```text
DATABASE_URL=postgres://<user>:<local-only-password>@127.0.0.1:5432/bwp-sonasea
```

Do not commit production credentials.

Use the existing local `backend/.env` file for development configuration.
Do not generate `.env.example` or any other example environment file. Never
commit local or production credentials.

---

# 30. Security Requirements

The database design must ensure:

- Plaintext passwords are never stored.
- Raw session tokens are never stored.
- Production secrets are never committed.
- Foreign key relationships preserve history.
- Audit logs do not contain secrets.
- SQL is parameterized.
- No string-concatenated SQL built from user input.
- Disabled users are preserved rather than deleted.
- Authentication session expiry is represented in the schema.

---

# 31. Performance Requirements

The project has an extremely high database-performance bar. SPEC-01 must make
the fast path predictable for future Fiber + pgx/pgxpool handlers while keeping the
schema correct and maintainable.

Performance optimization must be tied to known access patterns rather than
random index accumulation.

Required:

- Proper primary keys
- Foreign key indexes where needed
- Ticket status/date indexes
- Ticket assignment indexes
- Message ticket/date index
- Announcement published index
- Notification unread index
- Composite indexes must follow the relevant `WHERE`, `JOIN`, and `ORDER BY`
  columns in that order.
- Growing lists must have bounded pagination and deterministic ordering; use an
  `id` tie-breaker with timestamp ordering where the future query needs it.
- Query code must use explicit column lists, avoid N+1 round trips, and keep
  transactions short.
- Pool size and acquire timeout must be bounded and configurable.
- Representative query plans must be inspected with `EXPLAIN (ANALYZE, BUFFERS)`
  before performance is claimed.

Every index must serve a documented access pattern. Redundant indexes are not
an optimization because they increase write cost, storage, vacuum work, and
planner choices.

Do not add:

- Redis
- Materialized views
- Read replicas
- Sharding
- Partitioning
- Elasticsearch
- Event streaming
- Message queues

unless a future measured requirement justifies them.

---

# 32. Database Tests

SPEC-01 must include database-focused validation.

At minimum verify:

1. A clean database can run all migrations successfully.
2. Required tables exist.
3. Required enum values exist.
4. Foreign keys reject invalid references.
5. Unique usernames are enforced.
6. Unique employee codes are enforced.
7. Unique non-null emails are enforced.
8. Ticket statuses accept only valid values.
9. Core ticket indexes exist.
10. Chat index exists.
11. Development seed data can be inserted successfully.
12. The development admin password is stored as an Argon2id hash.
13. pgxpool can establish and ping a PostgreSQL connection.

Tests should not depend on production data.

---

# 33. Verification Commands

Before SPEC-01 is considered complete:

```bash
gofmt -d .
go vet ./...
go test ./...
go build ./...
go mod tidy
```

The Go migration runner must also successfully run against a fresh PostgreSQL database.

Example:

```bash
go run ./cmd/server
go run ./cmd/seed_development
```

If an existing development database is used, the migration history must also remain valid.

---

# 34. Non-Goals

SPEC-01 does **not** implement:

- Login API
- Logout API
- `/auth/me`
- Password verification as part of an authentication/login flow
- Cookie handling
- Authorization middleware
- Ticket CRUD APIs
- Ticket WebSocket
- Realtime chat
- File uploads
- SvelteKit UI
- Login screen
- Ticket screen
- Report screen
- Settings screen
- Admin UI
- Redis
- Docker production deployment
- Caddy
- Kubernetes
- Microservices

These belong to later specs.

---

# 35. Definition of Done

SPEC-01 is complete only when all of the following are true:

- PostgreSQL schema supports all currently planned BWP SonaSea features.
- All required tables exist.
- All required enums exist.
- All core relationships use valid foreign keys.
- Critical uniqueness constraints exist.
- Required indexes exist.
- Ticket status is limited to `pending`, `accepted`, and `closed`.
- Assignment is modeled separately from ticket status.
- Ticket history can be preserved.
- Chat messages and attachment metadata can be stored.
- Checklist items can be stored.
- Staff Meal menu metadata can be stored.
- Announcements can be stored and paginated.
- Notifications can be stored and marked read.
- Audit logs can preserve important actions.
- User preferences can store dark/light/system theme.
- Auth sessions can support future server-side session authentication.
- A development admin can be provisioned by username without storing a plaintext password.
- pgxpool can connect to PostgreSQL.
- All SQL migrations run from zero on a clean database.
- `gofmt` passes.
- `go vet ./...` passes.
- `go test ./...` passes.
- `go build ./...` passes.
- No unnecessary Go abstractions were introduced.
- No frontend work was started.
- No authentication business logic was started.

---

# 36. Handoff to SPEC-02

After SPEC-01 is complete, the project should be ready for:

```text
SPEC-02 — Backend Foundation
```

SPEC-02 will build the reusable Fiber application foundation around the database:

```text
config
AppState
pgxpool.Pool
routing
error handling
logging
CORS
health check
graceful shutdown
```

After that:

```text
SPEC-03 — Authentication Backend
```

will implement:

```text
POST /api/v1/auth/login
POST /api/v1/auth/logout
GET  /api/v1/auth/me
```

using the database created in SPEC-01.

The login request will use `username` and password. It will not use email,
public signup, or forgot-password/email-reset flows.

---

# 37. Final Implementation Rule

When implementing this specification:

> Prefer boring, explicit, production-quality Go.

Do not make the code more abstract merely to make it look advanced.

Optimize database structure around real access patterns.

Keep the schema easy to inspect, easy to migrate, and easy to understand several years later.
