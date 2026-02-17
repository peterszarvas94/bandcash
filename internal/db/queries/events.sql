-- name: CreateEvent :one
INSERT INTO events (title, time, description, amount)
VALUES (?, ?, ?, ?)
RETURNING *;

-- name: GetEvent :one
SELECT * FROM events
WHERE id = ?;

-- name: ListEvents :many
SELECT * FROM events
ORDER BY time ASC;

-- name: UpdateEvent :one
UPDATE events
SET title = ?, time = ?, description = ?, amount = ?
WHERE id = ?
RETURNING *;

-- name: DeleteEvent :exec
DELETE FROM events
WHERE id = ?;

-- name: DeleteAllEvents :exec
DELETE FROM events;
