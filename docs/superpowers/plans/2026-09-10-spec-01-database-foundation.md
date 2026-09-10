# SPEC-01 Database Foundation — Go/Fiber v3

This plan records the current foundation after the permanent backend migration.

## Scope

- Keep the existing PostgreSQL schema and all sequential SQL files under
  `backend/migrations/`.
- Run those files from a small Go migration runner that checks the existing
  `_sqlx_migrations` ledger and SHA-384 checksums before applying anything.
- Use `pgx/v5` and `pgxpool` for direct parameterized SQL and bounded pooling.
- Provide a separate development seed command for the admin account; never
  reset an existing password during startup or seed runs.
- Keep the first Fiber application limited to `GET /health`.

## Backend layout

```text
backend/
├── go.mod
├── go.sum
├── migrations/
├── cmd/server/main.go
├── cmd/seed_development/main.go
├── app/app.go
├── config/config.go
├── shared/database.go
└── admin/seed.go
```

No future feature folders, mock APIs, frontend changes, ORM, cache, queue, or
authentication endpoints are part of SPEC-01.

## Configuration

The only database connection setting is the local-only `DATABASE_URL`. Pool
maximum, pool minimum, and acquisition timeout are bounded environment values.
No `.env.example` file is generated or committed.

## Verification

Run from `backend/`:

```text
gofmt -d .
go vet ./...
go test ./...
go build ./...
go mod tidy
go run ./cmd/seed_development
```

The existing development database must remain intact, migration checksums must
match, the seed must be idempotent, and the Fiber health endpoint must return
HTTP 200 with `ok` after PostgreSQL startup checks succeed.

## SPEC-01 hardening pass

This continuation audits the implemented foundation only. It does not add a
later SPEC or application feature.

1. Add real-PostgreSQL behavioral tests in an isolated temporary schema for
   migration idempotency, schema/enums/indexes, relational constraints, the
   development seed workflow, Argon2id storage, and pgxpool connectivity.
2. Keep `0018_seed_development.sql` unchanged for checksum compatibility, but
   make the normal migration runner record that historical migration without
   executing its development-only SQL. Move fixture insertion to the explicit
   guarded development seed command.
3. Move the small Argon2id hash/verify utility out of the admin seed package;
   keep the current PHC behavior and tests without adding an authentication
   flow.
4. Require explicit TLS for non-loopback PostgreSQL URLs while preserving the
   current local development URL behavior; add focused configuration tests.
5. Run the required Go checks, dedicated PostgreSQL checks, security review,
   migration-history checks, and final GitHub push verification.
