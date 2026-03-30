-- name: CreateUser :one
INSERT INTO users (id, email, preferred_lang)
VALUES (sqlc.arg(id), sqlc.arg(email), COALESCE(NULLIF(sqlc.arg(preferred_lang), ''), 'hu'))
RETURNING *;

-- name: UpdateUserPreferredLang :exec
UPDATE users
SET preferred_lang = ?
WHERE id = ?;

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
SELECT g.*
FROM groups g
JOIN group_access ga ON ga.group_id = g.id
WHERE ga.user_id = sqlc.arg(user_id)
  AND ga.role IN ('owner', 'admin')
LIMIT 1;

-- name: ListGroupsByAdmin :many
SELECT g.*
FROM groups g
JOIN group_access ga ON ga.group_id = g.id
WHERE ga.user_id = sqlc.arg(user_id)
  AND ga.role IN ('owner', 'admin')
ORDER BY created_at DESC;

-- name: ListGroupsByReader :many
SELECT g.*
FROM groups g
JOIN group_access ga ON ga.group_id = g.id
WHERE ga.user_id = sqlc.arg(user_id)
  AND ga.role = 'viewer'
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
INSERT INTO group_access (id, user_id, group_id, role)
VALUES (?, ?, ?, 'viewer')
ON CONFLICT(user_id, group_id) DO UPDATE SET role = 'viewer'
WHERE group_access.role != 'owner'
RETURNING id, user_id, group_id, created_at;

-- name: GetGroupReaders :many
SELECT users.* FROM users
JOIN group_access ON group_access.user_id = users.id
WHERE group_access.group_id = ?
  AND group_access.role = 'viewer';

-- name: ListGroupUserAccess :many
SELECT
  users.id,
  users.email,
  users.created_at,
  users.preferred_lang,
  group_access.role,
  group_access.created_at AS access_created_at
FROM users
JOIN group_access ON group_access.user_id = users.id
WHERE group_access.group_id = ?
ORDER BY group_access.created_at ASC, LOWER(users.email) ASC;

-- name: CountGroupReadersFiltered :one
SELECT COUNT(*) FROM users
JOIN group_access ON group_access.user_id = users.id
WHERE group_access.group_id = sqlc.arg(group_id)
  AND group_access.role = 'viewer'
  AND (
    sqlc.arg(search) = ''
    OR LOWER(users.email) LIKE '%' || LOWER(sqlc.arg(search)) || '%'
  );

-- name: ListGroupReadersByEmailAscFiltered :many
SELECT users.* FROM users
JOIN group_access ON group_access.user_id = users.id
WHERE group_access.group_id = sqlc.arg(group_id)
  AND group_access.role = 'viewer'
  AND (
    sqlc.arg(search) = ''
    OR LOWER(users.email) LIKE '%' || LOWER(sqlc.arg(search)) || '%'
  )
ORDER BY LOWER(users.email) ASC
LIMIT sqlc.arg(limit) OFFSET sqlc.arg(offset);

-- name: ListGroupReadersByEmailDescFiltered :many
SELECT users.* FROM users
JOIN group_access ON group_access.user_id = users.id
WHERE group_access.group_id = sqlc.arg(group_id)
  AND group_access.role = 'viewer'
  AND (
    sqlc.arg(search) = ''
    OR LOWER(users.email) LIKE '%' || LOWER(sqlc.arg(search)) || '%'
  )
ORDER BY LOWER(users.email) DESC
LIMIT sqlc.arg(limit) OFFSET sqlc.arg(offset);

-- name: IsGroupReader :one
SELECT COUNT(*) FROM group_access
WHERE user_id = ?
  AND group_id = ?
  AND role = 'viewer';

-- name: GetGroupAccessRole :one
SELECT role
FROM group_access
WHERE user_id = ?
  AND group_id = ?;

-- name: RemoveGroupReader :exec
DELETE FROM group_access
WHERE user_id = ?
  AND group_id = ?
  AND role = 'viewer';

