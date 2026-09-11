# SPEC-05 Authenticated Shell and Tickets List

## Goal

Implement only the SPEC-05 authenticated shell and read-only tickets list on
top of the existing SPEC-04 authentication flow. Keep PostgreSQL schema and
migrations unchanged, use real database-backed ticket data, and avoid future
ticket actions or feature areas.

## Work plan

1. Establish the backend contract with failing tests for query validation,
   authentication, open/closed filtering, deterministic keyset pagination,
   compact joined identity fields, nullable values, assignments, and the
   isolated PostgreSQL test setup.
2. Add the small `backend/client/tickets` package with explicit Fiber handler,
   service, repository SQL, models, and route wiring through the existing auth
   service. Keep the page bounded to one base query plus two assignment bulk
   queries.
3. Establish the frontend ticket API contract with failing Bun tests, then add
   the focused typed client using relative URLs and `credentials: 'include'`.
4. Replace the temporary authenticated identity page with the SPEC-05 shell:
   desktop sidebar/table, mobile drawer/cards, Open/Closed keyset loading,
   retry/empty/loading states, and real `/me` identity only. Extend the existing
   CSS-variable theme without introducing a UI framework or fake data.
5. Update the current repository snapshot and commit the supplied SPEC-05
   document, then run the required backend/frontend checks, real PostgreSQL
   integration checks when available, browser verification, diff/scope review,
   and push the focused commit to `main` without staging unrelated worktree
   changes.

## Verification gates

- `gofmt -d .`, `go mod tidy`, `go vet ./...`, `go test ./... -count=1`, and
  `go build ./...` from `backend/`.
- `go test -race ./... -count=1` when the local environment supports it.
- `bun install --frozen-lockfile`, `bun run check`, `bun run build`, and
  `bun test` from `frontend/`.
- PostgreSQL-backed tests use `DATABASE_URL` plus an isolated temporary schema;
  no test database environment variable or runtime mock ticket data is added.
- Manual browser checks cover login, `/me`, open/closed ticket lists, refresh,
  logout, unauthenticated redirect, real empty state, and required viewport
  widths.
