# Logs

Bandcash uses `log/slog` with two outputs at the same time:

- colored human-readable logs to stdout (good for local development)
- JSON logs to a timestamped file (good for searching and parsing)

## Where Logs Go

- Default folder: `logs/`
- Default prefix: `bandcash`
- Final file format: `<prefix>_<YYYY_MM_DD_HHMMSS>.log`
- Example: `logs/bandcash_2026_04_06_093010.log`

At startup, the app prints the active log file path:

```text
writing to log file: logs/bandcash_2026_04_06_093010.log
```

## Environment Variables

- `LOG_LEVEL`: `debug`, `info`, `warn`, `error` (default from env config: `debug`)
- `LOG_FOLDER`: folder where log files are written (default: `logs`)
- `LOG_PREFIX`: base name for log files (default: `bandcash`)

If `LOG_LEVEL` is invalid in runtime logger setup, the logger falls back to `info`.

## Request Logging

Every HTTP request emits an `info` log from `internal/middleware/request_logger.go`:

- message: `http.request.completed`
- fields: `path`, `query`, `method`, `status`

`query` is logged as a decoded JSON-like object (not URL-encoded raw text). For `datastar`, the middleware parses the JSON payload and redacts sensitive keys recursively (`csrf`, `token`, `password`, `secret`, `authorization`, `cookie`).

Example JSON line:

```json
{"time":"2026-04-06T09:31:12.123456+02:00","level":"INFO","message":"http.request.completed","path":"/sse","query":{"datastar":{"mode":"single","csrf":"[REDACTED]","tab_id":"tab_123"}},"method":"GET","status":200}
```

## Log Message Style

Follow the existing style used across handlers:

- message format: `domain.action: detail`
- always include `"err", err` for errors
- include stable IDs when available (`group_id`, `event_id`, `member_id`, `user_id`)

Examples:

- `event.create.table: failed to create event`
- `group.users_edit_page: failed to get group`
- `participant.bulk: failed to commit tx`

## Quick Usage

- Run app: `mise run dev` (or `mise run run`)
- Watch latest file: `tail -f logs/*.log`
- Filter errors in JSON logs: `rg '"level":"ERROR"' logs/`

## Parsing JSON Logs With jq

Log files in `logs/*.log` are newline-delimited JSON, so `jq` works well for readable filtering.

Always check the latest timestamped log file first (the app creates a new file on each start):

- `latest_log="$(ls -1t logs/*.log | head -n 1)"`
- `jq . "$latest_log"`

- Pretty print all lines: `jq . logs/bandcash_2026_04_06_093010.log`
- Show request logs only: `jq 'select(.message == "http.request.completed")' logs/*.log`
- Show path + status + datastar mode: `jq 'select(.message == "http.request.completed") | {time, path, status, mode: .query.datastar.mode}' logs/*.log`
- Show errors only: `jq 'select(.level == "ERROR")' logs/*.log`

## Related Files

- `internal/utils/logger.go` - logger setup, console/file handlers, level parsing
- `internal/middleware/request_logger.go` - per-request structured log entries
- `internal/utils/env.go` - log-related env definitions and validation