-- name: CreateMagicLink :one
INSERT INTO magic_links (id, token, email, action, group_id, expires_at)
VALUES (?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: CreateInviteMagicLink :one
INSERT INTO magic_links (id, token, email, action, group_id, expires_at, invite_role)
VALUES (?, ?, ?, 'invite', ?, CURRENT_TIMESTAMP, ?)
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
WHERE action != 'invite'
  AND expires_at < CURRENT_TIMESTAMP
  AND used_at IS NULL;

-- name: ListGroupPendingInvites :many
SELECT * FROM magic_links
WHERE action = 'invite'
  AND group_id = ?
  AND used_at IS NULL
ORDER BY created_at DESC;

-- name: CountGroupPendingInvitesFiltered :one
SELECT COUNT(*) FROM magic_links
WHERE action = 'invite'
  AND group_id = sqlc.arg(group_id)
  AND used_at IS NULL
  AND (
    sqlc.arg(search) = ''
    OR LOWER(email) LIKE '%' || LOWER(sqlc.arg(search)) || '%'
  );

-- name: ListGroupPendingInvitesByEmailAscFiltered :many
SELECT * FROM magic_links
WHERE action = 'invite'
  AND group_id = sqlc.arg(group_id)
  AND used_at IS NULL
  AND (
    sqlc.arg(search) = ''
    OR LOWER(email) LIKE '%' || LOWER(sqlc.arg(search)) || '%'
  )
ORDER BY LOWER(email) ASC, created_at DESC
LIMIT sqlc.arg(limit) OFFSET sqlc.arg(offset);

-- name: ListGroupPendingInvitesByEmailDescFiltered :many
SELECT * FROM magic_links
WHERE action = 'invite'
  AND group_id = sqlc.arg(group_id)
  AND used_at IS NULL
  AND (
    sqlc.arg(search) = ''
    OR LOWER(email) LIKE '%' || LOWER(sqlc.arg(search)) || '%'
  )
ORDER BY LOWER(email) DESC, created_at DESC
LIMIT sqlc.arg(limit) OFFSET sqlc.arg(offset);

-- name: ListGroupPendingInvitesByCreatedAscFiltered :many
SELECT * FROM magic_links
WHERE action = 'invite'
  AND group_id = sqlc.arg(group_id)
  AND used_at IS NULL
  AND (
    sqlc.arg(search) = ''
    OR LOWER(email) LIKE '%' || LOWER(sqlc.arg(search)) || '%'
  )
ORDER BY created_at ASC, LOWER(email) ASC
LIMIT sqlc.arg(limit) OFFSET sqlc.arg(offset);

-- name: ListGroupPendingInvitesByCreatedDescFiltered :many
SELECT * FROM magic_links
WHERE action = 'invite'
  AND group_id = sqlc.arg(group_id)
  AND used_at IS NULL
  AND (
    sqlc.arg(search) = ''
    OR LOWER(email) LIKE '%' || LOWER(sqlc.arg(search)) || '%'
  )
ORDER BY created_at DESC, LOWER(email) ASC
LIMIT sqlc.arg(limit) OFFSET sqlc.arg(offset);

-- name: DeleteGroupPendingInvite :exec
DELETE FROM magic_links
WHERE id = ?
  AND action = 'invite'
  AND group_id = ?
  AND used_at IS NULL;

-- name: CreateGroupAdmin :one
INSERT INTO group_access (id, user_id, group_id, role)
VALUES (?, ?, ?, 'admin')
ON CONFLICT(user_id, group_id) DO UPDATE SET role = 'admin'
WHERE group_access.role != 'owner'
RETURNING id, user_id, group_id, created_at;

-- name: IsGroupAdmin :one
SELECT COUNT(*)
FROM group_access
WHERE user_id = ?
  AND group_id = ?
  AND role = 'admin';

-- name: RemoveGroupAdmin :exec
DELETE FROM group_access
WHERE user_id = ?
  AND group_id = ?
  AND role = 'admin';

-- name: ListGroupAdminUserIDs :many
SELECT user_id
FROM group_access
WHERE group_id = ?
  AND role = 'admin';

-- name: ListGroupAdmins :many
SELECT u.*
FROM users u
JOIN group_access ga ON ga.user_id = u.id
WHERE ga.group_id = ?
  AND ga.role IN ('owner', 'admin')
ORDER BY LOWER(u.email) ASC;

-- name: CountGroupAdminsFiltered :one
SELECT COUNT(*)
FROM users u
JOIN group_access ga ON ga.user_id = u.id
WHERE ga.group_id = sqlc.arg(group_id)
  AND ga.role IN ('owner', 'admin')
  AND (
    sqlc.arg(search) = ''
    OR LOWER(u.email) LIKE '%' || LOWER(sqlc.arg(search)) || '%'
  );

-- name: ListGroupAdminsByEmailAscFiltered :many
SELECT u.*
FROM users u
JOIN group_access ga ON ga.user_id = u.id
WHERE ga.group_id = sqlc.arg(group_id)
  AND ga.role IN ('owner', 'admin')
  AND (
    sqlc.arg(search) = ''
    OR LOWER(u.email) LIKE '%' || LOWER(sqlc.arg(search)) || '%'
  )
ORDER BY LOWER(email) ASC
LIMIT sqlc.arg(limit) OFFSET sqlc.arg(offset);

-- name: ListGroupAdminsByEmailDescFiltered :many
SELECT u.*
FROM users u
JOIN group_access ga ON ga.user_id = u.id
WHERE ga.group_id = sqlc.arg(group_id)
  AND ga.role IN ('owner', 'admin')
  AND (
    sqlc.arg(search) = ''
    OR LOWER(u.email) LIKE '%' || LOWER(sqlc.arg(search)) || '%'
  )
ORDER BY LOWER(email) DESC
LIMIT sqlc.arg(limit) OFFSET sqlc.arg(offset);

-- name: CountUserGroupsFiltered :one
SELECT COUNT(*)
FROM groups g
JOIN group_access ga ON ga.group_id = g.id
JOIN users u ON u.id = g.admin_user_id
WHERE ga.user_id = sqlc.arg(user_id)
  AND (
    sqlc.arg(search) = ''
    OR g.name LIKE '%' || sqlc.arg(search) || '%'
    OR u.email LIKE '%' || sqlc.arg(search) || '%'
  );

-- name: ListUserGroupsByNameAscFiltered :many
SELECT
  g.id,
  g.name,
  g.admin_user_id,
  g.created_at,
  ga.role
FROM groups g
JOIN group_access ga ON ga.group_id = g.id
JOIN users u ON u.id = g.admin_user_id
WHERE ga.user_id = sqlc.arg(user_id)
  AND (
    sqlc.arg(search) = ''
    OR g.name LIKE '%' || sqlc.arg(search) || '%'
    OR u.email LIKE '%' || sqlc.arg(search) || '%'
  )
ORDER BY g.name COLLATE NOCASE ASC, g.created_at DESC
LIMIT sqlc.arg(limit) OFFSET sqlc.arg(offset);

-- name: ListUserGroupsByNameDescFiltered :many
SELECT
  g.id,
  g.name,
  g.admin_user_id,
  g.created_at,
  ga.role
FROM groups g
JOIN group_access ga ON ga.group_id = g.id
JOIN users u ON u.id = g.admin_user_id
WHERE ga.user_id = sqlc.arg(user_id)
  AND (
    sqlc.arg(search) = ''
    OR g.name LIKE '%' || sqlc.arg(search) || '%'
    OR u.email LIKE '%' || sqlc.arg(search) || '%'
  )
ORDER BY g.name COLLATE NOCASE DESC, g.created_at DESC
LIMIT sqlc.arg(limit) OFFSET sqlc.arg(offset);

-- name: ListUserGroupsByCreatedAscFiltered :many
SELECT
  g.id,
  g.name,
  g.admin_user_id,
  g.created_at,
  ga.role
FROM groups g
JOIN group_access ga ON ga.group_id = g.id
JOIN users u ON u.id = g.admin_user_id
WHERE ga.user_id = sqlc.arg(user_id)
  AND (
    sqlc.arg(search) = ''
    OR g.name LIKE '%' || sqlc.arg(search) || '%'
    OR u.email LIKE '%' || sqlc.arg(search) || '%'
  )
ORDER BY g.created_at ASC, g.name COLLATE NOCASE ASC
LIMIT sqlc.arg(limit) OFFSET sqlc.arg(offset);

-- name: ListUserGroupsByCreatedDescFiltered :many
SELECT
  g.id,
  g.name,
  g.admin_user_id,
  g.created_at,
  ga.role
FROM groups g
JOIN group_access ga ON ga.group_id = g.id
JOIN users u ON u.id = g.admin_user_id
WHERE ga.user_id = sqlc.arg(user_id)
  AND (
    sqlc.arg(search) = ''
    OR g.name LIKE '%' || sqlc.arg(search) || '%'
    OR u.email LIKE '%' || sqlc.arg(search) || '%'
  )
ORDER BY g.created_at DESC, g.name COLLATE NOCASE ASC
LIMIT sqlc.arg(limit) OFFSET sqlc.arg(offset);

-- name: ListUserGroupsByAdminAscFiltered :many
SELECT
  g.id,
  g.name,
  g.admin_user_id,
  g.created_at,
  ga.role,
  u.email as admin_email
FROM groups g
JOIN group_access ga ON ga.group_id = g.id
JOIN users u ON u.id = g.admin_user_id
WHERE ga.user_id = sqlc.arg(user_id)
  AND (
    sqlc.arg(search) = ''
    OR g.name LIKE '%' || sqlc.arg(search) || '%'
    OR u.email LIKE '%' || sqlc.arg(search) || '%'
  )
ORDER BY admin_email COLLATE NOCASE ASC, g.name COLLATE NOCASE ASC
LIMIT sqlc.arg(limit) OFFSET sqlc.arg(offset);

-- name: ListUserGroupsByAdminDescFiltered :many
SELECT
  g.id,
  g.name,
  g.admin_user_id,
  g.created_at,
  ga.role,
  u.email as admin_email
FROM groups g
JOIN group_access ga ON ga.group_id = g.id
JOIN users u ON u.id = g.admin_user_id
WHERE ga.user_id = sqlc.arg(user_id)
  AND (
    sqlc.arg(search) = ''
    OR g.name LIKE '%' || sqlc.arg(search) || '%'
    OR u.email LIKE '%' || sqlc.arg(search) || '%'
  )
ORDER BY admin_email COLLATE NOCASE DESC, g.name COLLATE NOCASE ASC
LIMIT sqlc.arg(limit) OFFSET sqlc.arg(offset);

-- name: CreateUserSession :one
INSERT INTO user_sessions (id, user_id, token, expires_at)
VALUES (?, ?, ?, ?)
RETURNING *;

-- name: GetUserSessionByToken :one
SELECT * FROM user_sessions
WHERE token = ? AND expires_at > CURRENT_TIMESTAMP;

-- name: ListUserSessions :many
SELECT * FROM user_sessions
WHERE user_id = ? AND expires_at > CURRENT_TIMESTAMP
ORDER BY created_at DESC;

-- name: DeleteUserSession :exec
DELETE FROM user_sessions
WHERE id = ? AND user_id = ?;

-- name: DeleteOtherUserSessions :exec
DELETE FROM user_sessions
WHERE user_id = ? AND id != ?;

-- name: DeleteExpiredUserSessions :exec
DELETE FROM user_sessions
WHERE expires_at < CURRENT_TIMESTAMP;

-- name: DeleteAllUserSessions :exec
DELETE FROM user_sessions
WHERE user_id = ?;

-- name: DeleteUserSessionByID :exec
DELETE FROM user_sessions
WHERE id = ?;

-- name: CountSessionsFiltered :one
SELECT COUNT(*)
FROM user_sessions us
JOIN users u ON u.id = us.user_id
WHERE us.expires_at > CURRENT_TIMESTAMP
  AND (
    sqlc.arg(search) = ''
    OR u.email LIKE '%' || sqlc.arg(search) || '%'
    OR us.id LIKE '%' || sqlc.arg(search) || '%'
  );

-- name: ListSessionsByCreatedDescFiltered :many
SELECT us.id, us.user_id, u.email AS user_email, us.created_at, us.expires_at
FROM user_sessions us
JOIN users u ON u.id = us.user_id
WHERE us.expires_at > CURRENT_TIMESTAMP
  AND (
    sqlc.arg(search) = ''
    OR u.email LIKE '%' || sqlc.arg(search) || '%'
    OR us.id LIKE '%' || sqlc.arg(search) || '%'
  )
ORDER BY us.created_at DESC, u.email COLLATE NOCASE ASC
LIMIT sqlc.arg(limit) OFFSET sqlc.arg(offset);

-- name: ListSessionsByCreatedAscFiltered :many
SELECT us.id, us.user_id, u.email AS user_email, us.created_at, us.expires_at
FROM user_sessions us
JOIN users u ON u.id = us.user_id
WHERE us.expires_at > CURRENT_TIMESTAMP
  AND (
    sqlc.arg(search) = ''
    OR u.email LIKE '%' || sqlc.arg(search) || '%'
    OR us.id LIKE '%' || sqlc.arg(search) || '%'
  )
ORDER BY us.created_at ASC, u.email COLLATE NOCASE ASC
LIMIT sqlc.arg(limit) OFFSET sqlc.arg(offset);

-- name: ListSessionsByEmailAscFiltered :many
SELECT us.id, us.user_id, u.email AS user_email, us.created_at, us.expires_at
FROM user_sessions us
JOIN users u ON u.id = us.user_id
WHERE us.expires_at > CURRENT_TIMESTAMP
  AND (
    sqlc.arg(search) = ''
    OR u.email LIKE '%' || sqlc.arg(search) || '%'
    OR us.id LIKE '%' || sqlc.arg(search) || '%'
  )
ORDER BY u.email COLLATE NOCASE ASC, us.created_at DESC
LIMIT sqlc.arg(limit) OFFSET sqlc.arg(offset);

-- name: ListSessionsByEmailDescFiltered :many
SELECT us.id, us.user_id, u.email AS user_email, us.created_at, us.expires_at
FROM user_sessions us
JOIN users u ON u.id = us.user_id
WHERE us.expires_at > CURRENT_TIMESTAMP
  AND (
    sqlc.arg(search) = ''
    OR u.email LIKE '%' || sqlc.arg(search) || '%'
    OR us.id LIKE '%' || sqlc.arg(search) || '%'
  )
ORDER BY u.email COLLATE NOCASE DESC, us.created_at DESC
LIMIT sqlc.arg(limit) OFFSET sqlc.arg(offset);
