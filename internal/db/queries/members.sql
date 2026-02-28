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

-- name: CountMembersFiltered :one
SELECT COUNT(*) FROM members
WHERE group_id = sqlc.arg(group_id)
  AND (
    sqlc.arg(search) = ''
    OR name LIKE '%' || sqlc.arg(search) || '%'
    OR description LIKE '%' || sqlc.arg(search) || '%'
  );

-- name: ListMembersByNameAscFiltered :many
SELECT * FROM members
WHERE group_id = sqlc.arg(group_id)
  AND (
    sqlc.arg(search) = ''
    OR name LIKE '%' || sqlc.arg(search) || '%'
    OR description LIKE '%' || sqlc.arg(search) || '%'
  )
ORDER BY name COLLATE NOCASE ASC, created_at DESC
LIMIT sqlc.arg(limit) OFFSET sqlc.arg(offset);

-- name: ListMembersByNameDescFiltered :many
SELECT * FROM members
WHERE group_id = sqlc.arg(group_id)
  AND (
    sqlc.arg(search) = ''
    OR name LIKE '%' || sqlc.arg(search) || '%'
    OR description LIKE '%' || sqlc.arg(search) || '%'
  )
ORDER BY name COLLATE NOCASE DESC, created_at DESC
LIMIT sqlc.arg(limit) OFFSET sqlc.arg(offset);

-- name: ListMembersByCreatedAtAscFiltered :many
SELECT * FROM members
WHERE group_id = sqlc.arg(group_id)
  AND (
    sqlc.arg(search) = ''
    OR name LIKE '%' || sqlc.arg(search) || '%'
    OR description LIKE '%' || sqlc.arg(search) || '%'
  )
ORDER BY created_at ASC
LIMIT sqlc.arg(limit) OFFSET sqlc.arg(offset);

-- name: ListMembersByCreatedAtDescFiltered :many
SELECT * FROM members
WHERE group_id = sqlc.arg(group_id)
  AND (
    sqlc.arg(search) = ''
    OR name LIKE '%' || sqlc.arg(search) || '%'
    OR description LIKE '%' || sqlc.arg(search) || '%'
  )
ORDER BY created_at DESC
LIMIT sqlc.arg(limit) OFFSET sqlc.arg(offset);

-- name: UpdateMember :one
UPDATE members
SET name = ?, description = ?
WHERE id = ? AND group_id = ?
RETURNING *;

-- name: DeleteMember :exec
DELETE FROM members
WHERE id = ? AND group_id = ?;
