package db

import (
	"context"

	"github.com/uptrace/bun"
)

type MemberTableListParams struct {
	GroupID string
	Search  string
	Sort    string
	Dir     string
	Limit   int
	Offset  int
}

type memberFilter struct {
	GroupID string
	Search  string
}

func applyMemberTableFilters(q *bun.SelectQuery, groupID, search string) *bun.SelectQuery {
	q = q.Where("group_id = ?", groupID)
	return applyOptionalSearch(q, search, func(sq *bun.SelectQuery, search string) *bun.SelectQuery {
		like := "%" + search + "%"
		return sq.WhereGroup(" AND ", func(qq *bun.SelectQuery) *bun.SelectQuery {
			return qq.Where("name LIKE ?", like).WhereOr("description LIKE ?", like)
		})
	})
}

var memberTableOrderSpec = BunTableOrderSpec{
	DefaultSort: "createdAt",
	DefaultDir:  "DESC",
	AllowedSorts: map[string]func(*bun.SelectQuery, string) *bun.SelectQuery{
		"name":        func(q *bun.SelectQuery, dir string) *bun.SelectQuery { return q.OrderExpr("name " + dir) },
		"description": func(q *bun.SelectQuery, dir string) *bun.SelectQuery { return q.OrderExpr("description " + dir) },
		"createdAt":   func(q *bun.SelectQuery, dir string) *bun.SelectQuery { return q.OrderExpr("created_at " + dir) },
	},
	StableSort: func(q *bun.SelectQuery) *bun.SelectQuery { return q.OrderExpr("created_at DESC") },
}

func CountMembersTable(ctx context.Context, groupID, search string) (int64, error) {
	q := BunDB.NewSelect().TableExpr("members")
	q = applyMemberTableFilters(q, groupID, search)
	count, err := q.Count(ctx)
	return int64(count), err
}

func ListMembersTable(ctx context.Context, params MemberTableListParams) ([]Member, error) {
	rows := make([]Member, 0)
	q := BunDB.NewSelect().Model(&rows)
	q = applyMemberTableFilters(q, params.GroupID, params.Search)
	q = applyTableOrdering(q, memberTableOrderSpec, params.Sort, params.Dir)
	q = applyTablePagination(q, params.Limit, params.Offset)
	err := q.Scan(ctx)
	return rows, err
}
