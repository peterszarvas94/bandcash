---
name: datastar
description: Datastar attribute usage and patterns, based on official docs and examples.
---

## What I do

- Provide Datastar attribute conventions and safe defaults for this repo.
- Map common UI behaviors to Datastar attributes.
- Include current CSRF + security expectations for Datastar mutations.

## When to use me

Use this when implementing interactive UI behaviors with Datastar.

## Key references

- Attributes reference: https://data-star.dev/reference/attributes
- Example (active search): https://data-star.dev/examples/active_search

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
- Loading state: `data-indicator:fetching` + `data-attr:disabled="$fetching"`.

## Example: active search

- Use `data-bind` on the search input.
- Use `data-on:input__debounce` to trigger `@get` with query params.
- Render results with a table or list and keep requests idempotent.

## Notes

- Prefer `data-show` with `class="display: none;"` to avoid flicker.
- Keep signal names camelCase when possible.
- Assume `csrf` signal is sent on unsafe methods and validated server-side.
