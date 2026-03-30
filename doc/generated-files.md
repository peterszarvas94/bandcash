# Generated Files

Use this as a strict guardrail before editing generated outputs.

## Mandatory Rules

- Do not hand-edit `internal/db/*.sql.go` (sqlc generated).
- Do not hand-edit `*_templ.go` (templ generated).
- Edit source files (`*.sql`, `*.templ`) and regenerate outputs.
- Do not run `mise run templ` unless explicitly requested; normal dev flow handles templ regeneration.

## Related Workflows

- SQL changes: follow `doc/sqlc-goose.md`.
- Template/UI changes: follow `doc/ui-templates.md`.

## Rule of Thumb

Edit the source, regenerate, then verify compile/tests.
