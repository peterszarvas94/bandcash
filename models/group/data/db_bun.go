package data

import (
	"context"
	"database/sql"

	"bandcash/internal/db"
)

func GetGroupByID(ctx context.Context, id string) (db.Group, error) {
	var row db.Group
	err := db.BunDB.NewSelect().Model(&row).Where("id = ?", id).Scan(ctx)
	return row, err
}

func CreateGroupAdmin(ctx context.Context, arg CreateGroupAdminParams) (CreateGroupAdminRow, error) {
	var row CreateGroupAdminRow
	err := db.BunDB.QueryRowContext(
		ctx,
		"INSERT INTO group_access (id, user_id, group_id, role) VALUES (?, ?, ?, 'admin') ON CONFLICT(user_id, group_id) DO UPDATE SET role = 'admin' WHERE group_access.role != 'owner' RETURNING id, user_id, group_id, created_at",
		arg.ID,
		arg.UserID,
		arg.GroupID,
	).Scan(&row.ID, &row.UserID, &row.GroupID, &row.CreatedAt)
	return row, err
}

func CreateGroupReader(ctx context.Context, arg CreateGroupReaderParams) (CreateGroupReaderRow, error) {
	var row CreateGroupReaderRow
	err := db.BunDB.QueryRowContext(
		ctx,
		"INSERT INTO group_access (id, user_id, group_id, role) VALUES (?, ?, ?, 'viewer') ON CONFLICT(user_id, group_id) DO UPDATE SET role = 'viewer' WHERE group_access.role != 'owner' RETURNING id, user_id, group_id, created_at",
		arg.ID,
		arg.UserID,
		arg.GroupID,
	).Scan(&row.ID, &row.UserID, &row.GroupID, &row.CreatedAt)
	return row, err
}

func RemoveGroupAdmin(ctx context.Context, arg RemoveGroupAdminParams) error {
	_, err := db.BunDB.NewDelete().
		TableExpr("group_access").
		Where("user_id = ?", arg.UserID).
		Where("group_id = ?", arg.GroupID).
		Where("role = 'admin'").
		Exec(ctx)
	return err
}

func RemoveGroupReader(ctx context.Context, arg RemoveGroupReaderParams) error {
	_, err := db.BunDB.NewDelete().
		TableExpr("group_access").
		Where("user_id = ?", arg.UserID).
		Where("group_id = ?", arg.GroupID).
		Where("role = 'viewer'").
		Exec(ctx)
	return err
}

func UpdateGroupAdmin(ctx context.Context, arg UpdateGroupAdminParams) error {
	_, err := db.BunDB.NewUpdate().
		TableExpr("groups").
		Set("admin_user_id = ?", arg.AdminUserID).
		Where("id = ?", arg.ID).
		Exec(ctx)
	return err
}

func GetGroupAccessRole(ctx context.Context, arg GetGroupAccessRoleParams) (string, error) {
	var role string
	err := db.BunDB.NewSelect().
		TableExpr("group_access").
		Column("role").
		Where("user_id = ?", arg.UserID).
		Where("group_id = ?", arg.GroupID).
		Scan(ctx, &role)
	return role, err
}

func IsGroupReader(ctx context.Context, arg IsGroupReaderParams) (int64, error) {
	n, err := db.BunDB.NewSelect().
		TableExpr("group_access").
		Where("user_id = ?", arg.UserID).
		Where("group_id = ?", arg.GroupID).
		Where("role = 'viewer'").
		Count(ctx)
	return int64(n), err
}

func ListGroupAdmins(ctx context.Context, groupID string) ([]db.User, error) {
	rows := make([]db.User, 0)
	err := db.BunDB.NewSelect().
		TableExpr("users AS u").
		ColumnExpr("u.id, u.email, u.created_at, u.preferred_lang").
		Join("JOIN group_access ga ON ga.user_id = u.id").
		Where("ga.group_id = ?", groupID).
		Where("ga.role IN ('owner', 'admin')").
		OrderExpr("LOWER(u.email) ASC").
		Scan(ctx, &rows)
	return rows, err
}

func ListGroupUserAccess(ctx context.Context, groupID string) ([]ListGroupUserAccessRow, error) {
	rows := make([]ListGroupUserAccessRow, 0)
	err := db.BunDB.NewSelect().
		TableExpr("users").
		ColumnExpr("users.id").
		ColumnExpr("users.email").
		ColumnExpr("users.created_at").
		ColumnExpr("users.preferred_lang").
		ColumnExpr("group_access.role").
		ColumnExpr("group_access.created_at AS access_created_at").
		Join("JOIN group_access ON group_access.user_id = users.id").
		Where("group_access.group_id = ?", groupID).
		OrderExpr("group_access.created_at ASC").
		OrderExpr("LOWER(users.email) ASC").
		Scan(ctx, &rows)
	return rows, err
}

