# Baron Intent Brief

- ID: `intent-a9019e10c05faeee`
- Title: SPEC-05 authenticated shell and tickets read/list foundation
- Risk: `high`
- Confirmation: `confirmed`
- Updated: 2026-09-11T19:52:56+07:00

## Current Behavior

SPEC-04 authenticated identity entry page with no ticket list endpoint

## Target Behavior

Real authenticated shell with GET /api/v1/tickets, Open/Closed keyset-paginated lists, and responsive desktop/mobile UI

## Scope

backend/client/tickets, frontend authenticated shell and ticket list, focused tests, current repository snapshot, SPEC-05 documentation

## Non-Goals

- No ticket creation/detail/actions/chat/checklist/uploads/realtime/search/report/settings/admin UI or future SPEC implementation
- No migrations/schema changes, no runtime mock ticket data, no new environment variables

## Constraints

- Go Fiber v3 PostgreSQL pgx v5 pgxpool explicit SQL SvelteKit Svelte TypeScript Bun and server-side session auth remain locked

## Decisions

- Use bounded keyset pagination and three bounded SQL reads at most per page: base page, assigned departments, assigned users

## Required Proof

gofmt, go mod tidy, go vet, go test, go build, Bun install/check/build/test, real PostgreSQL temporary-schema ticket integration tests, real browser auth and ticket-list flow, responsive widths, schema and scope diff review

## Remaining Unknowns

- none recorded

## Agent Rules

- Read project, Vault, plan, Harness, and prior decisions before asking the user.
- Ask one missing high-value question at a time.
- Do not treat unknowns as facts.
- Medium/high-risk implementation requires this intent to be confirmed.
