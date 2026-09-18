# Baron Intent Brief

- ID: `intent-4f42a1bc0ba1de32`
- Title: SPEC-08.2 post-rebrand repository hardening
- Risk: `medium`
- Confirmation: `confirmed`
- Updated: 2026-09-18T19:27:46+07:00

## Current Behavior

Hotel Staff runtime and SPEC-08.1 are closed, but Baron active state is stale and repository hardening gaps remain: legacy public assets, Svelte favicon, permissive robots, frontend env-ignore exceptions, backup target safety, absent CI, and no explicit foreign-Origin mutation guard.

## Target Behavior

Reconcile Baron current state through supported tooling and implement SPEC-08.2 hardening without product/data/migration changes; leave SPEC-09 unimplemented.

## Scope

Baron supported lifecycle outputs; frontend public assets/favicon/robots/.gitignore/tests; backup script/tests; backend Origin middleware/tests; CI workflow; PROJECT_CONTEXT and SPEC-08.2 closure only if every gate passes.

## Non-Goals

- No SPEC-09, no schema migration 0021, no seed/data rewrite, no BWP master-data rename, no deploy topology, no rate limiting, no chat/assign/close behavior.

## Constraints

- Preserve pre-existing unstaged work; never commit backend/.env, dumps, unrelated images/notes/UI changes; use explicit staging only; do not hand-edit Baron-owned identity metadata.

## Decisions

- User explicitly approved immediate SPEC-08.2 implementation; treat legacy Baron project_slug as tooling-only unless a supported rename exists.

## Required Proof

Baron reconciliation/code-map evidence; static reference audit; focused red-green tests; frontend check/test/build; backend gofmt/list/vet/test/race/build; read-only DB before/after; runtime smoke; staged diff audit; GitHub Actions status only if actually observable.

## Remaining Unknowns

- Whether hosted GitHub Actions is accessible after push; no CI pass is assumed.

## Agent Rules

- Read project, Vault, plan, Harness, and prior decisions before asking the user.
- Ask one missing high-value question at a time.
- Do not treat unknowns as facts.
- Medium/high-risk implementation requires this intent to be confirmed.