func CreateGroup(ctx context.Context, arg CreateGroupParams) (db.Group, error) {
	group := db.Group{ID: arg.ID, Name: arg.Name, AdminUserID: arg.AdminUserID}
	if _, err := db.BunDB.NewInsert().Model(&group).Exec(ctx); err != nil {
		return db.Group{}, err
	}
	return GetGroupByID(ctx, arg.ID)
}

func DeleteGroup(ctx context.Context, id string) error {
	_, err := db.BunDB.NewDelete().TableExpr("groups").Where("id = ?", id).Exec(ctx)
	return err
}

func UpdateGroupName(ctx context.Context, arg UpdateGroupNameParams) (db.Group, error) {
	_, err := db.BunDB.NewUpdate().
		TableExpr("groups").
		Set("name = ?", arg.Name).
		Where("id = ?", arg.ID).
		Exec(ctx)
	if err != nil {
		return db.Group{}, err
	}
	return GetGroupByID(ctx, arg.ID)
}

func ListGroupsByAdmin(ctx context.Context, userID string) ([]db.Group, error) {
	rows := make([]db.Group, 0)
	err := db.BunDB.NewSelect().
		TableExpr("groups AS g").
		ColumnExpr("g.id, g.name, g.admin_user_id, g.created_at").
		Join("JOIN group_access ga ON ga.group_id = g.id").
		Where("ga.user_id = ?", userID).
		Where("ga.role IN ('owner', 'admin')").
		OrderExpr("g.created_at DESC").
		Scan(ctx, &rows)
	return rows, err
}

func ListGroupsByReader(ctx context.Context, userID string) ([]db.Group, error) {
	rows := make([]db.Group, 0)
	err := db.BunDB.NewSelect().
		TableExpr("groups AS g").
		ColumnExpr("g.id, g.name, g.admin_user_id, g.created_at").
		Join("JOIN group_access ga ON ga.group_id = g.id").
		Where("ga.user_id = ?", userID).
		Where("ga.role = 'viewer'").
		OrderExpr("g.created_at DESC").
		Scan(ctx, &rows)
	return rows, err
}

func ListGroupPendingInvites(ctx context.Context, groupID sql.NullString) ([]db.MagicLink, error) {
	rows := make([]db.MagicLink, 0)
	q := db.BunDB.NewSelect().
		Model(&rows).
		Where("action = 'invite'").
		Where("used_at IS NULL")
	if groupID.Valid {
		q = q.Where("group_id = ?", groupID.String)
	}
	err := q.OrderExpr("created_at DESC").Scan(ctx)
	return rows, err
}

func DeleteGroupPendingInvite(ctx context.Context, arg DeleteGroupPendingInviteParams) error {
	if !arg.GroupID.Valid {
		return sql.ErrNoRows
	}
	_, err := db.BunDB.NewDelete().
		TableExpr("magic_links").
		Where("id = ?", arg.ID).
		Where("action = 'invite'").
		Where("group_id = ?", arg.GroupID.String).
		Where("used_at IS NULL").
		Exec(ctx)
	return err
}

func CreateInviteMagicLink(ctx context.Context, arg CreateInviteMagicLinkParams) (db.MagicLink, error) {
	var row db.MagicLink
	err := db.BunDB.QueryRowContext(
		ctx,
		"INSERT INTO magic_links (id, token, email, action, group_id, expires_at, invite_role) VALUES (?, ?, ?, 'invite', ?, CURRENT_TIMESTAMP, ?) RETURNING id, token, email, action, group_id, expires_at, used_at, created_at, invite_role",
		arg.ID,
		arg.Token,
		arg.Email,
		arg.GroupID,
		arg.InviteRole,
	).Scan(&row.ID, &row.Token, &row.Email, &row.Action, &row.GroupID, &row.ExpiresAt, &row.UsedAt, &row.CreatedAt, &row.InviteRole)
	return row, err
}

func ListPaidOutgoingPaymentsByGroup(ctx context.Context, groupID string) ([]db.GroupOutgoingPayment, error) {
	rows := make([]db.GroupOutgoingPayment, 0)
	err := db.BunDB.NewSelect().
		TableExpr("group_outgoing_payments").
		Where("group_id = ?", groupID).
		Where("paid = 1").
		OrderExpr("COALESCE(paid_at, updated_at) DESC").
		OrderExpr("updated_at DESC").
		OrderExpr("payment_kind ASC").
		Scan(ctx, &rows)
	return rows, err
}

func ListUnpaidOutgoingPaymentsByGroup(ctx context.Context, groupID string) ([]db.GroupOutgoingPayment, error) {
	rows := make([]db.GroupOutgoingPayment, 0)
	err := db.BunDB.NewSelect().
		TableExpr("group_outgoing_payments").
		Where("group_id = ?", groupID).
		Where("paid = 0").
		OrderExpr("sort_date ASC").
		OrderExpr("updated_at DESC").
		OrderExpr("payment_kind ASC").
		Scan(ctx, &rows)
	return rows, err
}
