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
