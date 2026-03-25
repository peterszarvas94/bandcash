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
SELECT * FROM groups
WHERE admin_user_id = sqlc.arg(owner_user_id)
   OR id IN (
     SELECT group_id
     FROM group_admins
     WHERE user_id = sqlc.arg(user_id)
   )
LIMIT 1;

-- name: ListGroupsByAdmin :many
SELECT *
FROM groups
WHERE admin_user_id = sqlc.arg(owner_user_id)
   OR id IN (
     SELECT group_id
     FROM group_admins
     WHERE user_id = sqlc.arg(user_id)
   )
ORDER BY created_at DESC;

-- name: ListGroupsByReader :many
SELECT g.*
FROM groups g
JOIN group_readers gr ON gr.group_id = g.id
WHERE gr.user_id = ?
  AND g.admin_user_id != ?
  AND NOT EXISTS (
    SELECT 1
    FROM group_admins ga
    WHERE ga.group_id = g.id
      AND ga.user_id = gr.user_id
  )
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

-- name: CountGroupReadersFiltered :one
SELECT COUNT(*) FROM users
JOIN group_readers ON group_readers.user_id = users.id
WHERE group_readers.group_id = sqlc.arg(group_id)
  AND (
    sqlc.arg(search) = ''
    OR LOWER(users.email) LIKE '%' || LOWER(sqlc.arg(search)) || '%'
  );

-- name: ListGroupReadersByEmailAscFiltered :many
SELECT users.* FROM users
JOIN group_readers ON group_readers.user_id = users.id
WHERE group_readers.group_id = sqlc.arg(group_id)
  AND (
    sqlc.arg(search) = ''
    OR LOWER(users.email) LIKE '%' || LOWER(sqlc.arg(search)) || '%'
  )
ORDER BY LOWER(users.email) ASC
LIMIT sqlc.arg(limit) OFFSET sqlc.arg(offset);

-- name: ListGroupReadersByEmailDescFiltered :many
SELECT users.* FROM users
JOIN group_readers ON group_readers.user_id = users.id
WHERE group_readers.group_id = sqlc.arg(group_id)
  AND (
    sqlc.arg(search) = ''
    OR LOWER(users.email) LIKE '%' || LOWER(sqlc.arg(search)) || '%'
  )
ORDER BY LOWER(users.email) DESC
LIMIT sqlc.arg(limit) OFFSET sqlc.arg(offset);

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
INSERT INTO group_admins (id, user_id, group_id)
VALUES (?, ?, ?)
RETURNING *;

-- name: IsGroupAdmin :one
SELECT COUNT(*)
FROM group_admins
WHERE user_id = ?
  AND group_id = ?;

-- name: RemoveGroupAdmin :exec
DELETE FROM group_admins
WHERE user_id = ?
  AND group_id = ?;

-- name: ListGroupAdminUserIDs :many
SELECT user_id
FROM group_admins
WHERE group_id = ?;

-- name: CountGroupAdminsFiltered :one
SELECT COUNT(*) FROM (
  SELECT u.id
  FROM groups g
  JOIN users u ON u.id = g.admin_user_id
  WHERE g.id = sqlc.arg(group_id)
    AND (
      sqlc.arg(search) = ''
      OR LOWER(u.email) LIKE '%' || LOWER(sqlc.arg(search)) || '%'
    )
  UNION
  SELECT u.id
  FROM group_admins ga
  JOIN users u ON u.id = ga.user_id
  WHERE ga.group_id = sqlc.arg(group_id)
    AND (
      sqlc.arg(search) = ''
      OR LOWER(u.email) LIKE '%' || LOWER(sqlc.arg(search)) || '%'
    )
);

