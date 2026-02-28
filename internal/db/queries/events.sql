-- name: CreateEvent :one
INSERT INTO events (id, group_id, title, time, description, amount)
VALUES (?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: GetEvent :one
SELECT * FROM events
WHERE id = ? AND group_id = ?;

-- name: ListEvents :many
SELECT * FROM events
WHERE group_id = ?
ORDER BY time ASC;

-- name: CountEventsFiltered :one
SELECT COUNT(*) FROM events
WHERE group_id = sqlc.arg(group_id)
  AND (
    sqlc.arg(search) = ''
    OR title LIKE '%' || sqlc.arg(search) || '%'
    OR description LIKE '%' || sqlc.arg(search) || '%'
  );

-- name: ListEventsByTimeAscFiltered :many
SELECT * FROM events
WHERE group_id = sqlc.arg(group_id)
  AND (
    sqlc.arg(search) = ''
    OR title LIKE '%' || sqlc.arg(search) || '%'
    OR description LIKE '%' || sqlc.arg(search) || '%'
  )
ORDER BY time ASC, created_at DESC
LIMIT sqlc.arg(limit) OFFSET sqlc.arg(offset);

-- name: ListEventsByTimeDescFiltered :many
SELECT * FROM events
WHERE group_id = sqlc.arg(group_id)
  AND (
    sqlc.arg(search) = ''
    OR title LIKE '%' || sqlc.arg(search) || '%'
    OR description LIKE '%' || sqlc.arg(search) || '%'
  )
ORDER BY time DESC, created_at DESC
LIMIT sqlc.arg(limit) OFFSET sqlc.arg(offset);

-- name: ListEventsByTitleAscFiltered :many
SELECT * FROM events
WHERE group_id = sqlc.arg(group_id)
  AND (
    sqlc.arg(search) = ''
    OR title LIKE '%' || sqlc.arg(search) || '%'
    OR description LIKE '%' || sqlc.arg(search) || '%'
  )
ORDER BY title COLLATE NOCASE ASC, created_at DESC
LIMIT sqlc.arg(limit) OFFSET sqlc.arg(offset);

-- name: ListEventsByTitleDescFiltered :many
SELECT * FROM events
WHERE group_id = sqlc.arg(group_id)
  AND (
    sqlc.arg(search) = ''
    OR title LIKE '%' || sqlc.arg(search) || '%'
    OR description LIKE '%' || sqlc.arg(search) || '%'
  )
ORDER BY title COLLATE NOCASE DESC, created_at DESC
LIMIT sqlc.arg(limit) OFFSET sqlc.arg(offset);

-- name: ListEventsByAmountAscFiltered :many
SELECT * FROM events
WHERE group_id = sqlc.arg(group_id)
  AND (
    sqlc.arg(search) = ''
    OR title LIKE '%' || sqlc.arg(search) || '%'
    OR description LIKE '%' || sqlc.arg(search) || '%'
  )
ORDER BY amount ASC, created_at DESC
LIMIT sqlc.arg(limit) OFFSET sqlc.arg(offset);

-- name: ListEventsByAmountDescFiltered :many
SELECT * FROM events
WHERE group_id = sqlc.arg(group_id)
  AND (
    sqlc.arg(search) = ''
    OR title LIKE '%' || sqlc.arg(search) || '%'
    OR description LIKE '%' || sqlc.arg(search) || '%'
  )
ORDER BY amount DESC, created_at DESC
LIMIT sqlc.arg(limit) OFFSET sqlc.arg(offset);

-- name: UpdateEvent :one
UPDATE events
SET title = ?, time = ?, description = ?, amount = ?
WHERE id = ? AND group_id = ?
RETURNING *;

-- name: DeleteEvent :exec
DELETE FROM events
WHERE id = ? AND group_id = ?;

-- name: DeleteAllEvents :exec
DELETE FROM events;
