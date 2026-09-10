<!-- baron:architecture:start -->
# Architecture Boundaries

- Primary owner: `fullstack`
- Extensions: none
- UI clients may depend on declared contracts, never backend/database internals.
- Backend owns authorization and transaction decisions.
- Database migrations are versioned and reviewed with rollback/data-impact proof.
- Shared contracts have one declared owner and compatibility tests.
- Infrastructure depends on deployable interfaces, not application internals.
<!-- baron:architecture:end -->
