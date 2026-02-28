package expense

import (
	"fmt"

	"bandcash/internal/utils"
)

var tablePageSizes = []int{10, 50, 100, 200}

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

func pageSizeButtonClass(current, value int) string {
	if current == value {
		return "btn btn-sm btn-primary"
	}
	return "btn btn-sm"
}

func expenseSearchAction(data ExpensesData) string {
	return fmt.Sprintf("$tableQuery.page = 1; const params = new URLSearchParams(); if (($tableQuery.search || '').trim() !== '') { params.set('q', ($tableQuery.search || '').trim()) }; if ($tableQuery.sort) { params.set('sort', $tableQuery.sort); params.set('dir', $tableQuery.dir || 'asc') }; if (Number($tableQuery.pageSize || 50) !== 50) { params.set('pageSize', String($tableQuery.pageSize)) }; const url = '%s' + (params.toString() ? '?' + params.toString() : ''); history.pushState(null, '', url); @get(url)", expenseIndexPath(data.GroupID))
}

func expenseQueryAction(url string) string {
	return fmt.Sprintf("history.pushState(null, '', '%s'); @get('%s')", url, url)
}
