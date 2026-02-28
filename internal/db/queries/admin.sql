-- name: CountUsers :one
SELECT COUNT(*) FROM users;

-- name: CountGroups :one
SELECT COUNT(*) FROM groups;

-- name: CountEvents :one
SELECT COUNT(*) FROM events;

-- name: CountMembers :one
SELECT COUNT(*) FROM members;

-- name: ListRecentUsersWithBanStatus :many
SELECT
  users.id,
  users.email,
  users.created_at,
  CASE WHEN banned_users.user_id IS NULL THEN 0 ELSE 1 END AS is_banned
FROM users
LEFT JOIN banned_users ON banned_users.user_id = users.id
ORDER BY users.created_at DESC
LIMIT ?;

-- name: CountUsersFiltered :one
SELECT COUNT(*) FROM users
WHERE sqlc.arg(search) = '' OR LOWER(email) LIKE '%' || LOWER(sqlc.arg(search)) || '%';

-- name: ListUsersByEmailAscFiltered :many
SELECT
  users.id,
  users.email,
  users.created_at,
  CASE WHEN banned_users.user_id IS NULL THEN 0 ELSE 1 END AS is_banned
FROM users
LEFT JOIN banned_users ON banned_users.user_id = users.id
WHERE sqlc.arg(search) = '' OR LOWER(email) LIKE '%' || LOWER(sqlc.arg(search)) || '%'
ORDER BY LOWER(email) ASC, users.created_at DESC
LIMIT sqlc.arg(limit) OFFSET sqlc.arg(offset);

-- name: ListUsersByEmailDescFiltered :many
SELECT
  users.id,
  users.email,
  users.created_at,
  CASE WHEN banned_users.user_id IS NULL THEN 0 ELSE 1 END AS is_banned
FROM users
LEFT JOIN banned_users ON banned_users.user_id = users.id
WHERE sqlc.arg(search) = '' OR LOWER(email) LIKE '%' || LOWER(sqlc.arg(search)) || '%'
ORDER BY LOWER(email) DESC, users.created_at DESC
LIMIT sqlc.arg(limit) OFFSET sqlc.arg(offset);

-- name: ListUsersByCreatedAscFiltered :many
SELECT
  users.id,
  users.email,
  users.created_at,
  CASE WHEN banned_users.user_id IS NULL THEN 0 ELSE 1 END AS is_banned
FROM users
LEFT JOIN banned_users ON banned_users.user_id = users.id
WHERE sqlc.arg(search) = '' OR LOWER(email) LIKE '%' || LOWER(sqlc.arg(search)) || '%'
ORDER BY users.created_at ASC, LOWER(email) ASC
LIMIT sqlc.arg(limit) OFFSET sqlc.arg(offset);

-- name: ListUsersByCreatedDescFiltered :many
SELECT
  users.id,
  users.email,
  users.created_at,
  CASE WHEN banned_users.user_id IS NULL THEN 0 ELSE 1 END AS is_banned
FROM users
LEFT JOIN banned_users ON banned_users.user_id = users.id
WHERE sqlc.arg(search) = '' OR LOWER(email) LIKE '%' || LOWER(sqlc.arg(search)) || '%'
ORDER BY users.created_at DESC, LOWER(email) ASC
LIMIT sqlc.arg(limit) OFFSET sqlc.arg(offset);

-- name: ListRecentGroups :many
SELECT * FROM groups
ORDER BY created_at DESC
LIMIT ?;

-- name: ListAllGroups :many
SELECT * FROM groups
ORDER BY created_at DESC;

-- name: CountGroupsFiltered :one
SELECT COUNT(*) FROM groups
WHERE sqlc.arg(search) = '' OR LOWER(name) LIKE '%' || LOWER(sqlc.arg(search)) || '%';

-- name: ListGroupsByNameAscFiltered :many
SELECT
  groups.id,
  groups.name,
  groups.admin_user_id,
  groups.created_at
FROM groups
WHERE sqlc.arg(search) = '' OR LOWER(name) LIKE '%' || LOWER(sqlc.arg(search)) || '%'
ORDER BY LOWER(groups.name) ASC, groups.created_at DESC
LIMIT sqlc.arg(limit) OFFSET sqlc.arg(offset);

-- name: ListGroupsByNameDescFiltered :many
SELECT
  groups.id,
  groups.name,
  groups.admin_user_id,
  groups.created_at
FROM groups
WHERE sqlc.arg(search) = '' OR LOWER(name) LIKE '%' || LOWER(sqlc.arg(search)) || '%'
ORDER BY LOWER(groups.name) DESC, groups.created_at DESC
LIMIT sqlc.arg(limit) OFFSET sqlc.arg(offset);

-- name: ListGroupsByCreatedAscFiltered :many
SELECT
  groups.id,
  groups.name,
  groups.admin_user_id,
  groups.created_at
FROM groups
WHERE sqlc.arg(search) = '' OR LOWER(name) LIKE '%' || LOWER(sqlc.arg(search)) || '%'
ORDER BY groups.created_at ASC, LOWER(groups.name) ASC
LIMIT sqlc.arg(limit) OFFSET sqlc.arg(offset);

-- name: ListGroupsByCreatedDescFiltered :many
SELECT
  groups.id,
  groups.name,
  groups.admin_user_id,
  groups.created_at
FROM groups
WHERE sqlc.arg(search) = '' OR LOWER(name) LIKE '%' || LOWER(sqlc.arg(search)) || '%'
ORDER BY groups.created_at DESC, LOWER(groups.name) ASC
LIMIT sqlc.arg(limit) OFFSET sqlc.arg(offset);