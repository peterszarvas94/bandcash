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

type memberResource struct{}

func (memberResource) BaseCountQuery() *bun.SelectQuery {
	return BunDB.NewSelect().TableExpr("members")
}

func (memberResource) BaseListQuery(rows *[]Member) *bun.SelectQuery {
	return BunDB.NewSelect().Model(rows)
}

func (memberResource) ApplyFilter(q *bun.SelectQuery, filter memberFilter) *bun.SelectQuery {
	return applyMemberTableFilters(q, filter.GroupID, filter.Search)
}

func (memberResource) OrderSpec() BunTableOrderSpec {
	return memberTableOrderSpec
}

var membersRes memberResource

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
	return Count[Member, memberFilter](ctx, membersRes, memberFilter{GroupID: groupID, Search: search})
}

func ListMembersTable(ctx context.Context, params MemberTableListParams) ([]Member, error) {
	return List[Member, memberFilter](
		ctx,
		membersRes,
		memberFilter{GroupID: params.GroupID, Search: params.Search},
		params.Sort,
		params.Dir,
		params.Limit,
		params.Offset,
	)
}
