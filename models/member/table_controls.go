package member

import (
	"fmt"

	"bandcash/internal/utils"
)

var tablePageSizes = []int{10, 50, 100, 200}

func memberIndexPath(groupID string) string {
	return fmt.Sprintf("/groups/%s/members", groupID)
}

func memberSortURL(data MembersData, column string) string {
	next := utils.NextSortCycle(data.Query, column)
	page := 1
	return utils.BuildTableQueryURLWith(memberIndexPath(data.GroupID), data.Query, utils.TableQueryPatch{
		Sort: &next.Sort,
		Dir:  &next.Dir,
		Page: &page,
	})
}

func memberSortDir(data MembersData, column string) string {
	if data.Query.SortSet && data.Query.Sort == column {
		return data.Query.Dir
	}
	return ""
}

func memberPageURL(data MembersData, page int) string {
	if page < 1 {
		page = 1
	}
	if page > data.Pager.TotalPages {
		page = data.Pager.TotalPages
	}
	return utils.BuildTableQueryURLWith(memberIndexPath(data.GroupID), data.Query, utils.TableQueryPatch{Page: &page})
}

func memberPageSizeURL(data MembersData, pageSize int) string {
	page := 1
	return utils.BuildTableQueryURLWith(memberIndexPath(data.GroupID), data.Query, utils.TableQueryPatch{
		Page:     &page,
		PageSize: &pageSize,
	})
}

func pageSizeButtonClass(current, value int) string {
	if current == value {
		return "btn btn-sm btn-active"
	}
	return "btn btn-sm"
}

func memberSearchAction(data MembersData) string {
	return fmt.Sprintf("$tableQuery.page = 1; const params = new URLSearchParams(); if (($tableQuery.search || '').trim() !== '') { params.set('q', ($tableQuery.search || '').trim()) }; if ($tableQuery.sort) { params.set('sort', $tableQuery.sort); params.set('dir', $tableQuery.dir || 'asc') }; if (Number($tableQuery.pageSize || 50) !== 50) { params.set('pageSize', String($tableQuery.pageSize)) }; const url = '%s' + (params.toString() ? '?' + params.toString() : ''); history.pushState(null, '', url); @get(url)", memberIndexPath(data.GroupID))
}

func memberQueryAction(url string) string {
	return fmt.Sprintf("history.pushState(null, '', '%s'); @get('%s')", url, url)
}
