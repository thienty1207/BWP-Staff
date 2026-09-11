# Database

This directory contains the database backup artifacts, the full application
schema script, manual checks, and operational notes.

- `bwp-sonasea.dump` is the local PostgreSQL custom-format backup containing
  the current schema and data. Keep this file available for crash recovery;
  it is intentionally ignored by Git because it contains database data.
- `full_app_schema.sql` is the complete schema script for the application,
  generated from the current database without application rows or credentials.
- `scripts/backup_database.ps1` refreshes both files from the existing
  `backend/.env` `DATABASE_URL` configuration.

- `scripts/verify_foundation.sql` checks the foundation without changing data.
- `scripts/performance_baseline.sql` runs bounded representative `EXPLAIN`
  checks for the main ticket read paths.

Refresh the backup from the repository root with:

```powershell
pwsh -File database/scripts/backup_database.ps1
```

Runtime SQL migrations intentionally live in `backend/migrations/` because the
Go backend checks and runs them at startup. The Go runner keeps the existing
`_sqlx_migrations` ledger compatible with the database already in use; the
legacy ledger name is retained for data safety, and the runner is independent
of the old framework. The migration directory contains schema migrations only;
the historical fixture-only version 0018 is no longer in the active path.
The revised ticket foundation is applied by the forward-only
0020_ticket_assignment_and_request_fields.sql migration; version 0018 is
never reused. It adds boolean priority, optional due time, ticket-level
attachment metadata, and normalized department/user assignment tables while
preserving legacy single-user assignments before removing those columns.
Existing databases keep any old version-18 ledger row and data; startup does
not reset or delete them. Development departments, locations, and the admin
are inserted only by the explicitly guarded `go run ./cmd/seed_development`
command.
Do not copy or move those migrations into this directory. The dump and local
credentials are ignored by the root `.gitignore`; the schema script and
operational scripts are source files and remain reviewable.

The initial index set follows the known read paths: ticket queues and history
use status/owner/department/location plus timestamp and `id` tie-breakers;
ticket assignments use membership indexes; ticket-level attachments, chat and
activity use ticket plus timestamp; published announcements and unread
notifications use partial indexes; session expiry and audit lookup have
dedicated indexes. New indexes should be added only for a measured query
pattern and validated with `EXPLAIN (ANALYZE, BUFFERS)`.
