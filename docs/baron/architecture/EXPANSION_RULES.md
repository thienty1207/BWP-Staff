<!-- baron:architecture:start -->
# Expansion Rules

- Current primary: `fullstack`
- Current extensions: none
- `baron init --<platform>` on an initialized project adds an extension; it does not replace the primary.
- Reconcile shared contracts, auth, data ownership, deployment, and tests before adding code.
- Existing files are never moved or deleted merely to match a suggested layout.
- Structural migration requires plan, dry-run inventory, rollback, and proof.
- Architecture drift produces a correction proposal; automatic destructive restructuring is forbidden.
<!-- baron:architecture:end -->
