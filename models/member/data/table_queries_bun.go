package data

import (
	"context"
	"database/sql"
	"strings"

	"bandcash/internal/db"
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

type MemberEventFilter struct {
	MemberID string
	GroupID  string
	Search   string
	Year     string
	From     string
	To       string
}

type MemberEventRow struct {
	ID                 string         `bun:"id"`
	GroupID            string         `bun:"group_id"`
	Title              string         `bun:"title"`
	Time               string         `bun:"time"`
	Description        string         `bun:"description"`
	Amount             int64          `bun:"amount"`
	ParticipantAmount  int64          `bun:"participant_amount"`
	ParticipantExpense int64          `bun:"participant_expense"`
	ParticipantPaid    int64          `bun:"participant_paid"`
	ParticipantPaidAt  sql.NullString `bun:"participant_paid_at"`
}

type MemberEventTotals struct {
	TotalCut     int64 `bun:"total_cut"`
	TotalExpense int64 `bun:"total_expense"`
	TotalPayout  int64 `bun:"total_payout"`
	TotalPaid    int64 `bun:"total_paid"`
	TotalUnpaid  int64 `bun:"total_unpaid"`
}

type MemberEventListParams struct {
	MemberEventFilter
	Sort   string
	Dir    string
	Limit  int
	Offset int
}

func CountMembersTable(ctx context.Context, groupID, search string) (int64, error) {
	q := db.BunDB.NewSelect().TableExpr("members").Where("group_id = ?", groupID)
	q = applySearch(q, search, func(sq *bun.SelectQuery, s string) *bun.SelectQuery {
		like := "%" + s + "%"
		return sq.WhereGroup(" AND ", func(qq *bun.SelectQuery) *bun.SelectQuery {
			return qq.Where("name LIKE ?", like).WhereOr("description LIKE ?", like)
		})
	})
	n, err := q.Count(ctx)
	return int64(n), err
}

func ListMembersTable(ctx context.Context, params MemberTableListParams) ([]db.Member, error) {
	rows := make([]db.Member, 0)
	q := db.BunDB.NewSelect().Model(&rows).Where("group_id = ?", params.GroupID)
	q = applySearch(q, params.Search, func(sq *bun.SelectQuery, s string) *bun.SelectQuery {
		like := "%" + s + "%"
		return sq.WhereGroup(" AND ", func(qq *bun.SelectQuery) *bun.SelectQuery {
			return qq.Where("name LIKE ?", like).WhereOr("description LIKE ?", like)
		})
	})

	d := normalizeDir(params.Dir)
	switch params.Sort {
	case "name":
		q = q.OrderExpr("name " + d)
	case "description":
		q = q.OrderExpr("description " + d)
	case "createdAt":
		q = q.OrderExpr("created_at " + d)
	default:
		q = q.OrderExpr("created_at DESC")
	}
	q = q.OrderExpr("created_at DESC")

	if params.Limit > 0 {
		q = q.Limit(params.Limit)
	}
	if params.Offset > 0 {
		q = q.Offset(params.Offset)
	}

	err := q.Scan(ctx)
	return rows, err
}

func CountMemberEventsTable(ctx context.Context, filter MemberEventFilter) (int64, error) {
	q := db.BunDB.NewSelect().
		TableExpr("events").
		Join("JOIN participants ON participants.event_id = events.id")
	q = applyMemberEventFilters(q, filter)
	n, err := q.Count(ctx)
	return int64(n), err
}

func SumMemberEventTotalsTable(ctx context.Context, filter MemberEventFilter) (MemberEventTotals, error) {
	var totals MemberEventTotals
	q := db.BunDB.NewSelect().
		TableExpr("events").
		ColumnExpr("CAST(COALESCE(SUM(participants.amount), 0) AS INTEGER) AS total_cut").
		ColumnExpr("CAST(COALESCE(SUM(participants.expense), 0) AS INTEGER) AS total_expense").
		ColumnExpr("CAST(COALESCE(SUM(participants.amount + participants.expense), 0) AS INTEGER) AS total_payout").
		ColumnExpr("CAST(COALESCE(SUM(CASE WHEN participants.paid = 1 THEN participants.amount + participants.expense ELSE 0 END), 0) AS INTEGER) AS total_paid").
		ColumnExpr("CAST(COALESCE(SUM(CASE WHEN participants.paid = 0 THEN participants.amount + participants.expense ELSE 0 END), 0) AS INTEGER) AS total_unpaid").
		Join("JOIN participants ON participants.event_id = events.id")
	q = applyMemberEventFilters(q, filter)
	err := q.Scan(ctx, &totals)
	return totals, err
}

func ListMemberEventsTable(ctx context.Context, params MemberEventListParams) ([]MemberEventRow, error) {
	rows := make([]MemberEventRow, 0)
	q := db.BunDB.NewSelect().
		TableExpr("events").
		ColumnExpr("events.id").
		ColumnExpr("events.group_id").
		ColumnExpr("events.title").
		ColumnExpr("events.time").
		ColumnExpr("events.description").
		ColumnExpr("events.amount").
		ColumnExpr("participants.amount AS participant_amount").
		ColumnExpr("participants.expense AS participant_expense").
		ColumnExpr("participants.paid AS participant_paid").
		ColumnExpr("participants.paid_at AS participant_paid_at").
		Join("JOIN participants ON participants.event_id = events.id")

	q = applyMemberEventFilters(q, params.MemberEventFilter)
	d := normalizeDir(params.Dir)
	switch params.Sort {
	case "title":
		q = q.OrderExpr("events.title " + d)
	case "time":
		q = q.OrderExpr("events.time " + d)
	case "participant_amount":
		q = q.OrderExpr("participants.amount " + d)
	case "participant_expense":
		q = q.OrderExpr("participants.expense " + d)
	case "paid":
		q = q.OrderExpr("participants.paid " + d)
	case "paid_at":
		q = q.OrderExpr("participants.paid_at " + d)
	default:
		q = q.OrderExpr("events.time DESC")
	}
	q = q.OrderExpr("events.id ASC")

	if params.Limit > 0 {
		q = q.Limit(params.Limit)
	}
	if params.Offset > 0 {
		q = q.Offset(params.Offset)
	}
	err := q.Scan(ctx, &rows)
	return rows, err
}

func applyMemberEventFilters(q *bun.SelectQuery, filter MemberEventFilter) *bun.SelectQuery {
	q = q.Where("participants.member_id = ?", filter.MemberID).
		Where("participants.group_id = ?", filter.GroupID)
	q = applySearch(q, filter.Search, func(sq *bun.SelectQuery, s string) *bun.SelectQuery {
		return sq.Where("(events.title LIKE '%' || ? || '%' OR events.description LIKE '%' || ? || '%')", s, s)
	})
	return applyDateRangeOrYear(q, filter.From, filter.To, filter.Year, "events.time")
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
		return q.Where(columnExpr+" LIKE ? || '%'", year)
	}
	return q
}

func normalizeDir(dir string) string {
	if strings.EqualFold(strings.TrimSpace(dir), "asc") {
		return "ASC"
	}
	return "DESC"
}
