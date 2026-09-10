<!-- baron:architecture:start -->
# Project Structure Contract

## Observed Top-Level Paths
- `.agents`
- `.baron-vault`
- `.gitignore`
- `AGENTS.md`
- `README.md`
- `backend`
- `database`
- `docker`
- `frontend`
- `test`

## Adaptive Rule

- Keep existing paths unless an approved migration plan, dry-run inventory, rollback path, and proof justify movement.
- Recommendations describe responsibilities; they do not force framework-incompatible folder names.
- New top-level modules need one responsibility, an owner, allowed dependencies, and a validation path.
- Baron must not move existing paths automatically.

## Suggested Responsibility Areas
- frontend product surfaces
- backend services
- database/migrations
- shared API/data contracts
- infrastructure and end-to-end tests
<!-- baron:architecture:end -->
