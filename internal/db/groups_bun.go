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

type userGroupFilter struct {
	UserID string
	Search string
}

func applyUserGroupFilters(q *bun.SelectQuery, filter userGroupFilter) *bun.SelectQuery {
	q = q.Where("ga.user_id = ?", filter.UserID)
	return applyOptionalSearch(q, filter.Search, func(sq *bun.SelectQuery, value string) *bun.SelectQuery {
		return sq.Where("g.name LIKE ?", "%"+value+"%")
	})
}

var userGroupOrderSpec = BunTableOrderSpec{
	DefaultSort: "name",
	DefaultDir:  "ASC",
	AllowedSorts: map[string]func(*bun.SelectQuery, string) *bun.SelectQuery{
		"name": func(q *bun.SelectQuery, dir string) *bun.SelectQuery { return q.OrderExpr("g.name " + dir) },
	},
	StableSort: func(q *bun.SelectQuery) *bun.SelectQuery { return q.OrderExpr("g.created_at DESC") },
}

func CountUserGroupsTable(ctx context.Context, userID, search string) (int64, error) {
	q := BunDB.NewSelect().TableExpr("group_access ga").Join("JOIN groups g ON g.id = ga.group_id")
	q = applyUserGroupFilters(q, userGroupFilter{UserID: userID, Search: search})
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
		Join("JOIN groups g ON g.id = ga.group_id")
	q = applyUserGroupFilters(q, userGroupFilter{UserID: userID, Search: search})
	q = applyTableOrdering(q, userGroupOrderSpec, "name", "asc")
	q = applyTablePagination(q, limit, offset)
	err := q.Scan(ctx, &rows)
	return rows, err
}
