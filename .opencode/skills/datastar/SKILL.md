---
name: datastar
description: Datastar attribute usage and patterns, based on official docs and examples.
---

## What I do
- Provide Datastar attribute conventions and safe defaults for this repo.
- Map common UI behaviors to Datastar attributes.
- Point to official references when needed.

## When to use me
Use this when implementing interactive UI behaviors with Datastar.

## Key references
- Attributes reference: https://data-star.dev/reference/attributes
- Example (active search): https://data-star.dev/examples/active_search

## SSE conventions in this repo
- Global stream at `/sse`.
- Client sends `view` signal matching the route (e.g. `/entry/1`).
- Server patches `main#app` on updates.

## Common patterns
- Toggle UI: `data-signals="{open: false}"` with `data-show="$open"` and `data-on:click`.
- Form submit: `data-on:submit="@post('/path')"` to prevent default and send signals.
- SSE connect: `data-signals="{view: '/entry/1'}" data-init="@get('/sse')"`.
- Bind inputs: `data-bind="name"` on `input`, `select`, `textarea`.
- Loading state: `data-indicator:fetching` + `data-attr:disabled="$fetching"`.

## Example: active search
- Use `data-bind` on the search input.
- Use `data-on:input__debounce` to trigger `@get` with query params.
- Render results with a table or list and keep requests idempotent.

## Notes
- Prefer `data-show` with `style="display: none;"` to avoid flicker.
- Keep signal names camelCase when possible.
