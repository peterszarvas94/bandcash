-- name: CreateEntry :one
INSERT INTO entries (title, time, description, amount)
VALUES (?, ?, ?, ?)
RETURNING *;

-- name: GetEntry :one
SELECT * FROM entries
WHERE id = ?;

-- name: ListEntries :many
SELECT * FROM entries
ORDER BY time ASC;

-- name: UpdateEntry :one
UPDATE entries
SET title = ?, time = ?, description = ?, amount = ?
WHERE id = ?
RETURNING *;

-- name: DeleteEntry :exec
DELETE FROM entries
WHERE id = ?;

-- name: DeleteAllEntries :exec
DELETE FROM entries;
