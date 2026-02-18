-- name: CreateMember :one
INSERT INTO members (id, group_id, name, description)
VALUES (?, ?, ?, ?)
RETURNING *;

-- name: GetMember :one
SELECT * FROM members
WHERE id = ? AND group_id = ?;

-- name: ListMembers :many
SELECT * FROM members
WHERE group_id = ?
ORDER BY created_at DESC;

-- name: UpdateMember :one
UPDATE members
SET name = ?, description = ?
WHERE id = ? AND group_id = ?
RETURNING *;

-- name: DeleteMember :exec
DELETE FROM members
WHERE id = ? AND group_id = ?;
