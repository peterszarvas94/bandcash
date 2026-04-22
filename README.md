# Bandcash

Bandcash is a Go web app for managing shared group expenses.

## Tech Stack

- Go, Echo
- Datastar, templ
- SQLite, Bun ORM + Bun migrations

## Quick Start

```bash
mise trust
mise install
mise run db-up
mise run dev
```

Open `http://localhost:2222`.

## Environment

Development runs with built-in defaults from `internal/utils/env.go`.

Production deploy uses Kamal secrets.

- Put all production env values into `.kamal/secrets` (gitignored).
- Keep sensitive values there too (registry, smtp user/password, better stack tokens).
- Deploy with `mise run deploy`.

## Commands

```bash
mise run dev          # hot reload
mise run tunnel       # app tunnel only
mise run tunnel-dev   # hot reload + app tunnel
mise run run          # run server directly
mise run build        # build binary (tmp/server)
mise run test         # go test ./...
mise run test-pretty  # gotestsum testdox output
mise run format       # go fmt ./...
mise run vet          # go vet ./...
mise run check        # format + vet + test
mise run lint         # golangci-lint run
mise run templ        # regenerate *_templ.go
mise run db-up        # apply migrations
mise run db-down      # rollback last migration group
mise run db-status    # migration status
mise run db-create name=add_new_column # create migration files
mise run seed         # seed local data
```

## Deployment

```bash
mise run kamal deploy
```

Useful commands:

```bash
mise run kamal setup
mise run kamal app logs
mise run kamal app details
./scripts/pull_prod_db.sh
```

## Project Structure

```text
.
├── cmd/
│   ├── server/                # app bootstrap and route wiring
│   └── seed/                  # local/dev seed command
├── internal/
│   ├── db/
│   │   ├── bunmigrations/     # Bun SQL migrations
│   │   └── (Bun query code)   # typed + dynamic DB layer
│   ├── middleware/            # auth/guard middleware
│   ├── utils/                 # shared helpers
│   ├── i18n/                  # localization
│   └── email/                 # email rendering/sending
├── models/                    # feature modules
│   ├── auth/
│   ├── group/
│   ├── event/
│   ├── member/
│   ├── expense/
│   ├── account/
│   └── shared/                # shared templ components
├── static/                    # CSS and JS assets
├── mise.toml                  # tools and task runner
└── AGENTS.md                  # contributor/agent coding guide
```

## Notes

- Generated files are not edited manually:
  - `*_templ.go` (run `mise run templ`)
  - Keep typed DB access in `internal/db/*.go` and Bun migrations in `internal/db/bunmigrations/*.sql`.

## License

[O'Saasy](https://osaasy.dev/) (c) 2026 Peter Szarvas. See `LICENSE.md`.
