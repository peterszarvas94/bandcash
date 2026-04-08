package db

import (
	"context"

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

func applyExpenseTableFilters(q *bun.SelectQuery, filter ExpenseTableFilter) *bun.SelectQuery {
	q = q.Where("group_id = ?", filter.GroupID)

	q = applyOptionalSearch(q, filter.Search, func(sq *bun.SelectQuery, search string) *bun.SelectQuery {
		like := "%" + search + "%"
		return sq.WhereGroup(" AND ", func(qq *bun.SelectQuery) *bun.SelectQuery {
			return qq.Where("title LIKE ?", like).WhereOr("description LIKE ?", like)
		})
	})

	return applyDateRangeOrYear(
		q,
		filter.From,
		filter.To,
		filter.Year,
		func(sq *bun.SelectQuery, from, to string) *bun.SelectQuery {
			return sq.Where("date >= ?", from).Where("date <= ?", to)
		},
		func(sq *bun.SelectQuery, year string) *bun.SelectQuery {
			return sq.Where("substr(date, 1, 4) = ?", year)
		},
	)
}

var expenseTableOrderSpec = BunTableOrderSpec{
	DefaultSort: "date",
	DefaultDir:  "DESC",
	AllowedSorts: map[string]func(*bun.SelectQuery, string) *bun.SelectQuery{
		"date":   func(q *bun.SelectQuery, dir string) *bun.SelectQuery { return q.OrderExpr("date " + dir) },
		"title":  func(q *bun.SelectQuery, dir string) *bun.SelectQuery { return q.OrderExpr("title " + dir) },
		"amount": func(q *bun.SelectQuery, dir string) *bun.SelectQuery { return q.OrderExpr("amount " + dir) },
		"paid":   func(q *bun.SelectQuery, dir string) *bun.SelectQuery { return q.OrderExpr("paid " + dir) },
		"paid_at": func(q *bun.SelectQuery, dir string) *bun.SelectQuery {
			return q.OrderExpr("(paid_at IS NULL OR paid_at = '') " + dir).OrderExpr("paid_at " + dir)
		},
	},
	StableSort: func(q *bun.SelectQuery) *bun.SelectQuery { return q.OrderExpr("created_at DESC") },
}

func CountExpensesTable(ctx context.Context, filter ExpenseTableFilter) (int64, error) {
	q := BunDB.NewSelect().TableExpr("expenses")
	q = applyExpenseTableFilters(q, filter)
	count, err := q.Count(ctx)
	return int64(count), err
}

func ListExpensesTable(ctx context.Context, params ExpenseTableListParams) ([]Expense, error) {
	rows := make([]Expense, 0)
	q := BunDB.NewSelect().Model(&rows)
	q = applyExpenseTableFilters(q, params.ExpenseTableFilter)
	q = applyTableOrdering(q, expenseTableOrderSpec, params.Sort, params.Dir)
	q = applyTablePagination(q, params.Limit, params.Offset)

	err := q.Scan(ctx)
	return rows, err
}

func SumExpenseTotalsTable(ctx context.Context, filter ExpenseTableFilter) (ExpenseTotals, error) {
	var totals ExpenseTotals

	q := BunDB.NewSelect().TableExpr("expenses")
	q = applyExpenseTableFilters(q, filter)

	err := q.ColumnExpr("CAST(COALESCE(SUM(amount), 0) AS INTEGER) AS total").
		ColumnExpr("CAST(COALESCE(SUM(CASE WHEN paid = 1 THEN amount ELSE 0 END), 0) AS INTEGER) AS paid").
		Scan(ctx, &totals)
	return totals, err
}
