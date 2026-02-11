---
name: sqlc-goose
description: Manage database migrations with goose and type-safe queries with sqlc in this repo.
---

## What I do
- Add or update SQL migrations for the entry database.
- Add or update sqlc queries and regenerate Go code.
- Keep sqlc/goose tooling usage consistent with this repo.
 - Ensure mise tooling is used for `goose` and `sqlc`.

## When to use me
Use this when changing database schema or queries, or when you need to regenerate sqlc code.

## Conventions in this repo
- Migrations live in `internal/db/migrations` and use goose `Up/Down` blocks.
- Queries live in `internal/db/queries` and are compiled by sqlc.
- sqlc config is `sqlc.yaml` and outputs Go code into `internal/db`.
- App startup runs goose via `db.Migrate()` in `internal/db/init.go`.
- Tooling is installed via `mise.toml`; use `goose` and `sqlc` directly after `mise install`.

## Common commands
Install tools:
```bash
mise install
```

Run migrations:
```bash
mise run goose-up
```

Create a new migration:
```bash
mise run goose-create name=add_new_column
```

Check migration status:
```bash
mise run goose-status
```

Regenerate sqlc code:
```bash
mise run sqlc
```

Add a new query:
```bash
vim internal/db/queries/entries.sql
mise run sqlc
```

## Notes
- Generated files in `internal/db/*.go` should not be edited manually.
- After changing SQL queries, always run `sqlc generate`.
- If `goose` or `sqlc` is missing, ensure your shell is activated for mise (e.g. `eval "$(mise activate zsh)"`).
