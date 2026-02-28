-- name: CreateUser :one
INSERT INTO users (id, email)
VALUES (?, ?)
RETURNING *;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = ?;

-- name: GetUserByID :one
SELECT * FROM users
WHERE id = ?;

-- name: IsUserBanned :one
SELECT COUNT(*) FROM banned_users
WHERE user_id = ?;

-- name: BanUser :exec
INSERT INTO banned_users (id, user_id)
VALUES (?, ?)
ON CONFLICT(user_id) DO NOTHING;

-- name: UnbanUser :exec
DELETE FROM banned_users
WHERE user_id = ?;

-- name: CreateGroup :one
INSERT INTO groups (id, name, admin_user_id)
VALUES (?, ?, ?)
RETURNING *;

-- name: GetGroupByID :one
SELECT * FROM groups
WHERE id = ?;

-- name: GetGroupByAdmin :one
SELECT * FROM groups
WHERE admin_user_id = ?;

-- name: ListGroupsByAdmin :many
SELECT id, name, admin_user_id, created_at
FROM groups
WHERE admin_user_id = ?
ORDER BY created_at DESC;

-- name: ListGroupsByReader :many
SELECT g.id, g.name, g.admin_user_id, g.created_at
FROM groups g
JOIN group_readers gr ON gr.group_id = g.id
WHERE gr.user_id = ?
ORDER BY g.created_at DESC;

-- name: UpdateGroupAdmin :exec
UPDATE groups
SET admin_user_id = ?
WHERE id = ?;

-- name: UpdateGroupName :one
UPDATE groups
SET name = ?
WHERE id = ?
RETURNING *;

-- name: DeleteGroup :exec
DELETE FROM groups
WHERE id = ?;

-- name: CreateGroupReader :one
INSERT INTO group_readers (id, user_id, group_id)
VALUES (?, ?, ?)
RETURNING *;

-- name: GetGroupReaders :many
SELECT users.* FROM users
JOIN group_readers ON group_readers.user_id = users.id
WHERE group_readers.group_id = ?;

-- name: IsGroupReader :one
SELECT COUNT(*) FROM group_readers
WHERE user_id = ? AND group_id = ?;

-- name: RemoveGroupReader :exec
DELETE FROM group_readers
WHERE user_id = ? AND group_id = ?;

