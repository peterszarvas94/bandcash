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
- Shared layout/components: `models/shared/*.templ`
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
- Do not edit generated `*_templ.go` files by hand.
- Do not run `mise run templ` unless explicitly requested; normal dev flow handles templ regeneration.

## CSS conventions (utility-first)
- Prefer utilities for layout/spacing (`row`, `grid`, `pt`, `pb`, `show-mobile`, `hide-mobile`, etc.).
- Keep semantic component classes for real reusable objects (buttons, inputs, dialogs, notifications, app shell).
- Avoid one-off classes that only provide basic `display`, `gap`, `align-items`, or `flex-wrap`.
- Use `data-show` for stateful visibility and utility classes for responsive visibility.
- Prefer padding/utility spacing over margins for layout rhythm.
- Restrict `gap` usage to `1x` (`var(--space)`) or `0.5x` (`calc(var(--space) * 0.5)`).
