---
name: repo-navigation
description: Quick path map and scoped search rules for this repo.
---

## What I do
- Provide a fast map of key directories and file types.
- Encourage scoped searches instead of repo-wide scans.
- Point to the correct layer (routes, handlers, templates, db, assets).

## When to use me
Use this when you need to locate where to make changes without scanning the whole repo.

## Key paths
- Routes and handlers: `app/**/routes.go`, `app/**/handlers.go`
- Models and data access: `app/**/model.go`, `internal/db/**`
- SSE hub: `internal/hub/store.go`, `internal/sse/sse.go`
- HTML templates: `app/**/templates/*.html`
- Shared templates: `web/templates/*.html`
- CSS: `web/static/css/*.css`
- JS: `web/static/js/*.js`

## Scoped search rules
- Start with `app/<area>/templates/*.html` when UI is mentioned.
- Use `app/<area>/handlers.go` for request logic.
- Use `internal/db/queries/*.sql` for query changes, then run sqlc.
- Use `Glob` with narrow patterns before `Grep`.

## Notes
- Avoid repo-wide `Grep` unless the feature is unknown.
- Prefer reading a small set of files over broad scans.
