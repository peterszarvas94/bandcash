package data

import (
	"context"
	"strings"

	"bandcash/internal/db"
	"github.com/uptrace/bun"
)

type ExpenseTableFilter struct {
	GroupID string
	Search  string
	Year    string
	From    string
	To      string
}

type ExpenseTableListParams struct {
	ExpenseTableFilter
	Sort   string
	Dir    string
	Limit  int
	Offset int
}

type ExpenseTotals struct {
	Total int64
	Paid  int64
}

func CountExpensesTable(ctx context.Context, filter ExpenseTableFilter) (int64, error) {
	q := db.BunDB.NewSelect().TableExpr("expenses")
	q = applyExpenseTableFilters(q, filter)
	n, err := q.Count(ctx)
	return int64(n), err
}

func ListExpensesTable(ctx context.Context, params ExpenseTableListParams) ([]db.Expense, error) {
	rows := make([]db.Expense, 0)
	q := db.BunDB.NewSelect().Model(&rows)
	q = applyExpenseTableFilters(q, params.ExpenseTableFilter)
	q = orderExpenses(q, params.Sort, params.Dir)
	if params.Limit > 0 {
		q = q.Limit(params.Limit)
	}
	if params.Offset > 0 {
		q = q.Offset(params.Offset)
	}
	err := q.Scan(ctx)
	return rows, err
}

func SumExpenseTotalsTable(ctx context.Context, filter ExpenseTableFilter) (ExpenseTotals, error) {
	var totals ExpenseTotals
	q := db.BunDB.NewSelect().TableExpr("expenses")
	q = applyExpenseTableFilters(q, filter)
	err := q.ColumnExpr("CAST(COALESCE(SUM(amount), 0) AS INTEGER) AS total").
		ColumnExpr("CAST(COALESCE(SUM(CASE WHEN paid = 1 THEN amount ELSE 0 END), 0) AS INTEGER) AS paid").
		Scan(ctx, &totals)
	return totals, err
}

func applyExpenseTableFilters(q *bun.SelectQuery, filter ExpenseTableFilter) *bun.SelectQuery {
	q = q.Where("group_id = ?", filter.GroupID)
	q = applySearch(q, filter.Search, func(sq *bun.SelectQuery, search string) *bun.SelectQuery {
		like := "%" + search + "%"
		return sq.WhereGroup(" AND ", func(qq *bun.SelectQuery) *bun.SelectQuery {
			return qq.Where("title LIKE ?", like).WhereOr("description LIKE ?", like)
		})
	})
	return applyDateRangeOrYear(q, filter.From, filter.To, filter.Year, "date")
}

func orderExpenses(q *bun.SelectQuery, sort, dir string) *bun.SelectQuery {
	d := normalizeDir(dir)
	switch sort {
	case "date":
		q = q.OrderExpr("date " + d)
	case "title":
		q = q.OrderExpr("title " + d)
	case "amount":
		q = q.OrderExpr("amount " + d)
	case "paid":
		q = q.OrderExpr("paid " + d)
	case "paid_at":
		q = q.OrderExpr("(paid_at IS NULL OR paid_at = '') " + d).OrderExpr("paid_at " + d)
	default:
		q = q.OrderExpr("date DESC")
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
