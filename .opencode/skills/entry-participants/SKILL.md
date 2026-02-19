---
name: entry-participants
description: UI and handler conventions for entry participants (tables, add/remove).
---

## What I do
- Point to event detail templates and participant table markup.
- Describe where to add participant form actions in handlers.
- Keep participant UI consistent with existing forms and tables.

## When to use me
Use this when adding participant UI or wiring add/remove participant actions.

## Key paths
- Event detail template: `models/event/page_show.templ` and `models/event/component_show_*.templ`
- Event handlers: `models/event/handlers.go`
- Participant queries: `internal/db/queries/participants.sql`
- SSE hub: `internal/utils/hub.go`

## UI conventions
- Tables use `.table` and numeric values use `.text-right`.
- Forms use `class="form"` and `data-on:submit`.
- Number inputs use `type="number"` and `step="0.01"`.

## SSE update pattern
- Use `utils.SSEHub.PatchHTML(c, html)` to patch full view markup.
- Use `utils.SSEHub.PatchSignals(c, ...)` for form/UI state updates.
- Use `utils.SSEHub.Redirect(c, url)` after successful mutations.
