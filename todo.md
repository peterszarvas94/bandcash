# UI Refactor Todo

Goal: Use the same two-column layout pattern as `/entry` index (table/content on left, form panel on right) and avoid standalone create/edit pages.

## Forms to refactor

### Entry
- `models/entry/templates/show.html`
  - Use two-column layout: entry details + participants on left, entry edit + participant form on right.
  - Remove standalone `/entry/:id/edit` usage in UI.
- `models/entry/templates/new.html`
- `models/entry/templates/edit.html`
  - Eliminate standalone routes in UI. Keep for now or remove if not needed.

### Payee
- `models/payee/templates/index.html`
  - Convert to two-column layout like `/entry` index.
  - Place payee form in right panel; table on left.
- `models/payee/templates/show.html`
  - Two-column layout: payee details + entries on left, edit form on right.
- `models/payee/templates/new.html`
- `models/payee/templates/edit.html`
  - Eliminate standalone routes in UI. Keep for now or remove if not needed.

## Follow-ups
- Update routes/handlers to avoid linking to standalone edit/new pages.
- Confirm which routes can be removed entirely vs. left for direct URL access.
