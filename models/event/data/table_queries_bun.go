package data

import (
	"context"
	"strings"

	"bandcash/internal/db"
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

type ParticipantGroupTotals struct {
	TotalPaid   int64 `bun:"total_paid"`
	TotalUnpaid int64 `bun:"total_unpaid"`
}

func CountEventsTable(ctx context.Context, filter EventTableFilter) (int64, error) {
	q := db.BunDB.NewSelect().TableExpr("events")
	q = applyEventTableFilters(q, filter)
	n, err := q.Count(ctx)
	return int64(n), err
}

func ListEventsTable(ctx context.Context, params EventTableListParams) ([]db.Event, error) {
	rows := make([]db.Event, 0)
	q := db.BunDB.NewSelect().Model(&rows)
	q = applyEventTableFilters(q, params.EventTableFilter)
	q = orderEvents(q, params.Sort, params.Dir)
	if params.Limit > 0 {
		q = q.Limit(params.Limit)
	}
	if params.Offset > 0 {
		q = q.Offset(params.Offset)
	}
	err := q.Scan(ctx)
	return rows, err
}

func SumEventIncomeTotalsTable(ctx context.Context, filter EventTableFilter) (EventIncomeTotals, error) {
	var totals EventIncomeTotals
	q := db.BunDB.NewSelect().TableExpr("events")
	q = applyEventTableFilters(q, filter)
	err := q.ColumnExpr("CAST(COALESCE(SUM(amount), 0) AS INTEGER) AS total").
		ColumnExpr("CAST(COALESCE(SUM(CASE WHEN paid = 1 THEN amount ELSE 0 END), 0) AS INTEGER) AS paid").
		Scan(ctx, &totals)
	return totals, err
}

func SumParticipantTotalsByGroupTable(ctx context.Context, filter EventTableFilter) (ParticipantGroupTotals, error) {
	var totals ParticipantGroupTotals
	q := db.BunDB.NewSelect().
		TableExpr("participants").
		ColumnExpr("CAST(COALESCE(SUM(CASE WHEN participants.paid = 1 THEN participants.amount + participants.expense ELSE 0 END), 0) AS INTEGER) AS total_paid").
		ColumnExpr("CAST(COALESCE(SUM(CASE WHEN participants.paid = 0 THEN participants.amount + participants.expense ELSE 0 END), 0) AS INTEGER) AS total_unpaid").
		Join("JOIN events ON events.id = participants.event_id").
		Where("participants.group_id = ?", filter.GroupID)

	q = applySearch(q, filter.Search, func(sq *bun.SelectQuery, search string) *bun.SelectQuery {
		return sq.Where("(events.title LIKE '%' || ? || '%' OR events.description LIKE '%' || ? || '%')", search, search)
	})
	q = applyDateRangeOrYear(q, filter.From, filter.To, filter.Year, "events.time")

	err := q.Scan(ctx, &totals)
	return totals, err
}

func applyEventTableFilters(q *bun.SelectQuery, filter EventTableFilter) *bun.SelectQuery {
	q = q.Where("group_id = ?", filter.GroupID)
	q = applySearch(q, filter.Search, func(sq *bun.SelectQuery, search string) *bun.SelectQuery {
		like := "%" + search + "%"
		return sq.WhereGroup(" AND ", func(qq *bun.SelectQuery) *bun.SelectQuery {
			return qq.Where("title LIKE ?", like).WhereOr("place LIKE ?", like)
		})
	})
	return applyDateRangeOrYear(q, filter.From, filter.To, filter.Year, "COALESCE(NULLIF(date, ''), substr(time, 1, 10))")
}

func orderEvents(q *bun.SelectQuery, sort, dir string) *bun.SelectQuery {
	d := normalizeDir(dir)
	switch sort {
	case "date":
		q = q.OrderExpr("COALESCE(NULLIF(date, ''), substr(time, 1, 10)) " + d)
	case "time":
		q = q.OrderExpr("COALESCE(NULLIF(event_time, ''), substr(time, 12, 5)) " + d)
	case "title":
		q = q.OrderExpr("title " + d)
	case "place":
		q = q.OrderExpr("place " + d)
	case "amount":
		q = q.OrderExpr("amount " + d)
	case "description":
		q = q.OrderExpr("description " + d)
	case "paid":
		q = q.OrderExpr("paid " + d)
	case "paid_at":
		q = q.OrderExpr("(paid_at IS NULL OR paid_at = '') " + d).OrderExpr("paid_at " + d)
	default:
		q = q.OrderExpr("COALESCE(NULLIF(date, ''), substr(time, 1, 10)) DESC")
	}
	return q.OrderExpr("created_at DESC")
}

func applySearch(q *bun.SelectQuery, search string, fn func(*bun.SelectQuery, string) *bun.SelectQuery) *bun.SelectQuery {
	search = strings.TrimSpace(search)
	if search == "" {
		return q
	}
	return fn(q, search)
}

func applyDateRangeOrYear(q *bun.SelectQuery, from, to, year, columnExpr string) *bun.SelectQuery {
	from = strings.TrimSpace(from)
	to = strings.TrimSpace(to)
	year = strings.TrimSpace(year)
	if from != "" && to != "" {
		return q.Where(columnExpr+" >= ?", from).Where(columnExpr+" <= ?", to)
	}
	if year != "" {
		return q.Where("substr("+columnExpr+", 1, 4) = ?", year)
	}
	return q
}

func normalizeDir(dir string) string {
	if strings.EqualFold(strings.TrimSpace(dir), "asc") {
		return "ASC"
	}
	return "DESC"
}
