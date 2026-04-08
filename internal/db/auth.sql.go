// source: auth.sql

package db

import (
	"context"
	"database/sql"
	"time"
)

const banUser = `-- name: BanUser :exec
INSERT INTO banned_users (id, user_id)
VALUES (?, ?)
ON CONFLICT(user_id) DO NOTHING
`

type BanUserParams struct {
	ID     string `json:"id"`
	UserID string `json:"user_id"`
}

func (q *Queries) BanUser(ctx context.Context, arg BanUserParams) error {
	_, err := q.db.ExecContext(ctx, banUser, arg.ID, arg.UserID)
	return err
}

const countGroupAdminsFiltered = `-- name: CountGroupAdminsFiltered :one
SELECT COUNT(*)
FROM users u
JOIN group_access ga ON ga.user_id = u.id
WHERE ga.group_id = ?1
  AND ga.role IN ('owner', 'admin')
  AND (
    ?2 = ''
    OR LOWER(u.email) LIKE '%' || LOWER(?2) || '%'
  )
`

type CountGroupAdminsFilteredParams struct {
	GroupID string      `json:"group_id"`
	Search  interface{} `json:"search"`
}

func (q *Queries) CountGroupAdminsFiltered(ctx context.Context, arg CountGroupAdminsFilteredParams) (int64, error) {
	row := q.db.QueryRowContext(ctx, countGroupAdminsFiltered, arg.GroupID, arg.Search)
	var count int64
	err := row.Scan(&count)
	return count, err
}

const countGroupPendingInvitesFiltered = `-- name: CountGroupPendingInvitesFiltered :one
SELECT COUNT(*) FROM magic_links
WHERE action = 'invite'
  AND group_id = ?1
  AND used_at IS NULL
  AND (
    ?2 = ''
    OR LOWER(email) LIKE '%' || LOWER(?2) || '%'
  )
`

type CountGroupPendingInvitesFilteredParams struct {
	GroupID sql.NullString `json:"group_id"`
	Search  interface{}    `json:"search"`
}

func (q *Queries) CountGroupPendingInvitesFiltered(ctx context.Context, arg CountGroupPendingInvitesFilteredParams) (int64, error) {
	row := q.db.QueryRowContext(ctx, countGroupPendingInvitesFiltered, arg.GroupID, arg.Search)
	var count int64
	err := row.Scan(&count)
	return count, err
}

const countGroupReadersFiltered = `-- name: CountGroupReadersFiltered :one
SELECT COUNT(*) FROM users
JOIN group_access ON group_access.user_id = users.id
WHERE group_access.group_id = ?1
  AND group_access.role = 'viewer'
  AND (
    ?2 = ''
    OR LOWER(users.email) LIKE '%' || LOWER(?2) || '%'
  )
`

type CountGroupReadersFilteredParams struct {
	GroupID string      `json:"group_id"`
	Search  interface{} `json:"search"`
}

func (q *Queries) CountGroupReadersFiltered(ctx context.Context, arg CountGroupReadersFilteredParams) (int64, error) {
	row := q.db.QueryRowContext(ctx, countGroupReadersFiltered, arg.GroupID, arg.Search)
	var count int64
	err := row.Scan(&count)
	return count, err
}

const countSessionsFiltered = `-- name: CountSessionsFiltered :one
SELECT COUNT(*)
FROM user_sessions us
JOIN users u ON u.id = us.user_id
WHERE us.expires_at > CURRENT_TIMESTAMP
  AND (
    ?1 = ''
    OR u.email LIKE '%' || ?1 || '%'
    OR us.id LIKE '%' || ?1 || '%'
  )
`

func (q *Queries) CountSessionsFiltered(ctx context.Context, search interface{}) (int64, error) {
	row := q.db.QueryRowContext(ctx, countSessionsFiltered, search)
	var count int64
	err := row.Scan(&count)
	return count, err
}

const countUserGroupsFiltered = `-- name: CountUserGroupsFiltered :one
SELECT COUNT(*)
FROM groups g
JOIN group_access ga ON ga.group_id = g.id
JOIN users u ON u.id = g.admin_user_id
WHERE ga.user_id = ?1
  AND (
    ?2 = ''
    OR g.name LIKE '%' || ?2 || '%'
    OR u.email LIKE '%' || ?2 || '%'
  )
`

type CountUserGroupsFilteredParams struct {
	UserID string      `json:"user_id"`
	Search interface{} `json:"search"`
}

func (q *Queries) CountUserGroupsFiltered(ctx context.Context, arg CountUserGroupsFilteredParams) (int64, error) {
	row := q.db.QueryRowContext(ctx, countUserGroupsFiltered, arg.UserID, arg.Search)
	var count int64
	err := row.Scan(&count)
	return count, err
}

const createGroup = `-- name: CreateGroup :one
INSERT INTO groups (id, name, admin_user_id)
VALUES (?, ?, ?)
RETURNING id, name, admin_user_id, created_at
`

type CreateGroupParams struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	AdminUserID string `json:"admin_user_id"`
}

func (q *Queries) CreateGroup(ctx context.Context, arg CreateGroupParams) (Group, error) {
	row := q.db.QueryRowContext(ctx, createGroup, arg.ID, arg.Name, arg.AdminUserID)
	var i Group
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.AdminUserID,
		&i.CreatedAt,
	)
	return i, err
}

const createGroupAdmin = `-- name: CreateGroupAdmin :one
INSERT INTO group_access (id, user_id, group_id, role)
VALUES (?, ?, ?, 'admin')
ON CONFLICT(user_id, group_id) DO UPDATE SET role = 'admin'
WHERE group_access.role != 'owner'
RETURNING id, user_id, group_id, created_at
`

type CreateGroupAdminParams struct {
	ID      string `json:"id"`
	UserID  string `json:"user_id"`
	GroupID string `json:"group_id"`
}

type CreateGroupAdminRow struct {
	ID        string       `json:"id"`
	UserID    string       `json:"user_id"`
	GroupID   string       `json:"group_id"`
	CreatedAt sql.NullTime `json:"created_at"`
}

func (q *Queries) CreateGroupAdmin(ctx context.Context, arg CreateGroupAdminParams) (CreateGroupAdminRow, error) {
	row := q.db.QueryRowContext(ctx, createGroupAdmin, arg.ID, arg.UserID, arg.GroupID)
	var i CreateGroupAdminRow
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.GroupID,
		&i.CreatedAt,
	)
	return i, err
}

const createGroupReader = `-- name: CreateGroupReader :one
INSERT INTO group_access (id, user_id, group_id, role)
VALUES (?, ?, ?, 'viewer')
ON CONFLICT(user_id, group_id) DO UPDATE SET role = 'viewer'
WHERE group_access.role != 'owner'
RETURNING id, user_id, group_id, created_at
`

type CreateGroupReaderParams struct {
	ID      string `json:"id"`
	UserID  string `json:"user_id"`
	GroupID string `json:"group_id"`
}

type CreateGroupReaderRow struct {
	ID        string       `json:"id"`
	UserID    string       `json:"user_id"`
	GroupID   string       `json:"group_id"`
	CreatedAt sql.NullTime `json:"created_at"`
}

