# AGENTS Guide for `bandcash`

This document is for coding agents working in this repository.
Treat it as the default operational guide for build/test/lint and code style.

## Project Snapshot

- Stack: Go 1.26 + Echo + Datastar + templ + SQLite + sqlc + goose.
- Entrypoint: `cmd/server/main.go`.
- Main app areas are in `models/` (`event`, `member`, `group`, `auth`, `settings`, etc.).
- Shared internals are in `internal/` (`db`, `middleware`, `utils`, `i18n`, `email`).
- Templ source files are `*.templ`; generated files are `*_templ.go`.
- Sqlc query files are under `internal/db/queries/*.sql`; generated Go is `internal/db/*.sql.go`.

## Repository Layout (High Value Paths)

- `cmd/server/main.go`: server wiring, middleware chain, route registration, shutdown.
- `cmd/seed/main.go`: local/dev data seed command.
- `models/*/routes.go`: feature route registration and middleware boundaries.
- `models/*/handlers.go`: Echo handlers + Datastar signal parsing and SSE patching.
- `models/*/model.go`: DB access orchestration and page/view data composition.
- `internal/middleware/*.go`: auth, group/admin guards, locale, CSRF, compression.
- `internal/utils/*.go`: rendering, validation, IDs, notifications, JSON/signal helpers.
- `internal/db/migrations`: goose migration files.

## Build / Run / Lint / Test Commands

Prefer `mise` tasks first, then direct Go commands when needed.

### Dev and Run

- `mise run dev` - start with hot reload (`air`).
- `mise run run` - run server directly (`go run ./cmd/server/main.go`).
- `mise run build` - build binary at `tmp/server`.
- `mise run start` - run built binary.
- `mise run routes` - print registered routes and exit.

### Formatting and Static Checks

- `mise run format` - run `go fmt ./...`.
- `mise run vet` - run `go vet ./...`.
- `mise run lint` - run `golangci-lint run`.
- `mise run lsp` - run `gopls check` on tracked Go files, excluding generated DB files.
- `mise run check` - currently broken in this repo (`fmt` task referenced but not defined).
- Recommended replacement for `check`: `mise run format && mise run vet && mise run test`.

### Tests

- `mise run test` - run all tests (`go test -v ./...`).
- Single package: `go test -v ./models/event`.
- Single file (by package + regex): `go test -v ./models/event -run TestHandlerCreate`.
- Single test exact match: `go test -v ./models/event -run '^TestHandlerCreate$'`.
- Multiple tests by pattern: `go test -v ./models/event -run 'TestCreate|TestUpdate'`.
- With race detector: `go test -race ./...`.
- With coverage profile: `go test -coverprofile=coverage.out ./...`.

Notes:
- There are currently no `*_test.go` files in this repo, but the commands above are the expected patterns.
- For sqlc-dependent packages, ensure DB-facing code compiles after query changes.

### Database / Codegen / Templates

- `mise run goose-up` - apply migrations to `sqlite.db`.
- `mise run goose-status` - migration status.
- `mise run goose-create name=add_new_table` - create migration scaffold.
- `mise run sqlc` - regenerate sqlc code from SQL queries.
- `mise run templ` - regenerate templ components.
- `mise run seed` - seed local data.

## Mandatory Generated-File Rules

- Do not manually edit `internal/db/*.sql.go` (sqlc output).
- Do not manually edit `*_templ.go` (templ output).
- Modify sources (`*.sql`, `*.templ`) and regenerate with `mise run sqlc` / `mise run templ`.

## Code Style and Conventions

### Language and Formatting

- Keep code `gofmt`-clean; run `mise run format` before finishing.
- Prefer small functions and early returns for invalid input/error paths.
- Avoid large inline anonymous structs when a named type improves reuse/readability.

### Imports

- Use standard Go grouping: stdlib, third-party, local module (`bandcash/...`) with blank lines.
- Keep import aliases only when needed for clarity/conflict (`ctxi18n`, `echoMiddleware`, etc.).
- Remove unused imports; do not keep speculative imports.

### Types and Data Modeling

- Prefer concrete structs for Datastar signals (`...Signals`, `...Params`) near the handler.
- Use JSON tags matching frontend signal keys (mostly lower camelCase).
- IDs are string-based in routes and DB layer (e.g., `grp_*`, `evt_*`, `mem_*`, `tok_*`).
- Use `int64` for money/amount fields; do not introduce floating point for currency math.
- Pass `context.Context` from request down into model/db calls.

### Naming

- Exported handlers follow Echo conventions with action names like `Index`, `Show`, `Create`, `Update`, `Destroy`.
- Non-REST screens may use explicit page names (`GroupsPage`, `NewGroupPage`, `ViewersPage`).
- Use descriptive variable names: `groupID`, `userEmail`, `memberID`, `signals`, `data`.
- Error field lists are named `<entity>ErrorFields`; default signal maps use `default<Entity>Signals`.

### Handler Patterns

- In view handlers, call `utils.EnsureClientID(c)` before rendering.
- Parse route params/signals first; validate; then perform DB mutation/read.
- Use `utils.ValidateWithLocale` for payload validation and return validation errors via SSE signals.
- On Datastar signal parse failure, return `c.NoContent(400)`.
- For validation failures, return `c.NoContent(422)` and patch `errors` state.
- After successful mutations, notify via `utils.Notify` and patch HTML/signals via `utils.SSEHub`.

### Error Handling

- Follow fail-fast style with early `if err != nil` blocks.
- Log internal errors with `slog` and structured key/value pairs (`"err", err`).
- Return user-safe HTTP responses; avoid exposing raw DB/internal errors.
- For guarded routes, use middleware redirects (`/auth/login`, `/dashboard`) consistent with existing code.

### Logging

- Use `log/slog` consistently.
- Message format convention is `domain.action: detail` (for example: `event.update: failed to render`).
- Include stable keys (`group_id`, `event_id`, `member_id`, `err`), not printf interpolation.

### Routing and Middleware

- Group-scoped routes live under `/groups/:groupId` and use:
  - `middleware.RequireAuth()`
  - `middleware.RequireGroup()`
  - `middleware.RequireAdmin()` for mutations requiring admin access.
- Preserve current middleware chain order in `cmd/server/main.go` unless intentionally changing security behavior.

### Frontend / Templ / Datastar

- Keep signal names and payload shapes aligned between `*.templ` and handler structs.
- Reuse existing page/component composition patterns (`EventIndex`, `EventShow`, `MemberIndex`, etc.).
- Prefer patching via `utils.SSEHub.PatchHTML` + `PatchSignals` rather than ad-hoc response shapes.
- Keep notification usage consistent (`utils.Notify(c, "success"|"error"|"warning", message)`).

### i18n and User-Facing Strings

- Use `ctxi18n.T(ctx, key, args...)` for user-visible text where localized keys already exist.
- Avoid introducing hard-coded user-facing strings in handlers when a translation key pattern is available.

## Operational Expectations for Agents

- Make minimal, targeted changes; avoid broad refactors unless requested.
- Keep handlers thin; put reusable data assembly in model/helper methods.
- Respect existing conventions before introducing new abstractions.
- If migrations or queries change, regenerate sqlc and ensure code compiles.
- If templ files change, regenerate templ output and ensure build/test still pass.

## Cursor / Copilot Rule Audit

Checked paths:
- `.cursor/rules/`
- `.cursorrules`
- `.github/copilot-instructions.md`

Result in this repository:
- No Cursor rules found.
- No Copilot instruction file found.
