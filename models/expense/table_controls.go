package expense

import (
	"fmt"

	"bandcash/internal/utils"
)

var tablePageSizes = utils.StandardTablePageSizes

func expenseIndexPath(groupID string) string {
	return fmt.Sprintf("/groups/%s/expenses", groupID)
}

func expenseSortURL(data ExpensesData, column string) string {
	return utils.BuildTableSortURL(expenseIndexPath(data.GroupID), data.Query, column)
}

func expenseSortDir(data ExpensesData, column string) string {
	if data.Query.SortSet && data.Query.Sort == column {
		return data.Query.Dir
	}
	return ""
}

func expensePageURL(data ExpensesData, page int) string {
	return utils.BuildTablePageURL(expenseIndexPath(data.GroupID), data.Query, page, data.Pager.TotalPages)
}

func expensePageSizeURL(data ExpensesData, pageSize int) string {
	return utils.BuildTablePageSizeURL(expenseIndexPath(data.GroupID), data.Query, pageSize)
}

func expenseSearchAction(data ExpensesData) string {
	return utils.BuildTableSearchDatastarAction(expenseIndexPath(data.GroupID), utils.DefaultTablePageSize)
}
