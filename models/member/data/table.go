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

type MemberTableRow struct {
	ID          string       `bun:"id"`
	GroupID     string       `bun:"group_id"`
	Name        string       `bun:"name"`
	Description string       `bun:"description"`
	CreatedAt   sql.NullTime `bun:"created_at"`
	UpdatedAt   sql.NullTime `bun:"updated_at"`
	Unpaid      int64        `bun:"unpaid"`
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

func ListMembersTable(ctx context.Context, params MemberTableListParams) ([]MemberTableRow, error) {
	rows := make([]MemberTableRow, 0)
	q := db.BunDB.NewSelect().
		TableExpr("members").
		ColumnExpr("members.id").
		ColumnExpr("members.group_id").
		ColumnExpr("members.name").
		ColumnExpr("members.description").
		ColumnExpr("members.created_at").
		ColumnExpr("members.updated_at").
		ColumnExpr("COALESCE(SUM(CASE WHEN participants.paid = 0 THEN participants.amount + participants.expense ELSE 0 END), 0) AS unpaid").
		Join("LEFT JOIN participants ON participants.member_id = members.id AND participants.group_id = members.group_id").
		Where("members.group_id = ?", params.GroupID)
	q = applySearch(q, params.Search, func(sq *bun.SelectQuery, s string) *bun.SelectQuery {
		like := "%" + s + "%"
		return sq.WhereGroup(" AND ", func(qq *bun.SelectQuery) *bun.SelectQuery {
			return qq.Where("members.name LIKE ?", like).WhereOr("members.description LIKE ?", like)
		})
	})
	q = q.GroupExpr("members.id").GroupExpr("members.group_id").GroupExpr("members.name").GroupExpr("members.description").GroupExpr("members.created_at").GroupExpr("members.updated_at")

	d := normalizeDir(params.Dir)
	switch params.Sort {
	case "name":
		q = q.OrderExpr("members.name " + d)
	case "description":
		q = q.OrderExpr("members.description " + d)
	case "unpaid":
		q = q.OrderExpr("unpaid " + d)
	case "createdAt":
		q = q.OrderExpr("members.created_at " + d)
	default:
		q = q.OrderExpr("members.created_at DESC")
	}
	q = q.OrderExpr("members.created_at DESC")

	if params.Limit > 0 {
		q = q.Limit(params.Limit)
	}
	if params.Offset > 0 {
		q = q.Offset(params.Offset)
	}

	err := q.Scan(ctx, &rows)
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
	rows := make([]struct {
		Amount  int64 `bun:"amount"`
		Expense int64 `bun:"expense"`
		Paid    int64 `bun:"paid"`
	}, 0)
	q := db.BunDB.NewSelect().
		TableExpr("events").
		ColumnExpr("participants.amount").
		ColumnExpr("participants.expense").
		ColumnExpr("participants.paid").
		Join("JOIN participants ON participants.event_id = events.id")
	q = applyMemberEventFilters(q, filter)
	if err := q.Scan(ctx, &rows); err != nil {
		return MemberEventTotals{}, err
	}

	totals := MemberEventTotals{}
	for _, row := range rows {
		totals.TotalCut += row.Amount
		totals.TotalExpense += row.Expense
		payout := row.Amount + row.Expense
		totals.TotalPayout += payout
		if row.Paid == 1 {
			totals.TotalPaid += payout
		} else {
			totals.TotalUnpaid += payout
		}
	}
	return totals, nil
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
