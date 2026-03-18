package expense

import (
	"bandcash/internal/db"
	"bandcash/internal/utils"
)

type ExpensesData struct {
	Title              string
	Expenses           []db.Expense
	RecentYears        []int
	Query              utils.TableQuery
	Pager              utils.TablePagination
	Breadcrumbs        []utils.Crumb
	GroupID            string
	IsAdmin            bool
	UserEmail          string
	TotalExpenseAmount int64
	FilteredTotal      int64
	FilteredPaid       int64
	FilteredUnpaid     int64
	ExpensesTable      utils.TableLayout
}

type ExpenseData struct {
	Title       string
	Expense     *db.Expense
	Breadcrumbs []utils.Crumb
	GroupID     string
	IsAdmin     bool
	UserEmail   string
}
