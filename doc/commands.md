# Commands

Use `mise run <task>` first; fall back to raw `go test` only for targeted test runs.

## Run and Dev

- `mise run dev` - start app with hot reload (`air`) plus local mail service.
- `mise run run` - run server directly with `go run`.
- `mise run build` - compile server binary to `tmp/server`.
- `mise run start` - run built binary.
- `mise run routes` - print registered routes and exit.

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

## DB Migrations

- `mise run db-up` - apply pending Bun SQL migrations.
- `mise run db-down` - rollback last Bun migration group.
- `mise run db-status` - show migration status.
- `mise run db-create name=add_new_column` - create new SQL migration files.
