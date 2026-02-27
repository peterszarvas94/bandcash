-- name: CountUsers :one
SELECT COUNT(*) FROM users;

-- name: CountGroups :one
SELECT COUNT(*) FROM groups;

-- name: CountEvents :one
SELECT COUNT(*) FROM events;

-- name: CountMembers :one
SELECT COUNT(*) FROM members;

-- name: ListRecentUsers :many
SELECT * FROM users
ORDER BY created_at DESC
LIMIT ?;

-- name: ListRecentGroups :many
SELECT * FROM groups
ORDER BY created_at DESC
LIMIT ?;
