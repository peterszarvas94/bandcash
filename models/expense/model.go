package expense

import (
	"context"

	ctxi18n "github.com/invopop/ctxi18n/i18n"

	"bandcash/internal/db"
	"bandcash/internal/utils"
	expensestore "bandcash/models/expense/data"
	groupstore "bandcash/models/group/data"
)

func TableQuerySpec() utils.TableQuerySpec {
	return utils.StandardTableQuerySpec(utils.StandardTableQuerySpecParams{
		DefaultSort:  "date",
		DefaultDir:   "desc",
		AllowedSorts: []string{"date", "title", "amount", "paid", "paid_at"},
	})
}

func GetIndexData(ctx context.Context, groupID string, query utils.TableQuery) (ExpensesData, error) {
	group, err := groupstore.GetGroupByID(ctx, groupID)
	if err != nil {
		return ExpensesData{}, err
	}

	filters := expensestore.ExpenseTableFilter{
		GroupID: groupID,
		Search:  query.Search,
		Year:    query.Year,
		From:    query.From,
		To:      query.To,
	}

	totalItems, err := expensestore.CountExpensesTable(ctx, filters)
	if err != nil {
		return ExpensesData{}, err
	}
	query = utils.ClampPage(query, totalItems)

	expenses, err := expensestore.ListExpensesTable(ctx, expensestore.ExpenseTableListParams{
		ExpenseTableFilter: filters,
		Sort:               query.Sort,
		Dir:                query.Dir,
		Limit:              query.PageSize,
		Offset:             int(query.Offset()),
	})
	if err != nil {
		return ExpensesData{}, err
	}

	totalRows, err := expensestore.SumExpenseTotalsTable(ctx, filters)
	if err != nil {
		return ExpensesData{}, err
	}

	totals := expenseCalcTotals{
		TotalItems: totalItems,
		Total:      totalRows.Total,
		Paid:       totalRows.Paid,
		Unpaid:     totalRows.Total - totalRows.Paid,
	}

	return buildExpensesData(ctx, groupID, group, query, expenses, totals)
}

type expenseCalcTotals struct {
	TotalItems int64
	Total      int64
	Paid       int64
	Unpaid     int64
}

func buildExpensesData(ctx context.Context, groupID string, group db.Group, query utils.TableQuery, expenses []db.Expense, totals expenseCalcTotals) (ExpensesData, error) {
	query = utils.ClampPage(query, totals.TotalItems)

	// Calculate group totals for display
	groupTotals, err := utils.CalculateGroupTotals(ctx, groupID)
	if err != nil {
		return ExpensesData{}, err
	}

	return ExpensesData{
		Title:              ctxi18n.T(ctx, "expenses.page_title"),
		GroupName:          group.Name,
		Expenses:           expenses,
		RecentYears:        utils.RecentYears(3),
		Query:              query,
		Pager:              utils.BuildTablePagination(totals.TotalItems, query),
		GroupID:            groupID,
		TotalExpenseAmount: groupTotals.Expenses.All,
		TotalPaid:          groupTotals.Expenses.Paid,
		TotalUnpaid:        groupTotals.Expenses.Unpaid,
		FilteredTotal:      totals.Total,
		FilteredPaid:       totals.Paid,
		FilteredUnpaid:     totals.Unpaid,
		Breadcrumbs: []utils.Crumb{
			{Label: ctxi18n.T(ctx, "groups.title"), Href: "/groups"},
			{Label: group.Name, Href: "/groups/" + groupID + "/events"},
			{Label: ctxi18n.T(ctx, "expenses.title")},
		},
		ExpensesTable: ExpensesIndexTableLayout(),
	}, nil
}

func GetShowData(ctx context.Context, groupID, expenseID string) (ExpenseData, error) {
	group, err := groupstore.GetGroupByID(ctx, groupID)
	if err != nil {
		return ExpenseData{}, err
	}

	expense, err := expensestore.GetExpense(ctx, expensestore.GetExpenseParams{
		ID:      expenseID,
		GroupID: groupID,
	})
	if err != nil {
		return ExpenseData{}, err
	}

	return ExpenseData{
		Title:   "bandcash - " + expense.Title,
		Expense: &expense,
		GroupID: groupID,
		Breadcrumbs: []utils.Crumb{
			{Label: ctxi18n.T(ctx, "groups.title"), Href: "/groups"},
			{Label: group.Name, Href: "/groups/" + groupID + "/events"},
			{Label: ctxi18n.T(ctx, "expenses.title"), Href: "/groups/" + groupID + "/expenses"},
			{Label: expense.Title},
		},
	}, nil
}
