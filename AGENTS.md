# Agent Guide for bandcash

This repo is a Go + Echo + Datastar app with SQLite, sqlc, and goose.
Use this file as the default operating guide for agentic changes.

## Quick Orientation
- Entry point: `cmd/server/main.go`.
- Feature modules live in `models/` (event, member, home).
- Shared internals: `internal/` (config, db, logger, middleware, utils, view, hub, sse).
- Templates live in `models/shared/templates/`; static assets in `static/`.

## Build, Lint, and Test Commands
Use `mise` tasks where possible.

### Build and Run
- `mise run dev`        Hot reload server via air.
- `mise run build`      Build binary to `tmp/server`.
- `mise run start`      Run the built binary.
- `mise run run`        Run the server directly.

### Tests
- `mise run test`       Run all tests (`go test -v ./...`).
- Single package: `go test -v ./models/event`.
- Single test: `go test -v ./models/event -run TestName`.
- Single test exact: `go test -v ./models/event -run '^TestName$'`.
- Short mode: `go test -short -v ./...`.

### Format, Lint, Vet
- `mise run fmt`        Format Go code (`go fmt ./...`).
- `mise run vet`        Run `go vet`.
- `mise run lint`       Run `golangci-lint`.
- `mise run lsp`        Run `gopls check` (excludes generated sqlc files).
- `mise run check`      Run fmt + vet + test.

### Database and Codegen
- `mise run goose-up`       Run migrations.
- `mise run goose-status`   Show migration status.
- `mise run goose-create name=add_new_column`  Create migration.
- `mise run sqlc`           Regenerate sqlc code.
- `mise run seed`           Seed database.

## Code Style Guidelines
Follow existing patterns from `models/`, `internal/`, and `cmd/`.

### Formatting and Imports
- Run `gofmt` (via `mise run fmt`) before final changes.
- Imports are grouped: stdlib, third-party, local (`bandcash/...`) with blank lines.
- Use import aliases when common (`appmw` for `internal/middleware`).

### Types and Naming
- REST handler methods use Rails-style names: `Index`, `New`, `Show`, `Edit`, `Create`, `Update`, `Destroy`.
- Use data structs like `EventData`, `EventsData`, `MemberData` for template rendering.
- Use `int` for route params, convert to `int64` for database calls.
- Prefer explicit names over abbreviations; `memberID`, `eventID`, `clientID`.

### Error Handling and Responses
- Handle errors early; log with context and return user-safe responses.
- Pattern for handlers:
  - Parse/validate inputs.
  - Call model/db helpers.
  - Log failures: `log.Error("area.action: message", "err", err)`.
  - Return `c.String(500, "Internal Server Error")` on server errors.
  - Use `c.NoContent(400)` for bad signals or invalid payloads.
  - Use `c.Redirect(303, ...)` after create/update/delete.
- Use `utils.ParseRawInt64` when signals arrive as `json.RawMessage`.

### Logging
- Use request logger from middleware: `log := appmw.Logger(c)`.
- Use `slog` directly for app startup/shutdown and non-request contexts.
- Log keys are short, stable, and structured (no printf strings).

### HTTP and Handlers
- Call `utils.EnsureClientID(c)` early in view handlers.
- Use `datastar.ReadSignals` for form submissions; keep signal structs near handlers.
- Keep route parsing and validations in handler; DB work in model methods.

### Templates and UI
- Templates follow `index/new/show/edit` naming.
- Use the shared breadcrumbs partial `models/shared/templates/breadcrumbs.html`.
- View data includes `Breadcrumbs []utils.Crumb` when applicable.

### Datastar + SSE
- Single SSE stream at `/sse` and a `view` signal drives rendering.
- After mutations, use `hub.Hub.Render(clientID)` and `hub.Hub.PatchSignals`.
- Signal names are lower camelCase and match template usage.

### Database and sqlc
- SQL queries live in `internal/db/queries/*.sql`.
- Generated code is in `internal/db/*.sql.go` and should not be edited manually.
- After changing queries or migrations, run `mise run sqlc`.
- Migrations live in `internal/db/migrations/` and are managed by goose.
- The server does not auto-run migrations; use `mise run goose-up` explicitly.

## Conventions by Example
- Parsing params: `id, err := strconv.Atoi(c.Param("id"))`.
- On invalid ID: `return c.String(400, "Invalid ID")`.
- On signal read error: `return c.NoContent(400)` with warning log.
- On success: `return c.NoContent(200)` or redirect.

## Tooling Notes
- Go version is specified in `go.mod` (currently 1.24.0).
- Mise tasks are defined in `mise.toml`.
- `golangci-lint` is available via `mise`.

## Cursor / Copilot Rules
- No Cursor rules found (`.cursor/rules/` or `.cursorrules` not present).
- No Copilot instructions found (`.github/copilot-instructions.md` not present).

## When Adding New Code
- Match handler and model patterns in `models/event` and `models/member`.
- Keep handlers thin; prefer `GetXData` or `CreateX` helpers in the model.
- Ensure signal names are consistent across handlers and templates.
- Update breadcrumbs for new pages and include `utils.Crumb` data.

## Things to Avoid
- Do not edit generated sqlc files in `internal/db/*.sql.go`.
- Avoid adding new logging styles; stick to `slog` + `appmw.Logger`.
- Avoid changing existing routes without updating SSE view handling.