-- name: ListGroupAdminsByEmailAscFiltered :many
SELECT * FROM users
WHERE id IN (
  SELECT u.id
  FROM groups g
  JOIN users u ON u.id = g.admin_user_id
  WHERE g.id = sqlc.arg(group_id)
    AND (
      sqlc.arg(search) = ''
      OR LOWER(u.email) LIKE '%' || LOWER(sqlc.arg(search)) || '%'
    )
  UNION
  SELECT u.id
  FROM group_admins ga
  JOIN users u ON u.id = ga.user_id
  WHERE ga.group_id = sqlc.arg(group_id)
    AND (
      sqlc.arg(search) = ''
      OR LOWER(u.email) LIKE '%' || LOWER(sqlc.arg(search)) || '%'
    )
)
ORDER BY LOWER(email) ASC
LIMIT sqlc.arg(limit) OFFSET sqlc.arg(offset);

-- name: ListGroupAdminsByEmailDescFiltered :many
SELECT * FROM users
WHERE id IN (
  SELECT u.id
  FROM groups g
  JOIN users u ON u.id = g.admin_user_id
  WHERE g.id = sqlc.arg(group_id)
    AND (
      sqlc.arg(search) = ''
      OR LOWER(u.email) LIKE '%' || LOWER(sqlc.arg(search)) || '%'
    )
  UNION
  SELECT u.id
  FROM group_admins ga
  JOIN users u ON u.id = ga.user_id
  WHERE ga.group_id = sqlc.arg(group_id)
    AND (
      sqlc.arg(search) = ''
      OR LOWER(u.email) LIKE '%' || LOWER(sqlc.arg(search)) || '%'
    )
)
ORDER BY LOWER(email) DESC
LIMIT sqlc.arg(limit) OFFSET sqlc.arg(offset);

