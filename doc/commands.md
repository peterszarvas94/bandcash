# Commands

Use `mise run <task>` first; fall back to raw `go test` only for targeted test runs.

## Run and Dev

- `mise run dev` - start app with hot reload (`air`) only.
- `mise run tunnel` - run app tunnel (`cloudflared tunnel run --token "$(cloudflared tunnel token bandcash)" --url http://localhost:2222`).
- `mise run tunnel-dev` - start app (`air`) and tunnel in parallel.
- `mise run run` - run server directly with `go run`.
- `mise run build` - compile server binary to `tmp/server`.
- `mise run start` - run built binary.
- `mise run routes` - print registered routes and exit.

Dev env loading:

- `mise run dev`, `mise run tunnel-dev`, and `mise run run` can auto-load billing-related secrets from 1Password via `kamal secrets` before starting the server.
- Set `OP_ACCOUNT` and one of `OP_FROM_LOCALHOST` (preferred for local), `OP_FROM_DEVELOPMENT`, or `OP_FROM` in your shell, then run dev commands normally.
- Local 1Password entries should include full app env keys (clear + secret), e.g. `APP_ENV`, `PORT`, `URL`, `DB_PATH`, logging keys, `EMAIL_PROVIDER`, `EMAIL_FROM`, Mailtrap SMTP keys (`MAILTRAP_HOST`, `MAILTRAP_PORT`, `MAILTRAP_USERNAME`, `MAILTRAP_PASSWORD`) for sandbox, and Lemon keys (`LEMON_WEBHOOK_SECRET`, `LEMON_API_KEY`, `LEMON_CHECKOUT_URL`).
- This avoids storing plaintext local secret files while still enabling local webhook/API billing flows.
- `mise run tunnel` and `mise run tunnel-dev` expect access to the `bandcash` Cloudflare tunnel token via `cloudflared tunnel token bandcash`.
- If your managed tunnel has no ingress rule, the `--url http://localhost:2222` fallback keeps local dev routing working.

## Formatting and Static Checks

- `mise run format` - run `go fmt ./...`.
- `mise run vet` - run `go vet ./...`.
- `mise run lint` - run `golangci-lint run`.
- `mise run lsp` - run `gopls check` against tracked Go files (excluding generated DB files).
- `mise run check` - run format + vet + test in sequence.

## Tests

- `mise run test` - run all tests (`go test ./...`).
- `mise run test-pretty` - run tests with gotestsum output.
- Single package: `go test ./models/event`.
- Single test by name: `go test ./models/event -run TestHandlerCreate`.
- Single test exact match: `go test ./models/event -run '^TestHandlerCreate$'`.
- Multiple tests by regex: `go test ./models/event -run 'TestCreate|TestUpdate'`.
- Test one package with race detector: `go test -race ./models/event`.
- Full race run: `go test -race ./...`.
- Coverage profile: `go test -coverprofile=coverage.out ./...`.

Quick rules:

- `-run` filters by test function name regex inside the selected package.
- Go tests run by package, not by file path.
- After query/schema changes, run affected package tests and a compile check.
- dont build, because the dev server is running always, unless explicitly asked to debug

## DB Migrations

- `mise run db-up` - apply pending Bun SQL migrations.
- `mise run db-down` - rollback last Bun migration group.
- `mise run db-status` - show migration status.
- `mise run db-create name=add_new_column` - create new SQL migration files.
