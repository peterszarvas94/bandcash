package expense

import (
	"context"
	"sort"
	"strings"

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

func matchesExpenseFilters(expense db.Expense, query utils.TableQuery) bool {
	if query.Search != "" {
		searchLower := strings.ToLower(query.Search)
		if !strings.Contains(strings.ToLower(expense.Title), searchLower) &&
			!strings.Contains(strings.ToLower(expense.Description), searchLower) {
			return false
		}
	}

	if query.Year != "" && !strings.HasPrefix(expense.Date, query.Year) {
		return false
	}

	if query.From != "" && expense.Date < query.From {
		return false
	}
	if query.To != "" && expense.Date > query.To {
		return false
	}

	return true
}

func sortExpenses(expenses []db.Expense, sortField, dir string) {
	less := func(i, j int) bool {
		switch sortField {
		case "title":
			if dir == "desc" {
				return expenses[i].Title > expenses[j].Title
			}
			return expenses[i].Title < expenses[j].Title
		case "amount":
			if dir == "desc" {
				return expenses[i].Amount > expenses[j].Amount
			}
			return expenses[i].Amount < expenses[j].Amount
		case "paid":
			if dir == "desc" {
				return expenses[i].Paid > expenses[j].Paid
			}
			return expenses[i].Paid < expenses[j].Paid
		case "paid_at":
			if expenses[i].PaidAt.Valid && expenses[j].PaidAt.Valid {
				if dir == "desc" {
					return expenses[i].PaidAt.String > expenses[j].PaidAt.String
				}
				return expenses[i].PaidAt.String < expenses[j].PaidAt.String
			}
			if expenses[i].PaidAt.Valid != expenses[j].PaidAt.Valid {
				if dir == "desc" {
					return !expenses[i].PaidAt.Valid
				}
				return expenses[i].PaidAt.Valid
			}
			if dir == "asc" {
				return expenses[i].Date < expenses[j].Date
			}
			return expenses[i].Date > expenses[j].Date
		default: // date
			if dir == "asc" {
				return expenses[i].Date < expenses[j].Date
			}
			return expenses[i].Date > expenses[j].Date
		}
	}
	sort.Slice(expenses, less)
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
		Title:   "Bandcash - " + expense.Title,
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
