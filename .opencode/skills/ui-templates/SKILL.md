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
- Feature templates: `models/**/**/*.templ`
- Shared layout/components: `models/shared/*.templ` and `models/shared/*.go`
- Shared table components: `models/shared/table.templ`
- Static assets: `static/css/*`, `static/js/*`

## Form conventions
- Forms use `class="form"` and `data-on:submit` handlers.
- Inputs use `data-bind` for Datastar signals.
- Numeric inputs use `type="number"` and `step="0.01"`.
- Interactive pages include root `data-signals` + `data-init="@get('/sse')"`.
- Unsafe actions must have `csrf` present in root signals.
- Logout is POST-only (`/auth/logout`) and should be triggered by Datastar form/button.

## Table conventions
- Use `.table` class for tables.
- Align numeric values with `.text-right`.
- Prefer `shared.TableSearchForm(...)` and `shared.TableSortHeader(...)`.
- Build table actions with `utils.TableSearchAction(...)`.

## Notes
- Keep markup in existing sections; do not introduce new layout wrappers unless needed.
- Run `mise run templ` after template updates.
