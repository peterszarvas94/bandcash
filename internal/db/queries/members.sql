-- name: CreateMember :one
INSERT INTO members (name, description)
VALUES (?, ?)
RETURNING *;

-- name: GetMember :one
SELECT * FROM members
WHERE id = ?;

-- name: ListMembers :many
SELECT * FROM members
ORDER BY created_at DESC;

-- name: UpdateMember :one
UPDATE members
SET name = ?, description = ?
WHERE id = ?
RETURNING *;

-- name: DeleteMember :exec
DELETE FROM members
WHERE id = ?;
