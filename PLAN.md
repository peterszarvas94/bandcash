# DB Hard-Cutover Plan (No sqlc, No Legacy Query Layer)

## Goal

- Keep `internal/db` minimal and Bun-first.
- Remove sqlc-era runtime API entirely.
- Keep API surface small and consistent.
- Use Bun query builder (no embedded SQL query catalogs in Go).

## Current Context

- `Resource` + empty struct table abstractions are already removed from Bun table query files.
- `internal/db/bun_queries.go` was removed.
- `internal/db/query_compat.go` currently keeps legacy call sites compiling.
- Legacy layer still exists (`internal/db/db.go`, `internal/db/*.sql.go`) and is still referenced widely.

## Migration Steps

1. Replace transaction usage built on `db.New(tx)` with Bun-native transaction helpers.
2. Implement Bun-native auth/settings/session/flags APIs with minimal signatures.
3. Replace admin list/count endpoints with allowlisted generic table query APIs.
4. Replace member-participant sorted query fan-out (`ListParticipantsByMemberBy*`) with one Bun table API (allowlisted sort).
5. Replace group access/invite/readers/admin helpers with Bun-native APIs.
6. Remove compatibility/legacy files:
   - `internal/db/query_compat.go`
   - `internal/db/db.go`
   - `internal/db/*.sql.go`
7. Update integration tests to Bun-native DB APIs and run full test suite.

## Constraints

- Keep code minimal and avoid unnecessary wrappers.
- Prefer primitive args or small focused params structs.
- Keep sorting/filtering strictly allowlisted.
- Keep all DB operations in `internal/db` package.

## Done Criteria

- No `db.New(...)` usage remains.
- No references to `type Queries` or sqlc-style `*FilteredParams/*FilteredRow` contracts from app layers.
- No `internal/db/*.sql.go` or `internal/db/db.go` in repository.
- `go test ./...` passes.
