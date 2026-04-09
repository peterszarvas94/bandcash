package data

import "database/sql"

type CreateGroupParams struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	AdminUserID string `json:"admin_user_id"`
}

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

type RemoveGroupAdminParams struct {
	UserID  string `json:"user_id"`
	GroupID string `json:"group_id"`
}

type RemoveGroupReaderParams struct {
	UserID  string `json:"user_id"`
	GroupID string `json:"group_id"`
}

type UpdateGroupAdminParams struct {
	AdminUserID string `json:"admin_user_id"`
	ID          string `json:"id"`
}

type UpdateGroupNameParams struct {
	Name string `json:"name"`
	ID   string `json:"id"`
}

type CreateInviteMagicLinkParams struct {
	ID         string         `json:"id"`
	Token      string         `json:"token"`
	Email      string         `json:"email"`
	GroupID    sql.NullString `json:"group_id"`
	InviteRole string         `json:"invite_role"`
}

type DeleteGroupPendingInviteParams struct {
	ID      string         `json:"id"`
	GroupID sql.NullString `json:"group_id"`
}

type GetGroupAccessRoleParams struct {
	UserID  string `json:"user_id"`
	GroupID string `json:"group_id"`
}

type IsGroupReaderParams struct {
	UserID  string `json:"user_id"`
	GroupID string `json:"group_id"`
}

type ListGroupUserAccessRow struct {
	ID              string       `json:"id"`
	Email           string       `json:"email"`
	CreatedAt       sql.NullTime `json:"created_at"`
	PreferredLang   string       `json:"preferred_lang"`
	Role            string       `json:"role"`
	AccessCreatedAt sql.NullTime `json:"access_created_at"`
}
