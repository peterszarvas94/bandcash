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
- Routes and handlers: `models/**/routes.go`, `models/**/handlers.go`
- View/data structs: `models/**/view.go`, `models/**/model.go`, `models/**/signals.go`
- SSE + helpers: `internal/utils/hub.go`, `models/sse/**`
- Middleware: `internal/middleware/*.go`
- Templates: `models/**/*.templ` (+ generated `*_templ.go`)
- Database: `internal/db/queries/*.sql`, `internal/db/migrations/*.sql`, generated `internal/db/*.sql.go`
- Assets: `static/css/*.css`, `static/js/*.js`

## Scoped search rules
- Start with `models/<area>/*.templ` when UI is mentioned.
- Use `models/<area>/handlers.go` for request logic.
- Use `internal/db/queries/*.sql` for query changes, then run sqlc.
- Use `Glob` with narrow patterns before `Grep`.

## Notes
- Avoid repo-wide `Grep` unless the feature is unknown.
- Prefer reading a small set of files over broad scans.
