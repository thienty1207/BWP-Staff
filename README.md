# BWP-SonaSea

Monorepo layout for the SonaSea application:

```text
frontend/  SvelteKit + TypeScript, managed with Bun
backend/   Rust + Axum API
database/  PostgreSQL dump, full schema script, and database operations
docker/    reserved for container files when deployment needs them
```

## Development

Start the frontend:

```bash
cd frontend
bun run dev
```

Start the backend:

```bash
cd backend
cargo run
```

Before starting it, make sure the existing local `backend/.env` contains the
PostgreSQL values. Do not create or commit an example env file. The backend
runs the SQLx migrations from `backend/migrations/` before serving requests.

To initialize development reference data and the development admin account,
run:

```bash
cd backend
cargo run --bin seed_development
```

The development account uses username login. Accounts are created by an admin;
there is no public signup, email login, or forgot-password flow.

The backend exposes `GET http://127.0.0.1:3000/health` and returns `ok` after
the database foundation is ready.

The local full database backup is `database/bwp-sonasea.dump`; refresh it, and
the matching `database/full_app_schema.sql`, with:

```powershell
pwsh -File database/scripts/backup_database.ps1
```

The requested backend module boundaries live under `backend/src/`. Runtime
SQLx migrations remain under `backend/migrations/`; the root `database/`
directory is for the backup and operational schema artifacts.

Baron is initialized for the Codex integration at the project root. Baron Core
files live under `.baron/core/`; host-specific Codex files remain in `.codex/`.
