package expense

import (
	"bandcash/internal/db"
	"bandcash/internal/utils"
)

type ExpensesData struct {
	Title              string
	GroupName          string
	Expenses           []db.Expense
	RecentYears        []int
	Query              utils.TableQuery
	Pager              utils.TablePagination
	Breadcrumbs        []utils.Crumb
	Signals            map[string]any
	GroupID            string
	IsAdmin            bool
	IsAuthenticated    bool
	IsSuperAdmin       bool
	TotalExpenseAmount int64
	TotalPaid          int64
	TotalUnpaid        int64
	FilteredTotal      int64
	FilteredPaid       int64
	FilteredUnpaid     int64
	ExpensesTable      utils.TableLayout
}

type ExpenseData struct {
	Title           string
	Expense         *db.Expense
	Breadcrumbs     []utils.Crumb
	Signals         map[string]any
	GroupID         string
	IsAdmin         bool
	IsAuthenticated bool
	IsSuperAdmin    bool
}

type NewExpensePageData struct {
	Title           string
	Breadcrumbs     []utils.Crumb
	GroupID         string
	Signals         map[string]any
	IsAuthenticated bool
	IsSuperAdmin    bool
}

type EditExpensePageData struct {
	Title           string
	Breadcrumbs     []utils.Crumb
	GroupID         string
	Expense         *db.Expense
	Signals         map[string]any
	IsAuthenticated bool
	IsSuperAdmin    bool
}
