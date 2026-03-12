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

Environment loading is handled by `varlock` and 1Password.

- Schema: `.env.schema`
- Local secrets file: `.env.local` (gitignored)
- Production local secrets file: `.env.production.local` (gitignored, optional)
- Runtime/loading command wrapper: `varlock run -- ...`

Initial setup:

```bash
mise install
cp .env.local.example .env.local
```

Then set:

1. `OP_TOKEN` in `.env.local` (gitignored) for local runs.
2. `OP_ENVIRONMENT_ID` in `.env.local` for your local 1Password environment.
3. Ensure 1Password environments include deploy keys too (`KAMAL_REGISTRY_USERNAME`, `KAMAL_REGISTRY_PASSWORD`) where needed.

Varlock file precedence follows the default convention:

1. `.env.schema`
2. `.env.local`
3. `.env.production.local` (when `APP_ENV=production`)
4. process environment variables

For production/CI, inject `APP_ENV`, `OP_ENVIRONMENT_ID`, and `OP_TOKEN` from your platform secret store (do not commit them).

If you deploy from your local machine and do not want to export env vars manually, create `.env.production.local` (gitignored) with:

```bash
OP_TOKEN=...
OP_ENVIRONMENT_ID=...
```

`mise run deploy` already forces `APP_ENV=production`, so varlock will load `.env.production.local` for the deploy command.

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
