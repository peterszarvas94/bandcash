package group

import "bandcash/internal/utils"

var tablePageSizes = utils.StandardTablePageSizes

func groupsIndexPath() string {
	return "/dashboard"
}

func groupsSortURL(data GroupsPageData, column string) string {
	return utils.BuildTableSortURL(groupsIndexPath(), data.Query, column)
}

func groupsSortDir(data GroupsPageData, column string) string {
	if data.Query.SortSet && data.Query.Sort == column {
		return data.Query.Dir
	}
	return ""
}

func groupsPageURL(data GroupsPageData, page int) string {
	return utils.BuildTablePageURL(groupsIndexPath(), data.Query, page, data.Pagination.TotalPages)
}

func groupsPageSizeURL(data GroupsPageData, pageSize int) string {
	return utils.BuildTablePageSizeURL(groupsIndexPath(), data.Query, pageSize)
}

func groupsSearchAction(data GroupsPageData) string {
	return utils.BuildTableSearchDatastarAction(groupsIndexPath(), utils.DefaultTablePageSize)
}
