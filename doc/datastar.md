# datastar

## What I do

- Provide Datastar attribute conventions and safe defaults for this repo.
- Map common UI behaviors to Datastar attributes.
- Include current CSRF + security expectations for Datastar mutations.

## When to use me

Use this when implementing interactive UI behaviors with Datastar.

## Key references

- Attributes reference: https://data-star.dev/reference/attributes
- Example (active search): https://data-star.dev/examples/active_search

## Quick reference (major attributes)

- `data-signals` - declares reactive state for this element subtree.
  - Example: `data-signals="{ open: false, csrf: '' }"`
- `data-init` - runs an action when the element initializes.
  - Example: `data-init="@get('/sse')"`
- `data-bind` - two-way binds an input/select/textarea to a signal key.
  - Example: `data-bind="formData.name"`
- `data-show` - conditionally renders visibility based on an expression.
  - Example: `data-show="$open"` (use `style="display: none"` to avoid flicker)
- `data-on:<event>` - binds event handlers and can call server actions.
  - Example: `data-on:click="open = !open"`
  - Example: `data-on:submit="@post('/groups')"`
  - Common modifiers: `__debounce`, `__throttle`, `__once`, key filters (for example Enter).
- `data-indicator:<name>` - toggles a loading signal around network actions.
  - Example: `data-indicator:fetching` to expose `$fetching`
- `data-attr:<attrName>` - binds a DOM attribute to an expression.
  - Example: `data-attr:disabled="$fetching"`
- `data-class:<className>` - toggles a CSS class from an expression.
  - Example: `data-class:is-loading="$fetching"`

## Offline quick reference (server actions)

- `@get('/path')` - fetch read-only data/partials.
- `@post('/path')` - create/mutate server state.
- `@put('/path')` - full update.
- `@patch('/path')` - partial update.
- `@delete('/path')` - delete resource.
- Pattern: place actions in `data-on:*` or `data-init`, and include `csrf` in signals for unsafe methods.

## SSE conventions in this repo

- Global stream at `/sse`.
- Pages initialize signals with `data-signals` and connect with `data-init="@get('/sse')"`.
- Server updates UI through `utils.SSEHub` patch/redirect helpers.

## Common patterns

- Toggle UI: `data-signals="{open: false}"` with `data-show="$open"` and `data-on:click`.
- Form submit: `data-on:submit="@post('/path')"`.
- Mutation buttons: `data-on:click="@put('/path')"`, `@delete`, `@post`.
- Bind inputs/signals: `data-bind="name"` on `input`, `select`, `textarea`, and hidden fields.
- CSRF signal: include `csrf` in root `data-signals` on interactive pages.
- Logout action: use Datastar POST (`@post('/auth/logout')`) with bound `csrf` signal.
- Table search action: use `tableSearchAction('/path', $tableQuery, 50)`.
- Loading state: `data-indicator:fetching` + `data-attr:disabled="$fetching"`.
- Multi-step drafts: keep transient edits in signals, and post transition actions (`add/copy/remove`) that return patched UI/signals.

## Server-rendered loop caveat

- If rows/items are rendered by server-side templ loops, mutating signal arrays alone does not create/remove DOM rows.
- Use one of these approaches:
  - server-assisted transition endpoint that patches updated HTML, or
  - explicit client-side repeater rendering for that list.
- Keep persistence separate: write to DB only on final save when possible.

## Example: active search

- Use `data-bind` on the search input.
- Use `data-on:input__debounce` to trigger `@get` with query params.
- Render results with a table or list and keep requests idempotent.

## Notes

- Prefer `data-show` with `class="display: none;"` to avoid flicker.
- Keep signal names camelCase when possible.
- Assume `csrf` signal is sent on unsafe methods and validated server-side.
