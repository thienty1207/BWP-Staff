# Baron Intent Brief

- ID: `intent-1641ef668cc9652e`
- Title: SPEC-01 database foundation
- Risk: `medium`
- Confirmation: `confirmed`
- Updated: 2026-09-10T17:41:01+07:00

## Current Behavior

Backend is only an Axum health-check scaffold; no database schema, migrations, PgPool, or admin seed exists.

## Target Behavior

Backend has SQLx PostgreSQL migrations for the SPEC-01 schema, PgPool/config integration, an idempotent development admin seed using username login and Argon2id, and verified database tests.

## Scope

backend database foundation and the project context/spec authentication rules; no frontend or auth API implementation

## Non-Goals

- login/logout API
- public signup
- forgot-password or email-reset flow
- ticket/chat/UI business features

## Constraints

- Use the existing split DATABASE_HOST, DATABASE_PORT, DATABASE_NAME, DATABASE_USER, and DATABASE_PASSWORD environment variables.
- Never commit local credentials or plaintext passwords.
- Use Bun only for SvelteKit; no frontend changes in SPEC-01.

## Decisions

- Use username as the only login identifier; email is optional profile/contact data.
- Provision hothienty as the development admin; hash its local password with Argon2id before database insertion.

## Required Proof

Fresh PostgreSQL migration, schema/index/constraint checks, Argon2id seed verification, cargo fmt, cargo clippy with warnings denied, and cargo test.

## Remaining Unknowns

- none recorded

## Agent Rules

- Read project, Vault, plan, Harness, and prior decisions before asking the user.
- Ask one missing high-value question at a time.
- Do not treat unknowns as facts.
- Medium/high-risk implementation requires this intent to be confirmed.
