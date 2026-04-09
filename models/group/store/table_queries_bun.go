package store

import (
	"context"
	"database/sql"
	"strings"

	"bandcash/internal/db"
)

type UserGroupRow struct {
	ID          string
	Name        string
	AdminUserID string
	CreatedAt   sql.NullTime
	Role        string
}

func CountUserGroupsTable(ctx context.Context, userID, search string) (int64, error) {
	q := db.BunDB.NewSelect().TableExpr("group_access ga").Join("JOIN groups g ON g.id = ga.group_id")
	q = q.Where("ga.user_id = ?", userID)
	search = strings.TrimSpace(search)
	if search != "" {
		q = q.Where("g.name LIKE ?", "%"+search+"%")
	}
	n, err := q.Count(ctx)
	return int64(n), err
}

func ListUserGroupsTable(ctx context.Context, userID, search string, limit, offset int) ([]UserGroupRow, error) {
	rows := make([]UserGroupRow, 0)
	q := db.BunDB.NewSelect().
		ColumnExpr("g.id AS id").
		ColumnExpr("g.name AS name").
		ColumnExpr("g.admin_user_id AS admin_user_id").
		ColumnExpr("g.created_at AS created_at").
		ColumnExpr("ga.role AS role").
		TableExpr("group_access ga").
		Join("JOIN groups g ON g.id = ga.group_id").
		Where("ga.user_id = ?", userID)

	search = strings.TrimSpace(search)
	if search != "" {
		q = q.Where("g.name LIKE ?", "%"+search+"%")
	}

	q = q.OrderExpr("g.name ASC").OrderExpr("g.created_at DESC")
	if limit > 0 {
		q = q.Limit(limit)
	}
	if offset > 0 {
		q = q.Offset(offset)
	}

	err := q.Scan(ctx, &rows)
	return rows, err
}