func (q *Queries) CreateGroupReader(ctx context.Context, arg CreateGroupReaderParams) (CreateGroupReaderRow, error) {
	row := q.db.QueryRowContext(ctx, createGroupReader, arg.ID, arg.UserID, arg.GroupID)
	var i CreateGroupReaderRow
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.GroupID,
		&i.CreatedAt,
	)
	return i, err
}

const createInviteMagicLink = `-- name: CreateInviteMagicLink :one
INSERT INTO magic_links (id, token, email, action, group_id, expires_at, invite_role)
VALUES (?, ?, ?, 'invite', ?, CURRENT_TIMESTAMP, ?)
RETURNING id, token, email, "action", group_id, expires_at, used_at, created_at, invite_role
`

type CreateInviteMagicLinkParams struct {
	ID         string         `json:"id"`
	Token      string         `json:"token"`
	Email      string         `json:"email"`
	GroupID    sql.NullString `json:"group_id"`
	InviteRole string         `json:"invite_role"`
}

func (q *Queries) CreateInviteMagicLink(ctx context.Context, arg CreateInviteMagicLinkParams) (MagicLink, error) {
	row := q.db.QueryRowContext(ctx, createInviteMagicLink,
		arg.ID,
		arg.Token,
		arg.Email,
		arg.GroupID,
		arg.InviteRole,
	)
	var i MagicLink
	err := row.Scan(
		&i.ID,
		&i.Token,
		&i.Email,
		&i.Action,
		&i.GroupID,
		&i.ExpiresAt,
		&i.UsedAt,
		&i.CreatedAt,
		&i.InviteRole,
	)
	return i, err
}

const createMagicLink = `-- name: CreateMagicLink :one
INSERT INTO magic_links (id, token, email, action, group_id, expires_at)
VALUES (?, ?, ?, ?, ?, ?)
RETURNING id, token, email, "action", group_id, expires_at, used_at, created_at, invite_role
`

type CreateMagicLinkParams struct {
	ID        string         `json:"id"`
	Token     string         `json:"token"`
	Email     string         `json:"email"`
	Action    string         `json:"action"`
	GroupID   sql.NullString `json:"group_id"`
	ExpiresAt time.Time      `json:"expires_at"`
}

func (q *Queries) CreateMagicLink(ctx context.Context, arg CreateMagicLinkParams) (MagicLink, error) {
	row := q.db.QueryRowContext(ctx, createMagicLink,
		arg.ID,
		arg.Token,
		arg.Email,
		arg.Action,
		arg.GroupID,
		arg.ExpiresAt,
	)
	var i MagicLink
	err := row.Scan(
		&i.ID,
		&i.Token,
		&i.Email,
		&i.Action,
		&i.GroupID,
		&i.ExpiresAt,
		&i.UsedAt,
		&i.CreatedAt,
		&i.InviteRole,
	)
	return i, err
}

const createUser = `-- name: CreateUser :one
INSERT INTO users (id, email, preferred_lang)
VALUES (?1, ?2, COALESCE(NULLIF(?3, ''), 'hu'))
RETURNING id, email, created_at, preferred_lang
`

type CreateUserParams struct {
	ID            string      `json:"id"`
	Email         string      `json:"email"`
	PreferredLang interface{} `json:"preferred_lang"`
}

func (q *Queries) CreateUser(ctx context.Context, arg CreateUserParams) (User, error) {
	row := q.db.QueryRowContext(ctx, createUser, arg.ID, arg.Email, arg.PreferredLang)
	var i User
	err := row.Scan(
		&i.ID,
		&i.Email,
		&i.CreatedAt,
		&i.PreferredLang,
	)
	return i, err
}

const createUserSession = `-- name: CreateUserSession :one
INSERT INTO user_sessions (id, user_id, token, expires_at)
VALUES (?, ?, ?, ?)
RETURNING id, user_id, token, created_at, expires_at
`

type CreateUserSessionParams struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}

func (q *Queries) CreateUserSession(ctx context.Context, arg CreateUserSessionParams) (UserSession, error) {
	row := q.db.QueryRowContext(ctx, createUserSession,
		arg.ID,
		arg.UserID,
		arg.Token,
		arg.ExpiresAt,
	)
	var i UserSession
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.Token,
		&i.CreatedAt,
		&i.ExpiresAt,
	)
	return i, err
}

const deleteAllUserSessions = `-- name: DeleteAllUserSessions :exec
DELETE FROM user_sessions
WHERE user_id = ?
`

func (q *Queries) DeleteAllUserSessions(ctx context.Context, userID string) error {
	_, err := q.db.ExecContext(ctx, deleteAllUserSessions, userID)
	return err
}

const deleteExpiredMagicLinks = `-- name: DeleteExpiredMagicLinks :exec
DELETE FROM magic_links
WHERE action != 'invite'
  AND expires_at < CURRENT_TIMESTAMP
  AND used_at IS NULL
`

func (q *Queries) DeleteExpiredMagicLinks(ctx context.Context) error {
	_, err := q.db.ExecContext(ctx, deleteExpiredMagicLinks)
	return err
}

const deleteExpiredUserSessions = `-- name: DeleteExpiredUserSessions :exec
DELETE FROM user_sessions
WHERE expires_at < CURRENT_TIMESTAMP
`

func (q *Queries) DeleteExpiredUserSessions(ctx context.Context) error {
	_, err := q.db.ExecContext(ctx, deleteExpiredUserSessions)
	return err
}

const deleteGroup = `-- name: DeleteGroup :exec
DELETE FROM groups
WHERE id = ?
`

func (q *Queries) DeleteGroup(ctx context.Context, id string) error {
	_, err := q.db.ExecContext(ctx, deleteGroup, id)
	return err
}

const deleteGroupPendingInvite = `-- name: DeleteGroupPendingInvite :exec
DELETE FROM magic_links
WHERE id = ?
  AND action = 'invite'
  AND group_id = ?
  AND used_at IS NULL
`

type DeleteGroupPendingInviteParams struct {
	ID      string         `json:"id"`
	GroupID sql.NullString `json:"group_id"`
}

func (q *Queries) DeleteGroupPendingInvite(ctx context.Context, arg DeleteGroupPendingInviteParams) error {
	_, err := q.db.ExecContext(ctx, deleteGroupPendingInvite, arg.ID, arg.GroupID)
	return err
}

const deleteOtherUserSessions = `-- name: DeleteOtherUserSessions :exec
DELETE FROM user_sessions
WHERE user_id = ? AND id != ?
`

type DeleteOtherUserSessionsParams struct {
	UserID string `json:"user_id"`
	ID     string `json:"id"`
}

func (q *Queries) DeleteOtherUserSessions(ctx context.Context, arg DeleteOtherUserSessionsParams) error {
	_, err := q.db.ExecContext(ctx, deleteOtherUserSessions, arg.UserID, arg.ID)
	return err
}

const deleteUserSession = `-- name: DeleteUserSession :exec
DELETE FROM user_sessions
WHERE id = ? AND user_id = ?
`

type DeleteUserSessionParams struct {
	ID     string `json:"id"`
	UserID string `json:"user_id"`
}

func (q *Queries) DeleteUserSession(ctx context.Context, arg DeleteUserSessionParams) error {
	_, err := q.db.ExecContext(ctx, deleteUserSession, arg.ID, arg.UserID)
	return err
}

