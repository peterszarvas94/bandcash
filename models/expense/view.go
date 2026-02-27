package expense

import (
	"bandcash/internal/db"
	"bandcash/internal/utils"
)

type ExpensesData struct {
	Title       string
	Expenses    []db.Expense
	Breadcrumbs []utils.Crumb
	GroupID     string
	IsAdmin     bool
	UserEmail   string
}
