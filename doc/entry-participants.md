# entry-participants

## What I do
- Point to event detail templates and participant table markup.
- Describe where to add participant form actions in handlers.
- Keep participant UI consistent with existing forms and tables.

## When to use me
Use this when adding participant UI or wiring add/remove participant actions.

## Key paths
- Event detail template: `models/event/page_show.templ` and `models/event/component_show_*.templ`
- Event handlers: `models/event/handlers_pages.go`, `models/event/handlers_actions.go`
- Participant data access: `internal/db/bun_queries.go`, `internal/db/bun_api.go`
- SSE hub: `internal/utils/hub.go`

## UI conventions
- Tables use `.table` and numeric values use `.text-right`.
- Forms use `class="form"` and `data-on:submit`.
- Number inputs use `type="number"` and `step="0.01"`.
- Keep signal structs in `signals.go` aligned with `data-bind` names in templ.

## SSE update pattern
- Use `utils.SSEHub.PatchHTML(c, html)` to patch full view markup.
- Use `utils.SSEHub.PatchSignals(c, ...)` for form/UI state updates.
- Use `utils.SSEHub.Redirect(c, url)` after successful mutations.

## Draft-state pattern (preferred)
- Keep in-progress participant edits in client signals (`wizard.*`) until final save.
- For add/copy/remove row actions, treat handlers as transition endpoints: read current signals + action, compute next `wizard.rows`, then patch HTML/signals.
- Persist participant changes only on final save submit, after full validation.
- If table rows are server-rendered via templ loop, do not expect signal-only array updates to add/remove visible rows without a server patch.

## Handler pattern
- Parse signals first; on parse failure return `c.NoContent(http.StatusBadRequest)`.
- Validate with `utils.ValidateWithLocale`; return `c.NoContent(http.StatusUnprocessableEntity)` for validation failures.
- Keep participant mutation handlers thin; move query/data assembly into model methods when practical.
