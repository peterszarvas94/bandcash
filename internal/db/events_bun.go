package db

import (
	"context"

	"github.com/uptrace/bun"
)

type EventTableFilter struct {
	GroupID string
	Search  string
	Year    string
	From    string
	To      string
}

type EventTableListParams struct {
	EventTableFilter
	Sort   string
	Dir    string
	Limit  int
	Offset int
}

type EventIncomeTotals struct {
	Total int64
	Paid  int64
}

type eventResource struct{}

func (eventResource) BaseCountQuery() *bun.SelectQuery {
	return BunDB.NewSelect().TableExpr("events")
}

func (eventResource) BaseListQuery(rows *[]Event) *bun.SelectQuery {
	return BunDB.NewSelect().Model(rows)
}

func (eventResource) ApplyFilter(q *bun.SelectQuery, filter EventTableFilter) *bun.SelectQuery {
	return applyEventTableFilters(q, filter)
}

func (eventResource) OrderSpec() BunTableOrderSpec {
	return eventTableOrderSpec
}

var eventsRes eventResource

func applyEventTableFilters(q *bun.SelectQuery, filter EventTableFilter) *bun.SelectQuery {
	q = q.Where("group_id = ?", filter.GroupID)

	q = applyOptionalSearch(q, filter.Search, func(sq *bun.SelectQuery, search string) *bun.SelectQuery {
		like := "%" + search + "%"
		return sq.WhereGroup(" AND ", func(qq *bun.SelectQuery) *bun.SelectQuery {
			return qq.Where("title LIKE ?", like).WhereOr("place LIKE ?", like)
		})
	})

	dateExpr := "COALESCE(NULLIF(date, ''), substr(time, 1, 10))"

	return applyDateRangeOrYear(
		q,
		filter.From,
		filter.To,
		filter.Year,
		func(sq *bun.SelectQuery, from, to string) *bun.SelectQuery {
			return sq.Where(dateExpr+" >= ?", from).Where(dateExpr+" <= ?", to)
		},
		func(sq *bun.SelectQuery, year string) *bun.SelectQuery {
			return sq.Where("substr("+dateExpr+", 1, 4) = ?", year)
		},
	)
}

var eventTableOrderSpec = BunTableOrderSpec{
	DefaultSort: "date",
	DefaultDir:  "DESC",
	AllowedSorts: map[string]func(*bun.SelectQuery, string) *bun.SelectQuery{
		"date": func(q *bun.SelectQuery, dir string) *bun.SelectQuery {
			return q.OrderExpr("COALESCE(NULLIF(date, ''), substr(time, 1, 10)) " + dir)
		},
		"time": func(q *bun.SelectQuery, dir string) *bun.SelectQuery {
			return q.OrderExpr("COALESCE(NULLIF(event_time, ''), substr(time, 12, 5)) " + dir)
		},
		"title":       func(q *bun.SelectQuery, dir string) *bun.SelectQuery { return q.OrderExpr("title " + dir) },
		"place":       func(q *bun.SelectQuery, dir string) *bun.SelectQuery { return q.OrderExpr("place " + dir) },
		"amount":      func(q *bun.SelectQuery, dir string) *bun.SelectQuery { return q.OrderExpr("amount " + dir) },
		"description": func(q *bun.SelectQuery, dir string) *bun.SelectQuery { return q.OrderExpr("description " + dir) },
		"paid":        func(q *bun.SelectQuery, dir string) *bun.SelectQuery { return q.OrderExpr("paid " + dir) },
		"paid_at": func(q *bun.SelectQuery, dir string) *bun.SelectQuery {
			return q.OrderExpr("(paid_at IS NULL OR paid_at = '') " + dir).OrderExpr("paid_at " + dir)
		},
	},
	StableSort: func(q *bun.SelectQuery) *bun.SelectQuery { return q.OrderExpr("created_at DESC") },
}

func CountEventsTable(ctx context.Context, filter EventTableFilter) (int64, error) {
	return Count[Event, EventTableFilter](ctx, eventsRes, filter)
}

func ListEventsTable(ctx context.Context, params EventTableListParams) ([]Event, error) {
	return List[Event, EventTableFilter](ctx, eventsRes, params.EventTableFilter, params.Sort, params.Dir, params.Limit, params.Offset)
}

func SumEventIncomeTotalsTable(ctx context.Context, filter EventTableFilter) (EventIncomeTotals, error) {
	var totals EventIncomeTotals

	q := BunDB.NewSelect().TableExpr("events")
	q = applyEventTableFilters(q, filter)

	err := q.ColumnExpr("CAST(COALESCE(SUM(amount), 0) AS INTEGER) AS total").
		ColumnExpr("CAST(COALESCE(SUM(CASE WHEN paid = 1 THEN amount ELSE 0 END), 0) AS INTEGER) AS paid").
		Scan(ctx, &totals)
	return totals, err
}
