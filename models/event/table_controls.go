package event

import (
	"fmt"

	"bandcash/internal/utils"
)

var tablePageSizes = utils.StandardTablePageSizes

func eventIndexPath(groupID string) string {
	return fmt.Sprintf("/groups/%s/events", groupID)
}

func eventSortURL(data EventsData, column string) string {
	return utils.BuildTableSortURL(eventIndexPath(data.GroupID), data.Query, column)
}

func eventSortDir(data EventsData, column string) string {
	if data.Query.SortSet && data.Query.Sort == column {
		return data.Query.Dir
	}
	return ""
}

func eventPageURL(data EventsData, page int) string {
	return utils.BuildTablePageURL(eventIndexPath(data.GroupID), data.Query, page, data.Pager.TotalPages)
}

func eventPageSizeURL(data EventsData, pageSize int) string {
	return utils.BuildTablePageSizeURL(eventIndexPath(data.GroupID), data.Query, pageSize)
}

func eventSearchAction(data EventsData) string {
	return utils.BuildTableSearchDatastarAction(eventIndexPath(data.GroupID), utils.DefaultTablePageSize)
}
