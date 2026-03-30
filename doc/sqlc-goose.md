# sqlc-goose

## What I do
- Add or update SQL migrations for the app database.
- Add or update sqlc queries and regenerate Go code.
- Keep sqlc/goose tooling usage consistent with this repo.
- Ensure mise tooling is used for `goose` and `sqlc`.

## When to use me
Use this when changing database schema or queries, or when you need to regenerate sqlc code.

## Conventions in this repo
- Migrations live in `internal/db/migrations` and use goose `Up/Down` blocks.
- Queries live in `internal/db/queries` and are compiled by sqlc.
- sqlc config is `sqlc.yaml` and outputs Go code into `internal/db`.
- App startup runs migrations via `db.Migrate()`, but explicit migration commands are still used during schema work.
- Tooling is installed via `mise.toml`; prefer `mise run ...` tasks.

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
${EDITOR:-vim} internal/db/queries/<feature>.sql
mise run sqlc
```

## Notes
- Generated files in `internal/db/*.sql.go` should not be edited manually.
- After changing SQL queries, run `mise run sqlc`.
- Validate affected code with targeted tests (for example `go test ./models/event -run '^TestHandlerCreate$'`) and then broader checks as needed.
