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
	next := utils.NextSortCycle(data.Query, column)
	page := 1
	return utils.BuildTableQueryURLWith(expenseIndexPath(data.GroupID), data.Query, utils.TableQueryPatch{
		Sort: &next.Sort,
		Dir:  &next.Dir,
		Page: &page,
	})
}

func expenseSortDir(data ExpensesData, column string) string {
	if data.Query.SortSet && data.Query.Sort == column {
		return data.Query.Dir
	}
	return ""
}

func expensePageURL(data ExpensesData, page int) string {
	if page < 1 {
		page = 1
	}
	if page > data.Pager.TotalPages {
		page = data.Pager.TotalPages
	}
	return utils.BuildTableQueryURLWith(expenseIndexPath(data.GroupID), data.Query, utils.TableQueryPatch{Page: &page})
}

func expensePageSizeURL(data ExpensesData, pageSize int) string {
	page := 1
	return utils.BuildTableQueryURLWith(expenseIndexPath(data.GroupID), data.Query, utils.TableQueryPatch{
		Page:     &page,
		PageSize: &pageSize,
	})
}

func expenseSearchAction(data ExpensesData) string {
	return utils.BuildTableSearchDatastarAction(expenseIndexPath(data.GroupID), 50)
}
