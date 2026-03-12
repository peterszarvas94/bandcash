# Bandcash

Bandcash is a Go web app for managing shared group expenses.

## Tech Stack

- Go, Echo
- Datastar, templ
- SQLite, sqlc, goose

## Quick Start

```bash
mise trust
mise install
op signin
mise run goose-up
mise run dev
```

Open `http://localhost:2222`.

## Varlock + 1Password

Environment loading is handled only by `varlock` + 1Password.

Files:

- `.env.schema` is committed (schema + defaults)
- `.env.local` is gitignored (dev secret-zero)
- `.env.production.local` is gitignored (prod secret-zero for local deploy)

For dev:

```bash
cp .env.local.example .env.local
```

Put this in `.env.local`:

```bash
APP_ENV=development
OP_TOKEN=...
OP_ENVIRONMENT_ID=...
```

Then run:

```bash
mise run dev
```

For prod:

- Copy the same way as dev:

```bash
cp .env.production.local.example .env.production.local
```

- Put production values into `.env.production.local`:

```bash
APP_ENV=production
OP_TOKEN=...
OP_ENVIRONMENT_ID=...
```

- Deploy with:

```bash
mise run deploy
```

`mise run deploy` explicitly sources `.env.production.local` via `scripts/with_varlock_env.sh`, then runs `varlock` in production mode.

## Commands

```bash
mise run dev          # hot reload
mise run run          # run server directly
mise run build        # build binary (tmp/server)
mise run test         # go test ./...
mise run test-pretty  # gotestsum testdox output
mise run format       # go fmt ./...
mise run vet          # go vet ./...
mise run check        # format + vet + test
mise run lint         # golangci-lint run
mise run templ        # regenerate *_templ.go
mise run sqlc         # regenerate sqlc output
mise run goose-up     # apply migrations
mise run goose-status # migration status
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
│   │   ├── migrations/        # goose migrations
│   │   └── queries/           # sqlc query sources
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
│   ├── settings/
│   └── shared/                # shared templ components
├── static/                    # CSS and JS assets
├── mise.toml                  # tools and task runner
└── AGENTS.md                  # contributor/agent coding guide
```

## Notes

- Generated files are not edited manually:
  - `*_templ.go` (run `mise run templ`)
  - `internal/db/*.sql.go` (run `mise run sqlc`)

## License

[O'Saasy](https://osaasy.dev/) (c) 2026 Peter Szarvas. See `LICENSE.md`.
