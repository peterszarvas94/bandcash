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

type userGroupResource struct{}

func (userGroupResource) BaseCountQuery() *bun.SelectQuery {
	return BunDB.NewSelect().TableExpr("group_access ga").Join("JOIN groups g ON g.id = ga.group_id")
}

func (userGroupResource) BaseListQuery(rows *[]UserGroupRow) *bun.SelectQuery {
	return BunDB.NewSelect().
		ColumnExpr("g.id AS id").
		ColumnExpr("g.name AS name").
		ColumnExpr("g.admin_user_id AS admin_user_id").
		ColumnExpr("g.created_at AS created_at").
		ColumnExpr("ga.role AS role").
		TableExpr("group_access ga").
		Join("JOIN groups g ON g.id = ga.group_id")
}

func (userGroupResource) ApplyFilter(q *bun.SelectQuery, filter userGroupFilter) *bun.SelectQuery {
	q = q.Where("ga.user_id = ?", filter.UserID)
	return applyOptionalSearch(q, filter.Search, func(sq *bun.SelectQuery, value string) *bun.SelectQuery {
		return sq.Where("g.name LIKE ?", "%"+value+"%")
	})
}

func (userGroupResource) OrderSpec() BunTableOrderSpec {
	return userGroupOrderSpec
}

var groupsRes userGroupResource

var userGroupOrderSpec = BunTableOrderSpec{
	DefaultSort: "name",
	DefaultDir:  "ASC",
	AllowedSorts: map[string]func(*bun.SelectQuery, string) *bun.SelectQuery{
		"name": func(q *bun.SelectQuery, dir string) *bun.SelectQuery { return q.OrderExpr("g.name " + dir) },
	},
	StableSort: func(q *bun.SelectQuery) *bun.SelectQuery { return q.OrderExpr("g.created_at DESC") },
}

func CountUserGroupsTable(ctx context.Context, userID, search string) (int64, error) {
	return Count[UserGroupRow, userGroupFilter](ctx, groupsRes, userGroupFilter{UserID: userID, Search: search})
}

func ListUserGroupsTable(ctx context.Context, userID, search string, limit, offset int) ([]UserGroupRow, error) {
	return List[UserGroupRow, userGroupFilter](
		ctx,
		groupsRes,
		userGroupFilter{UserID: userID, Search: search},
		"name",
		"asc",
		limit,
		offset,
	)
}
