# Handler Patterns

Use this for request flow, Datastar signal handling, and SSE response behavior.

## Render and Request Flow

- In render handlers, call `utils.EnsureClientID(c)` before page rendering.
- Parse params/signals first, validate next, then mutate/read from DB.
- On signal parse failure, return `c.NoContent(http.StatusBadRequest)`.
- For validation failures, return `c.NoContent(http.StatusUnprocessableEntity)` and patch errors.
- Use `utils.ValidateWithLocale` for localized validation errors.

## Datastar + SSE

- Keep handler signal structs and templ signal usage in sync.
- Prefer `utils.SSEHub.PatchHTML` and `PatchSignals` for partial updates.
- Use `utils.Notify` for user-visible success/error/warning notifications.

## Multi-step and Dynamic UI

- For draft editors, prefer client-owned transient state (`wizard.*`) with server-assisted transitions.
- If rows are rendered by server templ loops, do not rely only on signal-array mutation for visible add/remove; trigger server transition endpoint or use explicit client-side repeater.

## Frontend Conventions

- Reuse shared table/query infrastructure (`models/shared/table.templ`, `internal/utils/table_query.go`, `static/js/table_query.js`).
- Prefer utility classes for layout/spacing; keep semantic component classes for reusable UI objects.
- Avoid one-off CSS wrappers for simple display/alignment rules.

## Status Codes

- `400` for malformed signal/body input.
- `422` for validation errors.
- `200/204` for successful patch/mutation flows.