-- name: CountUserGroupsFiltered :one
SELECT COUNT(*) FROM (
  SELECT g.id FROM groups g
  JOIN users u ON u.id = g.admin_user_id
  WHERE g.admin_user_id = sqlc.arg(user_id)
    AND (
      sqlc.arg(search) = ''
      OR g.name LIKE '%' || sqlc.arg(search) || '%'
      OR u.email LIKE '%' || sqlc.arg(search) || '%'
    )
  UNION
  SELECT g.id FROM groups g
  JOIN group_admins ga ON ga.group_id = g.id
  JOIN users u ON u.id = g.admin_user_id
  WHERE ga.user_id = sqlc.arg(user_id)
    AND g.admin_user_id != sqlc.arg(user_id)
    AND (
      sqlc.arg(search) = ''
      OR g.name LIKE '%' || sqlc.arg(search) || '%'
      OR u.email LIKE '%' || sqlc.arg(search) || '%'
    )
  UNION
  SELECT g.id FROM groups g
  JOIN group_readers gr ON gr.group_id = g.id
  JOIN users u ON u.id = g.admin_user_id
  WHERE gr.user_id = sqlc.arg(user_id)
    AND g.admin_user_id != sqlc.arg(user_id)
  AND NOT EXISTS (
    SELECT 1 FROM group_admins ga
    WHERE ga.group_id = g.id
        AND ga.user_id = gr.user_id
    )
    AND (
      sqlc.arg(search) = ''
      OR g.name LIKE '%' || sqlc.arg(search) || '%'
      OR u.email LIKE '%' || sqlc.arg(search) || '%'
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
JOIN users u ON u.id = g.admin_user_id
WHERE g.admin_user_id = sqlc.arg(user_id)
  AND (
    sqlc.arg(search) = ''
    OR g.name LIKE '%' || sqlc.arg(search) || '%'
    OR u.email LIKE '%' || sqlc.arg(search) || '%'
  )
UNION ALL
SELECT
  g.id,
  g.name,
  g.admin_user_id,
  g.created_at,
  'admin' as role
FROM groups g
JOIN group_admins ga ON ga.group_id = g.id
JOIN users u ON u.id = g.admin_user_id
WHERE ga.user_id = sqlc.arg(user_id)
  AND g.admin_user_id != sqlc.arg(user_id)
  AND (
    sqlc.arg(search) = ''
    OR g.name LIKE '%' || sqlc.arg(search) || '%'
    OR u.email LIKE '%' || sqlc.arg(search) || '%'
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
JOIN users u ON u.id = g.admin_user_id
WHERE gr.user_id = sqlc.arg(user_id)
  AND g.admin_user_id != sqlc.arg(user_id)
  AND NOT EXISTS (
    SELECT 1 FROM group_admins ga
    WHERE ga.group_id = g.id
      AND ga.user_id = gr.user_id
  )
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
  CASE WHEN g.admin_user_id = sqlc.arg(user_id) THEN 'admin' ELSE 'viewer' END as role
FROM groups g
JOIN users u ON u.id = g.admin_user_id
WHERE g.admin_user_id = sqlc.arg(user_id)
  AND (
    sqlc.arg(search) = ''
    OR g.name LIKE '%' || sqlc.arg(search) || '%'
    OR u.email LIKE '%' || sqlc.arg(search) || '%'
  )
UNION ALL
SELECT
  g.id,
  g.name,
  g.admin_user_id,
  g.created_at,
  'admin' as role
FROM groups g
JOIN group_admins ga ON ga.group_id = g.id
JOIN users u ON u.id = g.admin_user_id
WHERE ga.user_id = sqlc.arg(user_id)
  AND g.admin_user_id != sqlc.arg(user_id)
  AND (
    sqlc.arg(search) = ''
    OR g.name LIKE '%' || sqlc.arg(search) || '%'
    OR u.email LIKE '%' || sqlc.arg(search) || '%'
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
JOIN users u ON u.id = g.admin_user_id
WHERE gr.user_id = sqlc.arg(user_id)
  AND g.admin_user_id != sqlc.arg(user_id)
  AND NOT EXISTS (
    SELECT 1 FROM group_admins ga
    WHERE ga.group_id = g.id
      AND ga.user_id = gr.user_id
  )
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
  CASE WHEN g.admin_user_id = sqlc.arg(user_id) THEN 'admin' ELSE 'viewer' END as role
FROM groups g
JOIN users u ON u.id = g.admin_user_id
WHERE g.admin_user_id = sqlc.arg(user_id)
  AND (
    sqlc.arg(search) = ''
    OR g.name LIKE '%' || sqlc.arg(search) || '%'
    OR u.email LIKE '%' || sqlc.arg(search) || '%'
  )
UNION ALL
SELECT
  g.id,
  g.name,
  g.admin_user_id,
  g.created_at,
  'admin' as role
FROM groups g
JOIN group_admins ga ON ga.group_id = g.id
JOIN users u ON u.id = g.admin_user_id
WHERE ga.user_id = sqlc.arg(user_id)
  AND g.admin_user_id != sqlc.arg(user_id)
  AND (
    sqlc.arg(search) = ''
    OR g.name LIKE '%' || sqlc.arg(search) || '%'
    OR u.email LIKE '%' || sqlc.arg(search) || '%'
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
JOIN users u ON u.id = g.admin_user_id
WHERE gr.user_id = sqlc.arg(user_id)
  AND g.admin_user_id != sqlc.arg(user_id)
  AND NOT EXISTS (
    SELECT 1 FROM group_admins ga
    WHERE ga.group_id = g.id
      AND ga.user_id = gr.user_id
  )
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
  CASE WHEN g.admin_user_id = sqlc.arg(user_id) THEN 'admin' ELSE 'viewer' END as role
FROM groups g
JOIN users u ON u.id = g.admin_user_id
WHERE g.admin_user_id = sqlc.arg(user_id)
  AND (
    sqlc.arg(search) = ''
    OR g.name LIKE '%' || sqlc.arg(search) || '%'
    OR u.email LIKE '%' || sqlc.arg(search) || '%'
  )
UNION ALL
SELECT
  g.id,
  g.name,
  g.admin_user_id,
  g.created_at,
  'admin' as role
FROM groups g
JOIN group_admins ga ON ga.group_id = g.id
JOIN users u ON u.id = g.admin_user_id
WHERE ga.user_id = sqlc.arg(user_id)
  AND g.admin_user_id != sqlc.arg(user_id)
  AND (
    sqlc.arg(search) = ''
    OR g.name LIKE '%' || sqlc.arg(search) || '%'
    OR u.email LIKE '%' || sqlc.arg(search) || '%'
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
JOIN users u ON u.id = g.admin_user_id
WHERE gr.user_id = sqlc.arg(user_id)
  AND g.admin_user_id != sqlc.arg(user_id)
  AND NOT EXISTS (
    SELECT 1 FROM group_admins ga
    WHERE ga.group_id = g.id
      AND ga.user_id = gr.user_id
  )
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
  CASE WHEN g.admin_user_id = sqlc.arg(user_id) THEN 'admin' ELSE 'viewer' END as role,
  u.email as admin_email
FROM groups g
JOIN users u ON u.id = g.admin_user_id
WHERE g.admin_user_id = sqlc.arg(user_id)
  AND (
    sqlc.arg(search) = ''
    OR g.name LIKE '%' || sqlc.arg(search) || '%'
    OR u.email LIKE '%' || sqlc.arg(search) || '%'
  )
UNION ALL
SELECT
  g.id,
  g.name,
  g.admin_user_id,
  g.created_at,
  'admin' as role,
  u.email as admin_email
FROM groups g
JOIN group_admins ga ON ga.group_id = g.id
JOIN users u ON u.id = g.admin_user_id
WHERE ga.user_id = sqlc.arg(user_id)
  AND g.admin_user_id != sqlc.arg(user_id)
  AND (
    sqlc.arg(search) = ''
    OR g.name LIKE '%' || sqlc.arg(search) || '%'
    OR u.email LIKE '%' || sqlc.arg(search) || '%'
  )
UNION ALL
SELECT 
  g.id,
  g.name,
  g.admin_user_id,
  g.created_at,
  'viewer' as role,
  u.email as admin_email
FROM groups g
JOIN group_readers gr ON gr.group_id = g.id
JOIN users u ON u.id = g.admin_user_id
WHERE gr.user_id = sqlc.arg(user_id)
  AND g.admin_user_id != sqlc.arg(user_id)
  AND NOT EXISTS (
    SELECT 1 FROM group_admins ga
    WHERE ga.group_id = g.id
      AND ga.user_id = gr.user_id
  )
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
  CASE WHEN g.admin_user_id = sqlc.arg(user_id) THEN 'admin' ELSE 'viewer' END as role,
  u.email as admin_email
FROM groups g
JOIN users u ON u.id = g.admin_user_id
WHERE g.admin_user_id = sqlc.arg(user_id)
  AND (
    sqlc.arg(search) = ''
    OR g.name LIKE '%' || sqlc.arg(search) || '%'
    OR u.email LIKE '%' || sqlc.arg(search) || '%'
  )
UNION ALL
SELECT
  g.id,
  g.name,
  g.admin_user_id,
  g.created_at,
  'admin' as role,
  u.email as admin_email
FROM groups g
JOIN group_admins ga ON ga.group_id = g.id
JOIN users u ON u.id = g.admin_user_id
WHERE ga.user_id = sqlc.arg(user_id)
  AND g.admin_user_id != sqlc.arg(user_id)
  AND (
    sqlc.arg(search) = ''
    OR g.name LIKE '%' || sqlc.arg(search) || '%'
    OR u.email LIKE '%' || sqlc.arg(search) || '%'
  )
UNION ALL
SELECT 
  g.id,
  g.name,
  g.admin_user_id,
  g.created_at,
  'viewer' as role,
  u.email as admin_email
FROM groups g
JOIN group_readers gr ON gr.group_id = g.id
JOIN users u ON u.id = g.admin_user_id
WHERE gr.user_id = sqlc.arg(user_id)
  AND g.admin_user_id != sqlc.arg(user_id)
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
