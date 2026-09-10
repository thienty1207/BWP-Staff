# Manual database scripts

Run these scripts only against the database you explicitly selected. They do
not contain passwords.

- `backup_database.ps1` — reads the existing `backend/.env` and refreshes the
  full local `.dump` backup plus `database/full_app_schema.sql`.
- `verify_foundation.sql` — read-only schema, migration, and index checks.
- `performance_baseline.sql` — bounded ticket queue/history plans with
  `EXPLAIN (ANALYZE, BUFFERS)`.

Schema changes belong in the sequential SQLx migrations under
`backend/migrations/` so application startup and deployment use one source of
truth.
