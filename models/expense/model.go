package expense

import (
	"context"
	"sort"
	"strings"

	ctxi18n "github.com/invopop/ctxi18n/i18n"

	"bandcash/internal/db"
	"bandcash/internal/utils"
)

type Expenses struct{}

func (e *Expenses) TableQuerySpec() utils.TableQuerySpec {
	return utils.StandardTableQuerySpec(utils.StandardTableQuerySpecParams{
		DefaultSort:  "date",
		DefaultDir:   "desc",
		AllowedSorts: []string{"date", "title", "amount"},
	})
}

func New() *Expenses {
	return &Expenses{}
}

func (e *Expenses) GetIndexData(ctx context.Context, groupID string, query utils.TableQuery) (ExpensesData, error) {
	group, err := db.Qry.GetGroupByID(ctx, groupID)
	if err != nil {
		return ExpensesData{}, err
	}

	// Check cache first
	cacheKey := utils.ExpensesFilterKey(groupID, query.Search, query.Year, query.From, query.To)
	if cached, ok := utils.CalcCacheInstance.Get(cacheKey); ok {
		if result, valid := cached.(expenseCalcTotals); valid {
			return e.buildExpensesData(ctx, groupID, group, query, result)
		}
	}

	// Get all expenses for the group to calculate in-memory
	allExpenses, err := db.Qry.ListExpenses(ctx, groupID)
	if err != nil {
		return ExpensesData{}, err
	}

	// Filter and calculate totals in-memory
	filteredExpenses := make([]db.Expense, 0, len(allExpenses))
	var filteredTotal, totalPaid, totalUnpaid int64

	for _, expense := range allExpenses {
		if matchesExpenseFilters(expense, query) {
			filteredExpenses = append(filteredExpenses, expense)
			filteredTotal += expense.Amount
			if expense.Paid == 1 {
				totalPaid += expense.Amount
			} else {
				totalUnpaid += expense.Amount
			}
		}
	}

	// Sort filtered expenses
	sortExpenses(filteredExpenses, query.Sort, query.Dir)

	// Store in cache
	totals := expenseCalcTotals{
		Filtered: filteredExpenses,
		Total:    filteredTotal,
		Paid:     totalPaid,
		Unpaid:   totalUnpaid,
	}
	utils.CalcCacheInstance.Set(cacheKey, totals)

	return e.buildExpensesData(ctx, groupID, group, query, totals)
}

type expenseCalcTotals struct {
	Filtered []db.Expense
	Total    int64
	Paid     int64
	Unpaid   int64
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
		default: // date
			if dir == "asc" {
				return expenses[i].Date < expenses[j].Date
			}
			return expenses[i].Date > expenses[j].Date
		}
	}
	sort.Slice(expenses, less)
}

func (e *Expenses) buildExpensesData(ctx context.Context, groupID string, group db.Group, query utils.TableQuery, totals expenseCalcTotals) (ExpensesData, error) {
	totalItems := int64(len(totals.Filtered))
	query = utils.ClampPage(query, totalItems)

	start := query.Offset()
	end := start + int64(query.PageSize)
	if end > totalItems {
		end = totalItems
	}
	if start > totalItems {
		start = totalItems
	}

	var paginatedExpenses []db.Expense
	if start < totalItems {
		paginatedExpenses = totals.Filtered[start:end]
	}

	// Calculate group totals for display
	groupTotals, err := utils.CalculateGroupTotals(ctx, groupID)
	if err != nil {
		return ExpensesData{}, err
	}

	return ExpensesData{
		Title:              ctxi18n.T(ctx, "expenses.page_title"),
		Expenses:           paginatedExpenses,
		RecentYears:        utils.RecentYears(3),
		Query:              query,
		Pager:              utils.BuildTablePagination(totalItems, query),
		GroupID:            groupID,
		TotalExpenseAmount: groupTotals.TotalExpenseAmount,
		FilteredTotal:      totals.Total,
		FilteredPaid:       totals.Paid,
		FilteredUnpaid:     totals.Unpaid,
		Breadcrumbs: []utils.Crumb{
			{Label: ctxi18n.T(ctx, "groups.title"), Href: "/dashboard"},
			{Label: group.Name, Href: "/groups/" + groupID},
			{Label: ctxi18n.T(ctx, "expenses.title")},
		},
		ExpensesTable: utils.ExpensesIndexTableLayout(),
	}, nil
}

func (e *Expenses) GetShowData(ctx context.Context, groupID, expenseID string) (ExpenseData, error) {
	group, err := db.Qry.GetGroupByID(ctx, groupID)
	if err != nil {
		return ExpenseData{}, err
	}

	expense, err := db.Qry.GetExpense(ctx, db.GetExpenseParams{
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
			{Label: ctxi18n.T(ctx, "groups.title"), Href: "/dashboard"},
			{Label: group.Name, Href: "/groups/" + groupID},
			{Label: ctxi18n.T(ctx, "expenses.title"), Href: "/groups/" + groupID + "/expenses"},
			{Label: expense.Title},
		},
	}, nil
}
