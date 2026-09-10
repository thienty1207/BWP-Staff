# BWP-SonaSea

Monorepo layout for the SonaSea application:

```text
frontend/  SvelteKit + TypeScript, managed with Bun
backend/   Rust + Axum API
database/  schema, migrations, and local database notes
test/      cross-service and end-to-end test notes/assets
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

The backend exposes `GET http://127.0.0.1:3000/health` and returns `ok`.

Baron is initialized for the Codex integration at the project root. Baron Core
files live under `.baron/core/`; host-specific Codex files remain in `.codex/`.