const deleteUserSessionByID = `-- name: DeleteUserSessionByID :exec
DELETE FROM user_sessions
WHERE id = ?
`

func (q *Queries) DeleteUserSessionByID(ctx context.Context, id string) error {
	_, err := q.db.ExecContext(ctx, deleteUserSessionByID, id)
	return err
}

const getGroupAccessRole = `-- name: GetGroupAccessRole :one
SELECT role
FROM group_access
WHERE user_id = ?
  AND group_id = ?
`

type GetGroupAccessRoleParams struct {
	UserID  string `json:"user_id"`
	GroupID string `json:"group_id"`
}

func (q *Queries) GetGroupAccessRole(ctx context.Context, arg GetGroupAccessRoleParams) (string, error) {
	row := q.db.QueryRowContext(ctx, getGroupAccessRole, arg.UserID, arg.GroupID)
	var role string
	err := row.Scan(&role)
	return role, err
}

const getGroupByAdmin = `-- name: GetGroupByAdmin :one
SELECT g.id, g.name, g.admin_user_id, g.created_at
FROM groups g
JOIN group_access ga ON ga.group_id = g.id
WHERE ga.user_id = ?1
  AND ga.role IN ('owner', 'admin')
LIMIT 1
`

func (q *Queries) GetGroupByAdmin(ctx context.Context, userID string) (Group, error) {
	row := q.db.QueryRowContext(ctx, getGroupByAdmin, userID)
	var i Group
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.AdminUserID,
		&i.CreatedAt,
	)
	return i, err
}

const getGroupByID = `-- name: GetGroupByID :one
SELECT id, name, admin_user_id, created_at FROM groups
WHERE id = ?
`

func (q *Queries) GetGroupByID(ctx context.Context, id string) (Group, error) {
	row := q.db.QueryRowContext(ctx, getGroupByID, id)
	var i Group
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.AdminUserID,
		&i.CreatedAt,
	)
	return i, err
}

const getGroupReaders = `-- name: GetGroupReaders :many
SELECT users.id, users.email, users.created_at, users.preferred_lang FROM users
JOIN group_access ON group_access.user_id = users.id
WHERE group_access.group_id = ?
  AND group_access.role = 'viewer'
`

