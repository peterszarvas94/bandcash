package db

import (
	"context"
	"strings"

	"github.com/uptrace/bun"
)

type Resource[T any, F any] interface {
	BaseCountQuery() *bun.SelectQuery
	BaseListQuery(rows *[]T) *bun.SelectQuery
	ApplyFilter(q *bun.SelectQuery, filter F) *bun.SelectQuery
	OrderSpec() BunTableOrderSpec
}

func Count[T any, F any](ctx context.Context, r Resource[T, F], filter F) (int64, error) {
	q := r.BaseCountQuery()
	q = r.ApplyFilter(q, filter)
	count, err := q.Count(ctx)
	return int64(count), err
}

func List[T any, F any](
	ctx context.Context,
	r Resource[T, F],
	filter F,
	sort, dir string,
	limit, offset int,
) ([]T, error) {
	rows := make([]T, 0)
	q := r.BaseListQuery(&rows)
	q = r.ApplyFilter(q, filter)
	q = applyTableOrdering(q, r.OrderSpec(), sort, dir)
	q = applyTablePagination(q, limit, offset)
	err := q.Scan(ctx)
	return rows, err
}

type BunTableOrderSpec struct {
	DefaultSort  string
	DefaultDir   string
	AllowedSorts map[string]func(*bun.SelectQuery, string) *bun.SelectQuery
	StableSort   func(*bun.SelectQuery) *bun.SelectQuery
}

func normalizeSortKey(sort string, spec BunTableOrderSpec) string {
	sort = strings.TrimSpace(sort)
	if sort == "" {
		sort = spec.DefaultSort
	}
	if _, ok := spec.AllowedSorts[sort]; !ok {
		return spec.DefaultSort
	}
	return sort
}

func normalizeSortDir(dir, defaultDir string) string {
	defaultDir = strings.ToUpper(strings.TrimSpace(defaultDir))
	if defaultDir != "ASC" && defaultDir != "DESC" {
		defaultDir = "DESC"
	}

	dir = strings.ToUpper(strings.TrimSpace(dir))
	if dir != "ASC" && dir != "DESC" {
		return defaultDir
	}
	return dir
}

func applyTableOrdering(q *bun.SelectQuery, spec BunTableOrderSpec, sort, dir string) *bun.SelectQuery {
	sort = normalizeSortKey(sort, spec)
	dir = normalizeSortDir(dir, spec.DefaultDir)

	if applySort, ok := spec.AllowedSorts[sort]; ok {
		q = applySort(q, dir)
	}
	if spec.StableSort != nil {
		q = spec.StableSort(q)
	}
	return q
}

func applyTablePagination(q *bun.SelectQuery, limit, offset int) *bun.SelectQuery {
	if limit > 0 {
		q = q.Limit(limit)
	}
	if offset > 0 {
		q = q.Offset(offset)
	}
	return q
}

func applyOptionalSearch(q *bun.SelectQuery, search string, apply func(*bun.SelectQuery, string) *bun.SelectQuery) *bun.SelectQuery {
	search = strings.TrimSpace(search)
	if search == "" || apply == nil {
		return q
	}
	return apply(q, search)
}

func applyDateRangeOrYear(
	q *bun.SelectQuery,
	from, to, year string,
	applyRange func(*bun.SelectQuery, string, string) *bun.SelectQuery,
	applyYear func(*bun.SelectQuery, string) *bun.SelectQuery,
) *bun.SelectQuery {
	from = strings.TrimSpace(from)
	to = strings.TrimSpace(to)
	year = strings.TrimSpace(year)

	if from != "" && to != "" && applyRange != nil {
		return applyRange(q, from, to)
	}
	if year != "" && applyYear != nil {
		return applyYear(q, year)
	}
	return q
}
