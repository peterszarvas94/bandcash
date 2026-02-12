---
name: ui-templates
description: HTML template and form conventions for UI changes.
---

## What I do
- Point to template locations and shared head/layout usage.
- Capture form and table conventions used in templates.
- Keep UI edits minimal and consistent.

## When to use me
Use this for page layout, buttons, tables, and form changes.

## Template locations
- Entry pages: `app/entry/templates/*.html`
- Payee pages: `app/payee/templates/*.html`
- Shared head: `web/templates/head.html`

## Form conventions
- Forms use `class="form"` and `data-on:submit` handlers.
- Inputs use `data-bind` for Datastar signals.
- Numeric inputs use `type="number"` and `step="0.01"`.
- SSE views use `data-signals` + `data-init="@get('/sse')"`.

## Table conventions
- Use `.table` class for tables.
- Align numeric values with `.text-right`.

## Notes
- Keep markup in existing sections; do not introduce new layout wrappers unless needed.
