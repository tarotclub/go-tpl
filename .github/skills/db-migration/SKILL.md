---
name: db-migration
description: Create, review, or update database migrations for this Go project. Use when adding tables, altering columns, creating indexes, backfilling data, planning rollbacks, or changing persistence-related schema and SQL migration files. Ask the user first before any schema change or introducing a new migration dependency.
---

# DB Migration

## Overview

Handle database schema and data migrations conservatively. Prefer reversible, incremental changes that keep the application deployable throughout the migration window. This skill assumes the migration tool is golang-migrate. In this repository, schema changes and new dependencies require explicit user confirmation before proceeding.

## When to Use

- Creating a new migration
- Modifying an existing migration that has not shipped yet
- Adding or changing tables, columns, constraints, indexes, or views
- Writing data backfills or data cleanup SQL tied to a schema change
- Designing rollout and rollback steps for persistence changes
- Reviewing migration safety before code changes land

**Ask first:**

- Any database schema change
- Any new migration tool or database dependency
- Any destructive step such as dropping a column, rewriting large tables, or irreversible data transforms

## Goals

1. Keep migrations small, reviewable, and reversible.
2. Separate schema evolution from application behavior when practical.
3. Preserve compatibility during rolling deploys.
4. Verify the migration path, not just the final schema.

## Workflow

### 1. Confirm the Existing Migration Convention

Before editing anything, inspect the repository for:

- Existing migration directory and naming convention
- Database driver and dialect in use
- Whether migrations are raw SQL, embedded files, or Go-based migrations

This skill standardizes on golang-migrate. Confirm whether the repository already uses it and whether the migration files follow its paired file convention.

If golang-migrate is not present yet, do not introduce it silently. Ask the user first, because adding a migration tool is a new dependency and crosses a repository boundary rule.

### 2. Classify the Change

Decide which category the request falls into:

- Additive schema change: new table, nullable column, new index
- Contracting schema change: dropping column, tightening constraint, renaming field
- Data migration: backfill, normalization, deduplication
- Operational migration: index concurrently, repartitioning, long-running rewrite

Prefer additive-first rollout. For risky or breaking changes, use expand-and-contract:

1. Add new schema in a backward-compatible way.
2. Deploy application code that writes to both shapes or reads from the new shape.
3. Backfill data.
4. Switch reads fully.
5. Remove old schema in a later migration.

Do not collapse these steps into one migration unless the system is small, offline-only, or the user explicitly accepts the risk.

### 3. Author the Migration

Follow the repository's existing naming convention. For golang-migrate, prefer paired files in the form `NNNNNN_name.up.sql` and `NNNNNN_name.down.sql` under a dedicated migration directory.

Migration authoring rules:

- One logical change per migration.
- Keep `up` and `down` symmetric when feasible.
- Avoid editing an already-applied migration; create a new corrective migration instead.
- Use explicit constraint and index names.
- Avoid irreversible statements unless the user approved them.
- Keep application-specific backfill code out of ad hoc SQL when the logic is complex; prefer a separate, deliberate backfill step.

For data backfills:

- Make the migration resumable where possible.
- Avoid locking the entire table unless the user approved the downtime.
- Batch large updates if the database size or availability requirements justify it.

### 4. Validate Compatibility

Check that the migration is safe against the code rollout order:

- Old app with new schema
- New app with old schema, when applicable
- Partial rollout across multiple instances

Watch for these common hazards:

- Adding a non-null column without a default to a populated table
- Renaming columns that running code still references
- Creating unique constraints before deduplicating existing rows
- Backfills that assume data quality that does not actually hold
- DDL that blocks traffic for too long on large tables

### 5. Update Nearby Code Deliberately

Only after the migration strategy is clear:

- Update Go structs, repositories, queries, and config as needed
- Add or update tests that prove the new behavior
- Update docs only if the migration changes developer workflow or operational steps

Do not mix unrelated refactors with migration work.

## Verification

After making changes, run the narrowest relevant checks first:

1. Migration-specific validation with golang-migrate if the project already has a migration command or test harness
2. Focused package tests for the touched persistence layer
3. `go test ./...`

If the repository has no executable migration validation path yet, state that explicitly and fall back to reviewing:

- File naming convention
- SQL syntax shape and rollback coverage
- Compatibility assumptions

## Output Expectations

When completing a migration task, report:

- What schema or data change was made
- Whether it is backward-compatible
- How rollback works
- What risks remain, especially for large tables or production rollout
- Which tests or checks were run

## Go Project Defaults

For Go repositories, prefer these defaults unless the repo already established a different pattern:

- Keep migration files under a top-level `migrations/` directory
- Keep database connection and runner code under `internal/` if code support is needed
- Use golang-migrate as the migration tool
- Name migration files as paired `.up.sql` and `.down.sql` files compatible with golang-migrate
- Use raw SQL migrations unless the repository already standardizes on code-based migrations
- Use `go test ./...` as the baseline verification step

## Anti-Patterns

Avoid these unless the user explicitly asks for them:

- Silently introducing a migration framework
- Rewriting old applied migrations
- Combining schema changes, data backfill, and unrelated refactors in one step
- Dropping old columns immediately after introducing replacements
- Assuming rollback is unnecessary because the change is "small"
- Using destructive SQL without calling out blast radius