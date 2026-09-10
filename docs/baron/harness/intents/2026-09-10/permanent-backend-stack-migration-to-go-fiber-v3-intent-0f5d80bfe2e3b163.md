# Baron Intent Brief

- ID: `intent-0f5d80bfe2e3b163`
- Title: Permanent backend stack migration to Go/Fiber v3
- Risk: `high`
- Confirmation: `confirmed`
- Updated: 2026-09-10T21:29:38+07:00

## Current Behavior

SPEC-01 backend is a Rust/Axum scaffold with Cargo and SQLx integration

## Target Behavior

SPEC-01 backend uses Go, Fiber v3, pgx v5, pgxpool, explicit SQL migrations, and no Rust artifacts

## Scope

backend foundation, migration runner, development seed, context/spec documentation, database scripts, and verification

## Non-Goals

- future authentication APIs, client feature behavior, frontend changes, and production deployment

## Constraints

- preserve PostgreSQL schema, migration bytes, existing database, local dump, username-only admin seed, and no .env.example

## Decisions

- permanent Rust/Axum to Go/Fiber v3 migration explicitly requested by the user

## Required Proof

gofmt, go vet, go test, go build, Fiber health smoke test, existing and fresh PostgreSQL migration/seed checks, Bun check, backup/schema/index/EXPLAIN checks, and secret/artifact scan

## Remaining Unknowns

- none recorded

## Agent Rules

- Read project, Vault, plan, Harness, and prior decisions before asking the user.
- Ask one missing high-value question at a time.
- Do not treat unknowns as facts.
- Medium/high-risk implementation requires this intent to be confirmed.
