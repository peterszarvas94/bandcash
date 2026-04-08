package db

import (
	"context"
	"database/sql"

	"github.com/uptrace/bun"
)

type UserGroupRow struct {
	ID          string
	Name        string
	AdminUserID string
	CreatedAt   sql.NullTime
	Role        string
}

func CountUserGroupsTable(ctx context.Context, userID, search string) (int64, error) {
	q := BunDB.NewSelect().TableExpr("group_access ga").
		Join("JOIN groups g ON g.id = ga.group_id").
		Where("ga.user_id = ?", userID)
	q = applyOptionalSearch(q, search, func(sq *bun.SelectQuery, value string) *bun.SelectQuery {
		return sq.Where("g.name LIKE ?", "%"+value+"%")
	})

	count, err := q.Count(ctx)
	return int64(count), err
}

func ListUserGroupsTable(ctx context.Context, userID, search string, limit, offset int) ([]UserGroupRow, error) {
	rows := make([]UserGroupRow, 0)

	q := BunDB.NewSelect().
		ColumnExpr("g.id AS id").
		ColumnExpr("g.name AS name").
		ColumnExpr("g.admin_user_id AS admin_user_id").
		ColumnExpr("g.created_at AS created_at").
		ColumnExpr("ga.role AS role").
		TableExpr("group_access ga").
		Join("JOIN groups g ON g.id = ga.group_id").
		Where("ga.user_id = ?", userID)
	q = applyOptionalSearch(q, search, func(sq *bun.SelectQuery, value string) *bun.SelectQuery {
		return sq.Where("g.name LIKE ?", "%"+value+"%")
	})
	q = applyTableOrdering(q, BunTableOrderSpec{
		DefaultSort: "name",
		DefaultDir:  "ASC",
		AllowedSorts: map[string]func(*bun.SelectQuery, string) *bun.SelectQuery{
			"name": func(qq *bun.SelectQuery, dir string) *bun.SelectQuery { return qq.OrderExpr("g.name " + dir) },
		},
		StableSort: func(qq *bun.SelectQuery) *bun.SelectQuery { return qq.OrderExpr("g.created_at DESC") },
	}, "name", "asc")
	q = applyTablePagination(q, limit, offset)

	err := q.Scan(ctx, &rows)
	return rows, err
}
