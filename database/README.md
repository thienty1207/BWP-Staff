# Database

This directory contains the database backup artifacts, the full application
schema script, manual checks, and operational notes.

- `bwp-sonasea.dump` is the local PostgreSQL custom-format backup containing
  the current schema and data. Keep this file available for crash recovery;
  it is intentionally ignored by Git because it contains database data.
- `full_app_schema.sql` is the complete schema script for the application,
  generated from the current database without application rows or credentials.
- `scripts/backup_database.ps1` refreshes both files from the existing
  `backend/.env` configuration.

- `scripts/verify_foundation.sql` checks the foundation without changing data.
- `scripts/performance_baseline.sql` runs bounded representative `EXPLAIN`
  checks for the main ticket read paths.

Refresh the backup from the repository root with:

```powershell
pwsh -File database/scripts/backup_database.ps1
```

SQLx runtime migrations intentionally live in `backend/migrations/` because
the Rust backend embeds and runs them at startup. Do not copy or move those
migrations into this directory. The dump and local credentials are ignored by
the root `.gitignore`; the schema script and operational scripts are source
files and remain reviewable.

The initial index set follows the known read paths: ticket queues and history
use status/owner/department/location plus timestamp and `id` tie-breakers;
ticket chat and activity use ticket plus timestamp; published announcements
and unread notifications use partial indexes; session expiry and audit lookup
have dedicated indexes. New indexes should be added only for a measured query
pattern and validated with `EXPLAIN (ANALYZE, BUFFERS)`.
