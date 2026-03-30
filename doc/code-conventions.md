# Code Conventions

Keep edits minimal, idiomatic, and aligned with existing module patterns.

## Formatting and Structure

- Keep all Go code `gofmt` clean.
- Prefer small functions with early returns for invalid input and error paths.
- Keep handlers thin; move data assembly/queries into model methods where practical.
- Avoid broad refactors unless explicitly requested.

## Imports

- Group imports as: stdlib, third-party, local module (`bandcash/...`) separated by blank lines.
- Use aliases only when they improve clarity or avoid naming collisions.
- Remove unused imports.

## Types and Data Modeling

- Use explicit structs for Datastar payloads (for example `createGroupSignals`, `...Params`).
- Keep JSON tags aligned with frontend signal keys (usually lower camelCase).
- Pass `context.Context` from request to model and DB calls.
- Keep IDs as strings with existing prefixes (`grp_`, `evt_`, `mem_`, `tok_`).
- Use `int64` for currency and totals; do not introduce floating-point money math.

## Naming

- Follow Echo-style handler names: `Index`, `Show`, `Create`, `Update`, `Destroy`.
- Use explicit page names for non-REST screens (`GroupsPage`, `ViewersPage`, etc.).
- Prefer descriptive names like `groupID`, `memberID`, `signals`, `data`, `tableQuery`.
- Follow existing error/signal naming patterns (`<entity>ErrorFields`, `default<Entity>Signals`).

## Error Handling and Logging

- Use fail-fast `if err != nil` checks close to the source.
- Log internal errors with `slog` structured fields, especially `"err", err`.
- Keep logs consistent with `domain.action: detail` style messages.
- Include stable IDs in logs when available (`group_id`, `event_id`, `member_id`).
- Return safe HTTP responses; avoid leaking internal DB details.

## Routing and Middleware

- Group-scoped routes live under `/groups/:groupId`.
- Preserve expected middleware boundaries:
  - `middleware.RequireAuth()`
  - `middleware.RequireGroup()`
  - `middleware.RequireAdmin()` where mutation privileges require admin access.
- Do not reorder global middleware in `cmd/server/main.go` unless intentionally changing security behavior.

## i18n

- Use `ctxi18n.T(ctx, key, args...)` for user-facing strings whenever keys exist.
- Avoid hard-coded UI strings where localized keys are appropriate.

## Practical Defaults

- Prefer early returns.
- Prefer small focused changes.
- Keep business logic out of handlers when practical.
