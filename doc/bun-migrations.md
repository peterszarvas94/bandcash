# bun-migrations

## What I do
- Define and evolve database schema with Bun SQL migrations.
- Keep migration workflows consistent with this repository.

## Conventions in this repo
- Bun migrations live in `internal/db/bunmigrations` as `.up.sql` / `.down.sql` files.
- App startup runs migrations via `db.Migrate()` using Bun migrator.
- Use `mise run ...` tasks for migration operations.

## Common commands
Install tools:
```bash
mise install
```

Apply migrations:
```bash
mise run db-up
```

Rollback last migration group:
```bash
mise run db-down
```

Show migration status:
```bash
mise run db-status
```

Create new SQL migration files:
```bash
mise run db-create name=add_new_column
```

## Notes
- Keep schema changes explicit in migration files.
- Prefer additive migrations; keep destructive changes deliberate and reviewed.