-- name: CreateMagicLink :one
INSERT INTO magic_links (id, token, email, action, group_id, expires_at)
VALUES (?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: GetMagicLinkByToken :one
SELECT * FROM magic_links
WHERE token = ?;

-- name: UseMagicLink :exec
UPDATE magic_links
SET used_at = CURRENT_TIMESTAMP
WHERE id = ?;

-- name: DeleteExpiredMagicLinks :exec
DELETE FROM magic_links
WHERE expires_at < CURRENT_TIMESTAMP AND used_at IS NULL;

-- name: ListGroupPendingInvites :many
SELECT * FROM magic_links
WHERE action = 'invite'
  AND group_id = ?
  AND used_at IS NULL
  AND expires_at >= CURRENT_TIMESTAMP
ORDER BY created_at DESC;

-- name: DeleteGroupPendingInvite :exec
DELETE FROM magic_links
WHERE id = ?
  AND action = 'invite'
  AND group_id = ?
  AND used_at IS NULL;

-- name: CountUserGroupsFiltered :one
SELECT COUNT(*) FROM (
  SELECT g.id FROM groups g
  WHERE g.admin_user_id = sqlc.arg(user_id)
    AND (
      sqlc.arg(search) = ''
      OR g.name LIKE '%' || sqlc.arg(search) || '%'
    )
  UNION
  SELECT g.id FROM groups g
  JOIN group_readers gr ON gr.group_id = g.id
  WHERE gr.user_id = sqlc.arg(user_id)
    AND g.admin_user_id != sqlc.arg(user_id)
    AND (
      sqlc.arg(search) = ''
      OR g.name LIKE '%' || sqlc.arg(search) || '%'
    )
);

-- name: ListUserGroupsByNameAscFiltered :many
SELECT 
  g.id,
  g.name,
  g.admin_user_id,
  g.created_at,
  CASE WHEN g.admin_user_id = sqlc.arg(user_id) THEN 'admin' ELSE 'viewer' END as role
FROM groups g
WHERE g.admin_user_id = sqlc.arg(user_id)
  AND (
    sqlc.arg(search) = ''
    OR g.name LIKE '%' || sqlc.arg(search) || '%'
  )
UNION ALL
SELECT 
  g.id,
  g.name,
  g.admin_user_id,
  g.created_at,
  'viewer' as role
FROM groups g
JOIN group_readers gr ON gr.group_id = g.id
WHERE gr.user_id = sqlc.arg(user_id)
  AND g.admin_user_id != sqlc.arg(user_id)
  AND (
    sqlc.arg(search) = ''
    OR g.name LIKE '%' || sqlc.arg(search) || '%'
  )
ORDER BY name COLLATE NOCASE ASC, created_at DESC
LIMIT sqlc.arg(limit) OFFSET sqlc.arg(offset);

-- name: ListUserGroupsByNameDescFiltered :many
SELECT 
  g.id,
  g.name,
  g.admin_user_id,
  g.created_at,
  CASE WHEN g.admin_user_id = sqlc.arg(user_id) THEN 'admin' ELSE 'viewer' END as role
FROM groups g
WHERE g.admin_user_id = sqlc.arg(user_id)
  AND (
    sqlc.arg(search) = ''
    OR g.name LIKE '%' || sqlc.arg(search) || '%'
  )
UNION ALL
SELECT 
  g.id,
  g.name,
  g.admin_user_id,
  g.created_at,
  'viewer' as role
FROM groups g
JOIN group_readers gr ON gr.group_id = g.id
WHERE gr.user_id = sqlc.arg(user_id)
  AND g.admin_user_id != sqlc.arg(user_id)
  AND (
    sqlc.arg(search) = ''
    OR g.name LIKE '%' || sqlc.arg(search) || '%'
  )
ORDER BY name COLLATE NOCASE DESC, created_at DESC
LIMIT sqlc.arg(limit) OFFSET sqlc.arg(offset);

-- name: ListUserGroupsByCreatedAscFiltered :many
SELECT 
  g.id,
  g.name,
  g.admin_user_id,
  g.created_at,
  CASE WHEN g.admin_user_id = sqlc.arg(user_id) THEN 'admin' ELSE 'viewer' END as role
FROM groups g
WHERE g.admin_user_id = sqlc.arg(user_id)
  AND (
    sqlc.arg(search) = ''
    OR g.name LIKE '%' || sqlc.arg(search) || '%'
  )
UNION ALL
SELECT 
  g.id,
  g.name,
  g.admin_user_id,
  g.created_at,
  'viewer' as role
FROM groups g
JOIN group_readers gr ON gr.group_id = g.id
WHERE gr.user_id = sqlc.arg(user_id)
  AND g.admin_user_id != sqlc.arg(user_id)
  AND (
    sqlc.arg(search) = ''
    OR g.name LIKE '%' || sqlc.arg(search) || '%'
  )
ORDER BY created_at ASC, name COLLATE NOCASE ASC
LIMIT sqlc.arg(limit) OFFSET sqlc.arg(offset);

-- name: ListUserGroupsByCreatedDescFiltered :many
SELECT 
  g.id,
  g.name,
  g.admin_user_id,
  g.created_at,
  CASE WHEN g.admin_user_id = sqlc.arg(user_id) THEN 'admin' ELSE 'viewer' END as role
FROM groups g
WHERE g.admin_user_id = sqlc.arg(user_id)
  AND (
    sqlc.arg(search) = ''
    OR g.name LIKE '%' || sqlc.arg(search) || '%'
  )
UNION ALL
SELECT 
  g.id,
  g.name,
  g.admin_user_id,
  g.created_at,
  'viewer' as role
FROM groups g
JOIN group_readers gr ON gr.group_id = g.id
WHERE gr.user_id = sqlc.arg(user_id)
  AND g.admin_user_id != sqlc.arg(user_id)
  AND (
    sqlc.arg(search) = ''
    OR g.name LIKE '%' || sqlc.arg(search) || '%'
  )
ORDER BY created_at DESC, name COLLATE NOCASE ASC
LIMIT sqlc.arg(limit) OFFSET sqlc.arg(offset);
