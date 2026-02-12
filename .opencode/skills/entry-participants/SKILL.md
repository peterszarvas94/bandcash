---
name: entry-participants
description: UI and handler conventions for entry participants (tables, add/remove).
---

## What I do
- Point to entry detail templates and participant table markup.
- Describe where to add participant form actions in handlers.
- Keep participant UI consistent with existing forms and tables.

## When to use me
Use this when adding participant UI or wiring add/remove participant actions.

## Key paths
- Entry detail template: `app/entry/templates/show.html`
- Entry handlers: `app/entry/handlers.go`
- Participant queries: `internal/db/queries/participants.sql`
- SSE hub: `internal/hub/store.go`

## UI conventions
- Tables use `.table` and numeric values use `.text-right`.
- Forms use `class="form"` and `data-on:submit`.
- Number inputs use `type="number"` and `step="0.01"`.

## SSE update pattern
- Use `hub.Hub.Render(clientID)` to trigger re-render.
- Optionally `hub.Hub.PatchSignals(clientID, ...)` for UI state.
