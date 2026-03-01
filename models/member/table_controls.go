package member

import (
	"fmt"

	"bandcash/internal/utils"
)

var tablePageSizes = utils.StandardTablePageSizes

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

func memberSearchAction(data MembersData) string {
	return utils.BuildTableSearchDatastarAction(memberIndexPath(data.GroupID), 50)
}
