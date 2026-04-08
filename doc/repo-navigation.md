# repo-navigation

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
- Group payment tabs (split pages): `doc/group-payments.md`, `models/group/page_{to_pay,to_receive,recent_income,recent_outgoing}.templ`, `models/group/component_{to_pay_main,to_receive_main,recent_income_main,recent_outgoing_main}.templ`
- Shared tables: `models/shared/table.templ`, `internal/utils/table_query.go`, `static/js/table_query.js`
- Database: `internal/db/bunmigrations/*.sql`, `internal/db/*.go`
- Assets: `static/css/*.css`, `static/js/*.js`
- Deploy config: `config/deploy.yml`, `.kamal/secrets.example`, `.kamal/secrets`

## Scoped search rules
- Start with `models/<area>/*.templ` when UI is mentioned.
- Use `models/<area>/handlers.go` for request logic.
- Use `models/shared/table.templ` + `internal/utils/table_query.go` for table search/sort/pagination.
- Use `internal/db/bunmigrations/*.sql` for schema changes.
- Use `Glob` with narrow patterns before `Grep`.

## Notes
- Avoid repo-wide `Grep` unless the feature is unknown.
- Prefer reading a small set of files over broad scans.
- Generated files are read-only targets: `*_templ.go`.
