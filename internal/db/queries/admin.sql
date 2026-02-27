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

-- name: ListRecentGroups :many
SELECT * FROM groups
ORDER BY created_at DESC
LIMIT ?;
