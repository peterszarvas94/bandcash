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
	return utils.BuildTableSortURL(memberIndexPath(data.GroupID), data.Query, column)
}

func memberSortDir(data MembersData, column string) string {
	if data.Query.SortSet && data.Query.Sort == column {
		return data.Query.Dir
	}
	return ""
}

func memberPageURL(data MembersData, page int) string {
	return utils.BuildTablePageURL(memberIndexPath(data.GroupID), data.Query, page, data.Pager.TotalPages)
}

func memberPageSizeURL(data MembersData, pageSize int) string {
	return utils.BuildTablePageSizeURL(memberIndexPath(data.GroupID), data.Query, pageSize)
}

func memberSearchAction(data MembersData) string {
	return utils.BuildTableSearchDatastarAction(memberIndexPath(data.GroupID), utils.DefaultTablePageSize)
}
