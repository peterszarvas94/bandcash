-- name: GetAppFlagBool :one
SELECT bool_value
FROM app_flags
WHERE key = ?;

-- name: UpsertAppFlagBool :exec
INSERT INTO app_flags (key, bool_value)
VALUES (?, ?)
ON CONFLICT(key) DO UPDATE SET
  bool_value = excluded.bool_value,
  updated_at = CURRENT_TIMESTAMP;