func (q *Queries) GetGroupReaders(ctx context.Context, groupID string) ([]User, error) {
	rows, err := q.db.QueryContext(ctx, getGroupReaders, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []User{}
	for rows.Next() {
		var i User
		if err := rows.Scan(
			&i.ID,
			&i.Email,
			&i.CreatedAt,
			&i.PreferredLang,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getMagicLinkByToken = `-- name: GetMagicLinkByToken :one
SELECT id, token, email, "action", group_id, expires_at, used_at, created_at, invite_role FROM magic_links
WHERE token = ?
`

func (q *Queries) GetMagicLinkByToken(ctx context.Context, token string) (MagicLink, error) {
	row := q.db.QueryRowContext(ctx, getMagicLinkByToken, token)
	var i MagicLink
	err := row.Scan(
		&i.ID,
		&i.Token,
		&i.Email,
		&i.Action,
		&i.GroupID,
		&i.ExpiresAt,
		&i.UsedAt,
		&i.CreatedAt,
		&i.InviteRole,
	)
	return i, err
}

const getUserByEmail = `-- name: GetUserByEmail :one
SELECT id, email, created_at, preferred_lang FROM users
WHERE email = ?
`

func (q *Queries) GetUserByEmail(ctx context.Context, email string) (User, error) {
	row := q.db.QueryRowContext(ctx, getUserByEmail, email)
	var i User
	err := row.Scan(
		&i.ID,
		&i.Email,
		&i.CreatedAt,
		&i.PreferredLang,
	)
	return i, err
}

const getUserByID = `-- name: GetUserByID :one
SELECT id, email, created_at, preferred_lang FROM users
WHERE id = ?
`

func (q *Queries) GetUserByID(ctx context.Context, id string) (User, error) {
	row := q.db.QueryRowContext(ctx, getUserByID, id)
	var i User
	err := row.Scan(
		&i.ID,
		&i.Email,
		&i.CreatedAt,
		&i.PreferredLang,
	)
	return i, err
}

const getUserSessionByToken = `-- name: GetUserSessionByToken :one
SELECT id, user_id, token, created_at, expires_at FROM user_sessions
WHERE token = ? AND expires_at > CURRENT_TIMESTAMP
`

func (q *Queries) GetUserSessionByToken(ctx context.Context, token string) (UserSession, error) {
	row := q.db.QueryRowContext(ctx, getUserSessionByToken, token)
	var i UserSession
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.Token,
		&i.CreatedAt,
		&i.ExpiresAt,
	)
	return i, err
}

const isGroupAdmin = `-- name: IsGroupAdmin :one
SELECT COUNT(*)
FROM group_access
WHERE user_id = ?
  AND group_id = ?
  AND role = 'admin'
`

type IsGroupAdminParams struct {
	UserID  string `json:"user_id"`
	GroupID string `json:"group_id"`
}

func (q *Queries) IsGroupAdmin(ctx context.Context, arg IsGroupAdminParams) (int64, error) {
	row := q.db.QueryRowContext(ctx, isGroupAdmin, arg.UserID, arg.GroupID)
	var count int64
	err := row.Scan(&count)
	return count, err
}

const isGroupReader = `-- name: IsGroupReader :one
SELECT COUNT(*) FROM group_access
WHERE user_id = ?
  AND group_id = ?
  AND role = 'viewer'
`

type IsGroupReaderParams struct {
	UserID  string `json:"user_id"`
	GroupID string `json:"group_id"`
}

func (q *Queries) IsGroupReader(ctx context.Context, arg IsGroupReaderParams) (int64, error) {
	row := q.db.QueryRowContext(ctx, isGroupReader, arg.UserID, arg.GroupID)
	var count int64
	err := row.Scan(&count)
	return count, err
}

const isUserBanned = `-- name: IsUserBanned :one
SELECT COUNT(*) FROM banned_users
WHERE user_id = ?
`

func (q *Queries) IsUserBanned(ctx context.Context, userID string) (int64, error) {
	row := q.db.QueryRowContext(ctx, isUserBanned, userID)
	var count int64
	err := row.Scan(&count)
	return count, err
}

const listGroupAdminUserIDs = `-- name: ListGroupAdminUserIDs :many
SELECT user_id
FROM group_access
WHERE group_id = ?
  AND role = 'admin'
`

func (q *Queries) ListGroupAdminUserIDs(ctx context.Context, groupID string) ([]string, error) {
	rows, err := q.db.QueryContext(ctx, listGroupAdminUserIDs, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []string{}
	for rows.Next() {
		var user_id string
		if err := rows.Scan(&user_id); err != nil {
			return nil, err
		}
		items = append(items, user_id)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const listGroupAdmins = `-- name: ListGroupAdmins :many
SELECT u.id, u.email, u.created_at, u.preferred_lang
FROM users u
JOIN group_access ga ON ga.user_id = u.id
WHERE ga.group_id = ?
  AND ga.role IN ('owner', 'admin')
ORDER BY LOWER(u.email) ASC
`

func (q *Queries) ListGroupAdmins(ctx context.Context, groupID string) ([]User, error) {
	rows, err := q.db.QueryContext(ctx, listGroupAdmins, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []User{}
	for rows.Next() {
		var i User
		if err := rows.Scan(
			&i.ID,
			&i.Email,
			&i.CreatedAt,
			&i.PreferredLang,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const listGroupAdminsByEmailAscFiltered = `-- name: ListGroupAdminsByEmailAscFiltered :many
SELECT u.id, u.email, u.created_at, u.preferred_lang
FROM users u
JOIN group_access ga ON ga.user_id = u.id
WHERE ga.group_id = ?1
  AND ga.role IN ('owner', 'admin')
  AND (
    ?2 = ''
    OR LOWER(u.email) LIKE '%' || LOWER(?2) || '%'
  )
ORDER BY LOWER(email) ASC
LIMIT ?4 OFFSET ?3
`

type ListGroupAdminsByEmailAscFilteredParams struct {
	GroupID string      `json:"group_id"`
	Search  interface{} `json:"search"`
	Offset  int64       `json:"offset"`
	Limit   int64       `json:"limit"`
}

func (q *Queries) ListGroupAdminsByEmailAscFiltered(ctx context.Context, arg ListGroupAdminsByEmailAscFilteredParams) ([]User, error) {
	rows, err := q.db.QueryContext(ctx, listGroupAdminsByEmailAscFiltered,
		arg.GroupID,
		arg.Search,
		arg.Offset,
		arg.Limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []User{}
	for rows.Next() {
		var i User
		if err := rows.Scan(
			&i.ID,
			&i.Email,
			&i.CreatedAt,
			&i.PreferredLang,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const listGroupAdminsByEmailDescFiltered = `-- name: ListGroupAdminsByEmailDescFiltered :many
SELECT u.id, u.email, u.created_at, u.preferred_lang
FROM users u
JOIN group_access ga ON ga.user_id = u.id
WHERE ga.group_id = ?1
  AND ga.role IN ('owner', 'admin')
  AND (
    ?2 = ''
    OR LOWER(u.email) LIKE '%' || LOWER(?2) || '%'
  )
ORDER BY LOWER(email) DESC
LIMIT ?4 OFFSET ?3
`

type ListGroupAdminsByEmailDescFilteredParams struct {
	GroupID string      `json:"group_id"`
	Search  interface{} `json:"search"`
	Offset  int64       `json:"offset"`
	Limit   int64       `json:"limit"`
}

func (q *Queries) ListGroupAdminsByEmailDescFiltered(ctx context.Context, arg ListGroupAdminsByEmailDescFilteredParams) ([]User, error) {
	rows, err := q.db.QueryContext(ctx, listGroupAdminsByEmailDescFiltered,
		arg.GroupID,
		arg.Search,
		arg.Offset,
		arg.Limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []User{}
	for rows.Next() {
		var i User
		if err := rows.Scan(
			&i.ID,
			&i.Email,
			&i.CreatedAt,
			&i.PreferredLang,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const listGroupPendingInvites = `-- name: ListGroupPendingInvites :many
SELECT id, token, email, "action", group_id, expires_at, used_at, created_at, invite_role FROM magic_links
WHERE action = 'invite'
  AND group_id = ?
  AND used_at IS NULL
ORDER BY created_at DESC
`

func (q *Queries) ListGroupPendingInvites(ctx context.Context, groupID sql.NullString) ([]MagicLink, error) {
	rows, err := q.db.QueryContext(ctx, listGroupPendingInvites, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []MagicLink{}
	for rows.Next() {
		var i MagicLink
		if err := rows.Scan(
			&i.ID,
			&i.Token,
			&i.Email,
			&i.Action,
			&i.GroupID,
			&i.ExpiresAt,
			&i.UsedAt,
			&i.CreatedAt,
			&i.InviteRole,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const listGroupPendingInvitesByCreatedAscFiltered = `-- name: ListGroupPendingInvitesByCreatedAscFiltered :many
SELECT id, token, email, "action", group_id, expires_at, used_at, created_at, invite_role FROM magic_links
WHERE action = 'invite'
  AND group_id = ?1
  AND used_at IS NULL
  AND (
    ?2 = ''
    OR LOWER(email) LIKE '%' || LOWER(?2) || '%'
  )
ORDER BY created_at ASC, LOWER(email) ASC
LIMIT ?4 OFFSET ?3
`

type ListGroupPendingInvitesByCreatedAscFilteredParams struct {
	GroupID sql.NullString `json:"group_id"`
	Search  interface{}    `json:"search"`
	Offset  int64          `json:"offset"`
	Limit   int64          `json:"limit"`
}

func (q *Queries) ListGroupPendingInvitesByCreatedAscFiltered(ctx context.Context, arg ListGroupPendingInvitesByCreatedAscFilteredParams) ([]MagicLink, error) {
	rows, err := q.db.QueryContext(ctx, listGroupPendingInvitesByCreatedAscFiltered,
		arg.GroupID,
		arg.Search,
		arg.Offset,
		arg.Limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []MagicLink{}
	for rows.Next() {
		var i MagicLink
		if err := rows.Scan(
			&i.ID,
			&i.Token,
			&i.Email,
			&i.Action,
			&i.GroupID,
			&i.ExpiresAt,
			&i.UsedAt,
			&i.CreatedAt,
			&i.InviteRole,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const listGroupPendingInvitesByCreatedDescFiltered = `-- name: ListGroupPendingInvitesByCreatedDescFiltered :many
SELECT id, token, email, "action", group_id, expires_at, used_at, created_at, invite_role FROM magic_links
WHERE action = 'invite'
  AND group_id = ?1
  AND used_at IS NULL
  AND (
    ?2 = ''
    OR LOWER(email) LIKE '%' || LOWER(?2) || '%'
  )
ORDER BY created_at DESC, LOWER(email) ASC
LIMIT ?4 OFFSET ?3
`

type ListGroupPendingInvitesByCreatedDescFilteredParams struct {
	GroupID sql.NullString `json:"group_id"`
	Search  interface{}    `json:"search"`
	Offset  int64          `json:"offset"`
	Limit   int64          `json:"limit"`
}

func (q *Queries) ListGroupPendingInvitesByCreatedDescFiltered(ctx context.Context, arg ListGroupPendingInvitesByCreatedDescFilteredParams) ([]MagicLink, error) {
	rows, err := q.db.QueryContext(ctx, listGroupPendingInvitesByCreatedDescFiltered,
		arg.GroupID,
		arg.Search,
		arg.Offset,
		arg.Limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []MagicLink{}
	for rows.Next() {
		var i MagicLink
		if err := rows.Scan(
			&i.ID,
			&i.Token,
			&i.Email,
			&i.Action,
			&i.GroupID,
			&i.ExpiresAt,
			&i.UsedAt,
			&i.CreatedAt,
			&i.InviteRole,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const listGroupPendingInvitesByEmailAscFiltered = `-- name: ListGroupPendingInvitesByEmailAscFiltered :many
SELECT id, token, email, "action", group_id, expires_at, used_at, created_at, invite_role FROM magic_links
WHERE action = 'invite'
  AND group_id = ?1
  AND used_at IS NULL
  AND (
    ?2 = ''
    OR LOWER(email) LIKE '%' || LOWER(?2) || '%'
  )
ORDER BY LOWER(email) ASC, created_at DESC
LIMIT ?4 OFFSET ?3
`

type ListGroupPendingInvitesByEmailAscFilteredParams struct {
	GroupID sql.NullString `json:"group_id"`
	Search  interface{}    `json:"search"`
	Offset  int64          `json:"offset"`
	Limit   int64          `json:"limit"`
}

func (q *Queries) ListGroupPendingInvitesByEmailAscFiltered(ctx context.Context, arg ListGroupPendingInvitesByEmailAscFilteredParams) ([]MagicLink, error) {
	rows, err := q.db.QueryContext(ctx, listGroupPendingInvitesByEmailAscFiltered,
		arg.GroupID,
		arg.Search,
		arg.Offset,
		arg.Limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []MagicLink{}
	for rows.Next() {
		var i MagicLink
		if err := rows.Scan(
			&i.ID,
			&i.Token,
			&i.Email,
			&i.Action,
			&i.GroupID,
			&i.ExpiresAt,
			&i.UsedAt,
			&i.CreatedAt,
			&i.InviteRole,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const listGroupPendingInvitesByEmailDescFiltered = `-- name: ListGroupPendingInvitesByEmailDescFiltered :many
SELECT id, token, email, "action", group_id, expires_at, used_at, created_at, invite_role FROM magic_links
WHERE action = 'invite'
  AND group_id = ?1
  AND used_at IS NULL
  AND (
    ?2 = ''
    OR LOWER(email) LIKE '%' || LOWER(?2) || '%'
  )
ORDER BY LOWER(email) DESC, created_at DESC
LIMIT ?4 OFFSET ?3
`

type ListGroupPendingInvitesByEmailDescFilteredParams struct {
	GroupID sql.NullString `json:"group_id"`
	Search  interface{}    `json:"search"`
	Offset  int64          `json:"offset"`
	Limit   int64          `json:"limit"`
}

func (q *Queries) ListGroupPendingInvitesByEmailDescFiltered(ctx context.Context, arg ListGroupPendingInvitesByEmailDescFilteredParams) ([]MagicLink, error) {
	rows, err := q.db.QueryContext(ctx, listGroupPendingInvitesByEmailDescFiltered,
		arg.GroupID,
		arg.Search,
		arg.Offset,
		arg.Limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []MagicLink{}
	for rows.Next() {
		var i MagicLink
		if err := rows.Scan(
			&i.ID,
			&i.Token,
			&i.Email,
			&i.Action,
			&i.GroupID,
			&i.ExpiresAt,
			&i.UsedAt,
			&i.CreatedAt,
			&i.InviteRole,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const listGroupReadersByEmailAscFiltered = `-- name: ListGroupReadersByEmailAscFiltered :many
SELECT users.id, users.email, users.created_at, users.preferred_lang FROM users
JOIN group_access ON group_access.user_id = users.id
WHERE group_access.group_id = ?1
  AND group_access.role = 'viewer'
  AND (
    ?2 = ''
    OR LOWER(users.email) LIKE '%' || LOWER(?2) || '%'
  )
ORDER BY LOWER(users.email) ASC
LIMIT ?4 OFFSET ?3
`

type ListGroupReadersByEmailAscFilteredParams struct {
	GroupID string      `json:"group_id"`
	Search  interface{} `json:"search"`
	Offset  int64       `json:"offset"`
	Limit   int64       `json:"limit"`
}

func (q *Queries) ListGroupReadersByEmailAscFiltered(ctx context.Context, arg ListGroupReadersByEmailAscFilteredParams) ([]User, error) {
	rows, err := q.db.QueryContext(ctx, listGroupReadersByEmailAscFiltered,
		arg.GroupID,
		arg.Search,
		arg.Offset,
		arg.Limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []User{}
	for rows.Next() {
		var i User
		if err := rows.Scan(
			&i.ID,
			&i.Email,
			&i.CreatedAt,
			&i.PreferredLang,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const listGroupReadersByEmailDescFiltered = `-- name: ListGroupReadersByEmailDescFiltered :many
SELECT users.id, users.email, users.created_at, users.preferred_lang FROM users
JOIN group_access ON group_access.user_id = users.id
WHERE group_access.group_id = ?1
  AND group_access.role = 'viewer'
  AND (
    ?2 = ''
    OR LOWER(users.email) LIKE '%' || LOWER(?2) || '%'
  )
ORDER BY LOWER(users.email) DESC
LIMIT ?4 OFFSET ?3
`

type ListGroupReadersByEmailDescFilteredParams struct {
	GroupID string      `json:"group_id"`
	Search  interface{} `json:"search"`
	Offset  int64       `json:"offset"`
	Limit   int64       `json:"limit"`
}

func (q *Queries) ListGroupReadersByEmailDescFiltered(ctx context.Context, arg ListGroupReadersByEmailDescFilteredParams) ([]User, error) {
	rows, err := q.db.QueryContext(ctx, listGroupReadersByEmailDescFiltered,
		arg.GroupID,
		arg.Search,
		arg.Offset,
		arg.Limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []User{}
	for rows.Next() {
		var i User
		if err := rows.Scan(
			&i.ID,
			&i.Email,
			&i.CreatedAt,
			&i.PreferredLang,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const listGroupUserAccess = `-- name: ListGroupUserAccess :many
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
ORDER BY group_access.created_at ASC, LOWER(users.email) ASC
`

type ListGroupUserAccessRow struct {
	ID              string       `json:"id"`
	Email           string       `json:"email"`
	CreatedAt       sql.NullTime `json:"created_at"`
	PreferredLang   string       `json:"preferred_lang"`
	Role            string       `json:"role"`
	AccessCreatedAt sql.NullTime `json:"access_created_at"`
}

func (q *Queries) ListGroupUserAccess(ctx context.Context, groupID string) ([]ListGroupUserAccessRow, error) {
	rows, err := q.db.QueryContext(ctx, listGroupUserAccess, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []ListGroupUserAccessRow{}
	for rows.Next() {
		var i ListGroupUserAccessRow
		if err := rows.Scan(
			&i.ID,
			&i.Email,
			&i.CreatedAt,
			&i.PreferredLang,
			&i.Role,
			&i.AccessCreatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const listGroupsByAdmin = `-- name: ListGroupsByAdmin :many
SELECT g.id, g.name, g.admin_user_id, g.created_at
FROM groups g
JOIN group_access ga ON ga.group_id = g.id
WHERE ga.user_id = ?1
  AND ga.role IN ('owner', 'admin')
ORDER BY created_at DESC
`

func (q *Queries) ListGroupsByAdmin(ctx context.Context, userID string) ([]Group, error) {
	rows, err := q.db.QueryContext(ctx, listGroupsByAdmin, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Group{}
	for rows.Next() {
		var i Group
		if err := rows.Scan(
			&i.ID,
			&i.Name,
			&i.AdminUserID,
			&i.CreatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const listGroupsByReader = `-- name: ListGroupsByReader :many
SELECT g.id, g.name, g.admin_user_id, g.created_at
FROM groups g
JOIN group_access ga ON ga.group_id = g.id
WHERE ga.user_id = ?1
  AND ga.role = 'viewer'
ORDER BY g.created_at DESC
`

func (q *Queries) ListGroupsByReader(ctx context.Context, userID string) ([]Group, error) {
	rows, err := q.db.QueryContext(ctx, listGroupsByReader, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Group{}
	for rows.Next() {
		var i Group
		if err := rows.Scan(
			&i.ID,
			&i.Name,
			&i.AdminUserID,
			&i.CreatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const listSessionsByCreatedAscFiltered = `-- name: ListSessionsByCreatedAscFiltered :many
SELECT us.id, us.user_id, u.email AS user_email, us.created_at, us.expires_at
FROM user_sessions us
JOIN users u ON u.id = us.user_id
WHERE us.expires_at > CURRENT_TIMESTAMP
  AND (
    ?1 = ''
    OR u.email LIKE '%' || ?1 || '%'
    OR us.id LIKE '%' || ?1 || '%'
  )
ORDER BY us.created_at ASC, u.email COLLATE NOCASE ASC
LIMIT ?3 OFFSET ?2
`

type ListSessionsByCreatedAscFilteredParams struct {
	Search interface{} `json:"search"`
	Offset int64       `json:"offset"`
	Limit  int64       `json:"limit"`
}

type ListSessionsByCreatedAscFilteredRow struct {
	ID        string       `json:"id"`
	UserID    string       `json:"user_id"`
	UserEmail string       `json:"user_email"`
	CreatedAt sql.NullTime `json:"created_at"`
	ExpiresAt time.Time    `json:"expires_at"`
}

func (q *Queries) ListSessionsByCreatedAscFiltered(ctx context.Context, arg ListSessionsByCreatedAscFilteredParams) ([]ListSessionsByCreatedAscFilteredRow, error) {
	rows, err := q.db.QueryContext(ctx, listSessionsByCreatedAscFiltered, arg.Search, arg.Offset, arg.Limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []ListSessionsByCreatedAscFilteredRow{}
	for rows.Next() {
		var i ListSessionsByCreatedAscFilteredRow
		if err := rows.Scan(
			&i.ID,
			&i.UserID,
			&i.UserEmail,
			&i.CreatedAt,
			&i.ExpiresAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const listSessionsByCreatedDescFiltered = `-- name: ListSessionsByCreatedDescFiltered :many
SELECT us.id, us.user_id, u.email AS user_email, us.created_at, us.expires_at
FROM user_sessions us
JOIN users u ON u.id = us.user_id
WHERE us.expires_at > CURRENT_TIMESTAMP
  AND (
    ?1 = ''
    OR u.email LIKE '%' || ?1 || '%'
    OR us.id LIKE '%' || ?1 || '%'
  )
ORDER BY us.created_at DESC, u.email COLLATE NOCASE ASC
LIMIT ?3 OFFSET ?2
`

type ListSessionsByCreatedDescFilteredParams struct {
	Search interface{} `json:"search"`
	Offset int64       `json:"offset"`
	Limit  int64       `json:"limit"`
}

type ListSessionsByCreatedDescFilteredRow struct {
	ID        string       `json:"id"`
	UserID    string       `json:"user_id"`
	UserEmail string       `json:"user_email"`
	CreatedAt sql.NullTime `json:"created_at"`
	ExpiresAt time.Time    `json:"expires_at"`
}

func (q *Queries) ListSessionsByCreatedDescFiltered(ctx context.Context, arg ListSessionsByCreatedDescFilteredParams) ([]ListSessionsByCreatedDescFilteredRow, error) {
	rows, err := q.db.QueryContext(ctx, listSessionsByCreatedDescFiltered, arg.Search, arg.Offset, arg.Limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []ListSessionsByCreatedDescFilteredRow{}
	for rows.Next() {
		var i ListSessionsByCreatedDescFilteredRow
		if err := rows.Scan(
			&i.ID,
			&i.UserID,
			&i.UserEmail,
			&i.CreatedAt,
			&i.ExpiresAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const listSessionsByEmailAscFiltered = `-- name: ListSessionsByEmailAscFiltered :many
SELECT us.id, us.user_id, u.email AS user_email, us.created_at, us.expires_at
FROM user_sessions us
JOIN users u ON u.id = us.user_id
WHERE us.expires_at > CURRENT_TIMESTAMP
  AND (
    ?1 = ''
    OR u.email LIKE '%' || ?1 || '%'
    OR us.id LIKE '%' || ?1 || '%'
  )
ORDER BY u.email COLLATE NOCASE ASC, us.created_at DESC
LIMIT ?3 OFFSET ?2
`

type ListSessionsByEmailAscFilteredParams struct {
	Search interface{} `json:"search"`
	Offset int64       `json:"offset"`
	Limit  int64       `json:"limit"`
}

type ListSessionsByEmailAscFilteredRow struct {
	ID        string       `json:"id"`
	UserID    string       `json:"user_id"`
	UserEmail string       `json:"user_email"`
	CreatedAt sql.NullTime `json:"created_at"`
	ExpiresAt time.Time    `json:"expires_at"`
}

func (q *Queries) ListSessionsByEmailAscFiltered(ctx context.Context, arg ListSessionsByEmailAscFilteredParams) ([]ListSessionsByEmailAscFilteredRow, error) {
	rows, err := q.db.QueryContext(ctx, listSessionsByEmailAscFiltered, arg.Search, arg.Offset, arg.Limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []ListSessionsByEmailAscFilteredRow{}
	for rows.Next() {
		var i ListSessionsByEmailAscFilteredRow
		if err := rows.Scan(
			&i.ID,
			&i.UserID,
			&i.UserEmail,
			&i.CreatedAt,
			&i.ExpiresAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const listSessionsByEmailDescFiltered = `-- name: ListSessionsByEmailDescFiltered :many
SELECT us.id, us.user_id, u.email AS user_email, us.created_at, us.expires_at
FROM user_sessions us
JOIN users u ON u.id = us.user_id
WHERE us.expires_at > CURRENT_TIMESTAMP
  AND (
    ?1 = ''
    OR u.email LIKE '%' || ?1 || '%'
    OR us.id LIKE '%' || ?1 || '%'
  )
ORDER BY u.email COLLATE NOCASE DESC, us.created_at DESC
LIMIT ?3 OFFSET ?2
`

type ListSessionsByEmailDescFilteredParams struct {
	Search interface{} `json:"search"`
	Offset int64       `json:"offset"`
	Limit  int64       `json:"limit"`
}

type ListSessionsByEmailDescFilteredRow struct {
	ID        string       `json:"id"`
	UserID    string       `json:"user_id"`
	UserEmail string       `json:"user_email"`
	CreatedAt sql.NullTime `json:"created_at"`
	ExpiresAt time.Time    `json:"expires_at"`
}

func (q *Queries) ListSessionsByEmailDescFiltered(ctx context.Context, arg ListSessionsByEmailDescFilteredParams) ([]ListSessionsByEmailDescFilteredRow, error) {
	rows, err := q.db.QueryContext(ctx, listSessionsByEmailDescFiltered, arg.Search, arg.Offset, arg.Limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []ListSessionsByEmailDescFilteredRow{}
	for rows.Next() {
		var i ListSessionsByEmailDescFilteredRow
		if err := rows.Scan(
			&i.ID,
			&i.UserID,
			&i.UserEmail,
			&i.CreatedAt,
			&i.ExpiresAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const listUserGroupsByAdminAscFiltered = `-- name: ListUserGroupsByAdminAscFiltered :many
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
WHERE ga.user_id = ?1
  AND (
    ?2 = ''
    OR g.name LIKE '%' || ?2 || '%'
    OR u.email LIKE '%' || ?2 || '%'
  )
ORDER BY admin_email COLLATE NOCASE ASC, g.name COLLATE NOCASE ASC
LIMIT ?4 OFFSET ?3
`

type ListUserGroupsByAdminAscFilteredParams struct {
	UserID string      `json:"user_id"`
	Search interface{} `json:"search"`
	Offset int64       `json:"offset"`
	Limit  int64       `json:"limit"`
}

type ListUserGroupsByAdminAscFilteredRow struct {
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	AdminUserID string       `json:"admin_user_id"`
	CreatedAt   sql.NullTime `json:"created_at"`
	Role        string       `json:"role"`
	AdminEmail  string       `json:"admin_email"`
}

func (q *Queries) ListUserGroupsByAdminAscFiltered(ctx context.Context, arg ListUserGroupsByAdminAscFilteredParams) ([]ListUserGroupsByAdminAscFilteredRow, error) {
	rows, err := q.db.QueryContext(ctx, listUserGroupsByAdminAscFiltered,
		arg.UserID,
		arg.Search,
		arg.Offset,
		arg.Limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []ListUserGroupsByAdminAscFilteredRow{}
	for rows.Next() {
		var i ListUserGroupsByAdminAscFilteredRow
		if err := rows.Scan(
			&i.ID,
			&i.Name,
			&i.AdminUserID,
			&i.CreatedAt,
			&i.Role,
			&i.AdminEmail,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const listUserGroupsByAdminDescFiltered = `-- name: ListUserGroupsByAdminDescFiltered :many
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
WHERE ga.user_id = ?1
  AND (
    ?2 = ''
    OR g.name LIKE '%' || ?2 || '%'
    OR u.email LIKE '%' || ?2 || '%'
  )
ORDER BY admin_email COLLATE NOCASE DESC, g.name COLLATE NOCASE ASC
LIMIT ?4 OFFSET ?3
`

type ListUserGroupsByAdminDescFilteredParams struct {
	UserID string      `json:"user_id"`
	Search interface{} `json:"search"`
	Offset int64       `json:"offset"`
	Limit  int64       `json:"limit"`
}

type ListUserGroupsByAdminDescFilteredRow struct {
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	AdminUserID string       `json:"admin_user_id"`
	CreatedAt   sql.NullTime `json:"created_at"`
	Role        string       `json:"role"`
	AdminEmail  string       `json:"admin_email"`
}

func (q *Queries) ListUserGroupsByAdminDescFiltered(ctx context.Context, arg ListUserGroupsByAdminDescFilteredParams) ([]ListUserGroupsByAdminDescFilteredRow, error) {
	rows, err := q.db.QueryContext(ctx, listUserGroupsByAdminDescFiltered,
		arg.UserID,
		arg.Search,
		arg.Offset,
		arg.Limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []ListUserGroupsByAdminDescFilteredRow{}
	for rows.Next() {
		var i ListUserGroupsByAdminDescFilteredRow
		if err := rows.Scan(
			&i.ID,
			&i.Name,
			&i.AdminUserID,
			&i.CreatedAt,
			&i.Role,
			&i.AdminEmail,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const listUserGroupsByCreatedAscFiltered = `-- name: ListUserGroupsByCreatedAscFiltered :many
SELECT
  g.id,
  g.name,
  g.admin_user_id,
  g.created_at,
  ga.role
FROM groups g
JOIN group_access ga ON ga.group_id = g.id
JOIN users u ON u.id = g.admin_user_id
WHERE ga.user_id = ?1
  AND (
    ?2 = ''
    OR g.name LIKE '%' || ?2 || '%'
    OR u.email LIKE '%' || ?2 || '%'
  )
ORDER BY g.created_at ASC, g.name COLLATE NOCASE ASC
LIMIT ?4 OFFSET ?3
`

type ListUserGroupsByCreatedAscFilteredParams struct {
	UserID string      `json:"user_id"`
	Search interface{} `json:"search"`
	Offset int64       `json:"offset"`
	Limit  int64       `json:"limit"`
}

type ListUserGroupsByCreatedAscFilteredRow struct {
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	AdminUserID string       `json:"admin_user_id"`
	CreatedAt   sql.NullTime `json:"created_at"`
	Role        string       `json:"role"`
}

func (q *Queries) ListUserGroupsByCreatedAscFiltered(ctx context.Context, arg ListUserGroupsByCreatedAscFilteredParams) ([]ListUserGroupsByCreatedAscFilteredRow, error) {
	rows, err := q.db.QueryContext(ctx, listUserGroupsByCreatedAscFiltered,
		arg.UserID,
		arg.Search,
		arg.Offset,
		arg.Limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []ListUserGroupsByCreatedAscFilteredRow{}
	for rows.Next() {
		var i ListUserGroupsByCreatedAscFilteredRow
		if err := rows.Scan(
			&i.ID,
			&i.Name,
			&i.AdminUserID,
			&i.CreatedAt,
			&i.Role,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const listUserGroupsByCreatedDescFiltered = `-- name: ListUserGroupsByCreatedDescFiltered :many
SELECT
  g.id,
  g.name,
  g.admin_user_id,
  g.created_at,
  ga.role
FROM groups g
JOIN group_access ga ON ga.group_id = g.id
JOIN users u ON u.id = g.admin_user_id
WHERE ga.user_id = ?1
  AND (
    ?2 = ''
    OR g.name LIKE '%' || ?2 || '%'
    OR u.email LIKE '%' || ?2 || '%'
  )
ORDER BY g.created_at DESC, g.name COLLATE NOCASE ASC
LIMIT ?4 OFFSET ?3
`

type ListUserGroupsByCreatedDescFilteredParams struct {
	UserID string      `json:"user_id"`
	Search interface{} `json:"search"`
	Offset int64       `json:"offset"`
	Limit  int64       `json:"limit"`
}

type ListUserGroupsByCreatedDescFilteredRow struct {
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	AdminUserID string       `json:"admin_user_id"`
	CreatedAt   sql.NullTime `json:"created_at"`
	Role        string       `json:"role"`
}

func (q *Queries) ListUserGroupsByCreatedDescFiltered(ctx context.Context, arg ListUserGroupsByCreatedDescFilteredParams) ([]ListUserGroupsByCreatedDescFilteredRow, error) {
	rows, err := q.db.QueryContext(ctx, listUserGroupsByCreatedDescFiltered,
		arg.UserID,
		arg.Search,
		arg.Offset,
		arg.Limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []ListUserGroupsByCreatedDescFilteredRow{}
	for rows.Next() {
		var i ListUserGroupsByCreatedDescFilteredRow
		if err := rows.Scan(
			&i.ID,
			&i.Name,
			&i.AdminUserID,
			&i.CreatedAt,
			&i.Role,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const listUserGroupsByNameAscFiltered = `-- name: ListUserGroupsByNameAscFiltered :many
SELECT
  g.id,
  g.name,
  g.admin_user_id,
  g.created_at,
  ga.role
FROM groups g
JOIN group_access ga ON ga.group_id = g.id
JOIN users u ON u.id = g.admin_user_id
WHERE ga.user_id = ?1
  AND (
    ?2 = ''
    OR g.name LIKE '%' || ?2 || '%'
    OR u.email LIKE '%' || ?2 || '%'
  )
ORDER BY g.name COLLATE NOCASE ASC, g.created_at DESC
LIMIT ?4 OFFSET ?3
`

type ListUserGroupsByNameAscFilteredParams struct {
	UserID string      `json:"user_id"`
	Search interface{} `json:"search"`
	Offset int64       `json:"offset"`
	Limit  int64       `json:"limit"`
}

type ListUserGroupsByNameAscFilteredRow struct {
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	AdminUserID string       `json:"admin_user_id"`
	CreatedAt   sql.NullTime `json:"created_at"`
	Role        string       `json:"role"`
}

func (q *Queries) ListUserGroupsByNameAscFiltered(ctx context.Context, arg ListUserGroupsByNameAscFilteredParams) ([]ListUserGroupsByNameAscFilteredRow, error) {
	rows, err := q.db.QueryContext(ctx, listUserGroupsByNameAscFiltered,
		arg.UserID,
		arg.Search,
		arg.Offset,
		arg.Limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []ListUserGroupsByNameAscFilteredRow{}
	for rows.Next() {
		var i ListUserGroupsByNameAscFilteredRow
		if err := rows.Scan(
			&i.ID,
			&i.Name,
			&i.AdminUserID,
			&i.CreatedAt,
			&i.Role,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const listUserGroupsByNameDescFiltered = `-- name: ListUserGroupsByNameDescFiltered :many
SELECT
  g.id,
  g.name,
  g.admin_user_id,
  g.created_at,
  ga.role
FROM groups g
JOIN group_access ga ON ga.group_id = g.id
JOIN users u ON u.id = g.admin_user_id
WHERE ga.user_id = ?1
  AND (
    ?2 = ''
    OR g.name LIKE '%' || ?2 || '%'
    OR u.email LIKE '%' || ?2 || '%'
  )
ORDER BY g.name COLLATE NOCASE DESC, g.created_at DESC
LIMIT ?4 OFFSET ?3
`

type ListUserGroupsByNameDescFilteredParams struct {
	UserID string      `json:"user_id"`
	Search interface{} `json:"search"`
	Offset int64       `json:"offset"`
	Limit  int64       `json:"limit"`
}

type ListUserGroupsByNameDescFilteredRow struct {
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	AdminUserID string       `json:"admin_user_id"`
	CreatedAt   sql.NullTime `json:"created_at"`
	Role        string       `json:"role"`
}

func (q *Queries) ListUserGroupsByNameDescFiltered(ctx context.Context, arg ListUserGroupsByNameDescFilteredParams) ([]ListUserGroupsByNameDescFilteredRow, error) {
	rows, err := q.db.QueryContext(ctx, listUserGroupsByNameDescFiltered,
		arg.UserID,
		arg.Search,
		arg.Offset,
		arg.Limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []ListUserGroupsByNameDescFilteredRow{}
	for rows.Next() {
		var i ListUserGroupsByNameDescFilteredRow
		if err := rows.Scan(
			&i.ID,
			&i.Name,
			&i.AdminUserID,
			&i.CreatedAt,
			&i.Role,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const listUserSessions = `-- name: ListUserSessions :many
SELECT id, user_id, token, created_at, expires_at FROM user_sessions
WHERE user_id = ? AND expires_at > CURRENT_TIMESTAMP
ORDER BY created_at DESC
`

func (q *Queries) ListUserSessions(ctx context.Context, userID string) ([]UserSession, error) {
	rows, err := q.db.QueryContext(ctx, listUserSessions, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []UserSession{}
	for rows.Next() {
		var i UserSession
		if err := rows.Scan(
			&i.ID,
			&i.UserID,
			&i.Token,
			&i.CreatedAt,
			&i.ExpiresAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const removeGroupAdmin = `-- name: RemoveGroupAdmin :exec
DELETE FROM group_access
WHERE user_id = ?
  AND group_id = ?
  AND role = 'admin'
`

type RemoveGroupAdminParams struct {
	UserID  string `json:"user_id"`
	GroupID string `json:"group_id"`
}

func (q *Queries) RemoveGroupAdmin(ctx context.Context, arg RemoveGroupAdminParams) error {
	_, err := q.db.ExecContext(ctx, removeGroupAdmin, arg.UserID, arg.GroupID)
	return err
}

const removeGroupReader = `-- name: RemoveGroupReader :exec
DELETE FROM group_access
WHERE user_id = ?
  AND group_id = ?
  AND role = 'viewer'
`

type RemoveGroupReaderParams struct {
	UserID  string `json:"user_id"`
	GroupID string `json:"group_id"`
}

func (q *Queries) RemoveGroupReader(ctx context.Context, arg RemoveGroupReaderParams) error {
	_, err := q.db.ExecContext(ctx, removeGroupReader, arg.UserID, arg.GroupID)
	return err
}

const unbanUser = `-- name: UnbanUser :exec
DELETE FROM banned_users
WHERE user_id = ?
`

func (q *Queries) UnbanUser(ctx context.Context, userID string) error {
	_, err := q.db.ExecContext(ctx, unbanUser, userID)
	return err
}

const updateGroupAdmin = `-- name: UpdateGroupAdmin :exec
UPDATE groups
SET admin_user_id = ?
WHERE id = ?
`

type UpdateGroupAdminParams struct {
	AdminUserID string `json:"admin_user_id"`
	ID          string `json:"id"`
}

func (q *Queries) UpdateGroupAdmin(ctx context.Context, arg UpdateGroupAdminParams) error {
	_, err := q.db.ExecContext(ctx, updateGroupAdmin, arg.AdminUserID, arg.ID)
	return err
}

const updateGroupName = `-- name: UpdateGroupName :one
UPDATE groups
SET name = ?
WHERE id = ?
RETURNING id, name, admin_user_id, created_at
`

type UpdateGroupNameParams struct {
	Name string `json:"name"`
	ID   string `json:"id"`
}

func (q *Queries) UpdateGroupName(ctx context.Context, arg UpdateGroupNameParams) (Group, error) {
	row := q.db.QueryRowContext(ctx, updateGroupName, arg.Name, arg.ID)
	var i Group
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.AdminUserID,
		&i.CreatedAt,
	)
	return i, err
}

const updateUserPreferredLang = `-- name: UpdateUserPreferredLang :exec
UPDATE users
SET preferred_lang = ?
WHERE id = ?
`

type UpdateUserPreferredLangParams struct {
	PreferredLang string `json:"preferred_lang"`
	ID            string `json:"id"`
}

func (q *Queries) UpdateUserPreferredLang(ctx context.Context, arg UpdateUserPreferredLangParams) error {
	_, err := q.db.ExecContext(ctx, updateUserPreferredLang, arg.PreferredLang, arg.ID)
	return err
}

const useMagicLink = `-- name: UseMagicLink :exec
UPDATE magic_links
SET used_at = CURRENT_TIMESTAMP
WHERE id = ?
`

func (q *Queries) UseMagicLink(ctx context.Context, id string) error {
	_, err := q.db.ExecContext(ctx, useMagicLink, id)
	return err
}
