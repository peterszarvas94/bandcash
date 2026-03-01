package group

import "bandcash/internal/utils"

var tablePageSizes = utils.StandardTablePageSizes

func groupsIndexPath() string {
	return "/dashboard"
}

func groupsSortURL(data GroupsPageData, column string) string {
	next := utils.NextSortCycle(data.Query, column)
	page := 1
	return utils.BuildTableQueryURLWith(groupsIndexPath(), data.Query, utils.TableQueryPatch{
		Sort: &next.Sort,
		Dir:  &next.Dir,
		Page: &page,
	})
}

func groupsSortDir(data GroupsPageData, column string) string {
	if data.Query.SortSet && data.Query.Sort == column {
		return data.Query.Dir
	}
	return ""
}

func groupsPageURL(data GroupsPageData, page int) string {
	if page < 1 {
		page = 1
	}
	if page > data.Pagination.TotalPages {
		page = data.Pagination.TotalPages
	}
	return utils.BuildTableQueryURLWith(groupsIndexPath(), data.Query, utils.TableQueryPatch{Page: &page})
}

func groupsPageSizeURL(data GroupsPageData, pageSize int) string {
	page := 1
	return utils.BuildTableQueryURLWith(groupsIndexPath(), data.Query, utils.TableQueryPatch{
		Page:     &page,
		PageSize: &pageSize,
	})
}

func groupsSearchAction(data GroupsPageData) string {
	return utils.BuildTableSearchDatastarAction(groupsIndexPath(), 50)
}
