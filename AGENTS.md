# AGENTS Guide for `bandcash`

This is the default operating guide for coding agents in this repo.
Use it for command selection, code style, and safe edit boundaries.

## Project Snapshot

- Stack: Go 1.26, Echo, Datastar, templ, SQLite, sqlc, goose.
- Main entrypoint: `cmd/server/main.go`.
- Feature modules live in `models/` (`group`, `event`, `member`, `expense`, `auth`, etc.).
- Shared internals are in `internal/` (`db`, `utils`, `middleware`, `i18n`, `email`).
- UI sources are `*.templ`; generated templ files are `*_templ.go`.
- SQL query sources are `internal/db/queries/*.sql`; generated sqlc code is `internal/db/*.sql.go`.

## High-Value Paths

- `cmd/server/main.go`: middleware chain, route registration, startup and shutdown.
- `cmd/seed/main.go`: dev seed command.
- `models/*/routes.go`: route groups and auth/admin boundaries.
- `models/*/handlers.go`: request handling, signal parsing, SSE patching.
- `models/*/model.go`: page data composition and DB orchestration.
- `models/shared/table.templ`: shared table rendering helpers.
- `internal/utils/table_query.go`: shared sort/search/pagination query utilities.
- `internal/middleware/*.go`: auth, CSRF, locale, security middleware.
- `internal/db/migrations/*.sql`: goose migrations.

## Build, Run, Lint, Test

Prefer `mise run <task>` commands first.

### Run and Dev

- `mise run dev` - start app with hot reload (`air`) plus local mail service.
- `mise run run` - run server directly with `go run`.
- `mise run build` - compile server binary to `tmp/server`.
- `mise run start` - run built binary.
- `mise run routes` - print registered routes and exit.

### Formatting and Static Checks

- `mise run format` - run `go fmt ./...`.
- `mise run vet` - run `go vet ./...`.
- `mise run lint` - run `golangci-lint run`.
- `mise run lsp` - run `gopls check` against tracked Go files (excluding generated DB files).
- `mise run check` - run format + vet + test in sequence.

### Tests (including single-test workflows)

- `mise run test` - run all tests (`go test ./...`).
- `mise run test-pretty` - run tests with gotestsum output.
- Single package: `go test ./models/event`.
- Single test by name: `go test ./models/event -run TestHandlerCreate`.
- Single test exact match: `go test ./models/event -run '^TestHandlerCreate$'`.
- Multiple tests by regex: `go test ./models/event -run 'TestCreate|TestUpdate'`.
- Test one package with race detector: `go test -race ./models/event`.
- Full race run: `go test -race ./...`.
- Coverage profile: `go test -coverprofile=coverage.out ./...`.

Tips:

- `-run` filters by test function name regex within the selected package.
- Go does not directly run tests by file path; choose package + `-run` pattern.
- After DB query changes, run tests for affected packages and ensure compile passes.

### Database and Codegen

- `mise run goose-up` - apply migrations to local `sqlite.db`.
- `mise run goose-status` - show migration status.
- `mise run goose-create name=add_new_table` - create migration scaffold.
- `mise run sqlc` - regenerate sqlc outputs after SQL query changes.
- `mise run seed` - seed local DB data.

## Generated File Rules (Mandatory)

- Do not hand-edit `internal/db/*.sql.go` (sqlc generated).
- Do not hand-edit `*_templ.go` (templ generated).
- Edit source files (`*.sql`, `*.templ`) and regenerate only when needed.
- Do not run `mise run templ` unless explicitly requested; normal dev flow handles templ regeneration.

## Code Style and Conventions

### Formatting and Structure

- Keep all Go code `gofmt` clean.
- Prefer small functions with early returns for invalid input and error paths.
- Keep handlers thin; move data assembly/queries into model methods where practical.
- Avoid broad refactors unless the task explicitly asks for them.

### Imports

- Group imports as: stdlib, third-party, local module (`bandcash/...`) separated by blank lines.
- Use aliases only when they improve clarity or avoid naming collisions.
- Remove unused imports and avoid speculative imports.

### Types and Data Modeling

- Use explicit structs for Datastar payloads (for example `createGroupSignals`, `...Params`).
- Keep JSON tags aligned with frontend signal keys (usually lower camelCase).
- Pass `context.Context` from request to model and DB calls.
- Keep IDs as strings with existing prefixes (`grp_`, `evt_`, `mem_`, `tok_`).
- Use `int64` for currency and totals; do not introduce floating-point money math.

### Naming

- Follow Echo-style handler names: `Index`, `Show`, `Create`, `Update`, `Destroy`.
- Use explicit page names for non-REST screens (`GroupsPage`, `ViewersPage`, etc.).
- Prefer descriptive names like `groupID`, `memberID`, `signals`, `data`, `tableQuery`.
- Follow existing error/signal naming patterns (`<entity>ErrorFields`, `default<Entity>Signals`).

### Handler and Datastar Patterns

- In render handlers, call `utils.EnsureClientID(c)` before page rendering.
- Parse params/signals first, validate next, then mutate/read from DB.
- On signal parse failure, return `c.NoContent(http.StatusBadRequest)`.
- For validation failures, return `c.NoContent(http.StatusUnprocessableEntity)` and patch errors.
- Use `utils.ValidateWithLocale` for localized validation errors.
- Prefer `utils.SSEHub.PatchHTML` and `PatchSignals` for partial updates.
- Use `utils.Notify` for user-visible success/error/warning notifications.

### Error Handling and Logging

- Use fail-fast `if err != nil` checks close to the source.
- Log internal errors with `slog` structured fields, especially `"err", err`.
- Keep logs consistent with `domain.action: detail` style messages.
- Include stable IDs in logs when available (`group_id`, `event_id`, `member_id`).
- Return safe HTTP responses; avoid leaking internal DB details.

### Routing and Middleware

- Group-scoped routes live under `/groups/:groupId`.
- Preserve expected middleware boundaries:
  - `middleware.RequireAuth()`
  - `middleware.RequireGroup()`
  - `middleware.RequireAdmin()` where mutation privileges require admin access.
- Do not reorder global middleware in `cmd/server/main.go` unless intentionally changing security behavior.

### Frontend (templ/CSS/JS)

- Keep handler signal structs and templ signal usage in sync.
- Reuse shared table/query infrastructure (`models/shared/table.templ`, `internal/utils/table_query.go`, `static/js/table_query.js`).
- Prefer utility classes for layout/spacing and keep semantic component classes for reusable UI objects.
- Avoid one-off CSS wrappers for simple display/alignment rules.
- Keep table interactions and SSE updates consistent with existing patterns.

### i18n

- Use `ctxi18n.T(ctx, key, args...)` for user-facing strings whenever keys exist.
- Avoid hard-coded UI strings in handlers where localized keys are appropriate.

## Agent Operating Expectations

- Make minimal, targeted edits; preserve established architecture.
- Do not revert unrelated dirty-worktree changes.
- If SQL changes, regenerate sqlc and verify affected packages compile/test.
- If only templ source changes are requested, avoid manually editing generated templ Go files.

## Cursor and Copilot Rules

Checked rule locations:

- `.cursor/rules/`
- `.cursorrules`
- `.github/copilot-instructions.md`

Current repository state:

- No Cursor rules found.
- No Copilot instruction file found.

If these files are added later, treat them as higher-priority repository instructions and update this guide.
