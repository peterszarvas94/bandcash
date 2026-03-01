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
	next := utils.NextSortCycle(data.Query, column)
	page := 1
	return utils.BuildTableQueryURLWith(eventIndexPath(data.GroupID), data.Query, utils.TableQueryPatch{
		Sort: &next.Sort,
		Dir:  &next.Dir,
		Page: &page,
	})
}

func eventSortDir(data EventsData, column string) string {
	if data.Query.SortSet && data.Query.Sort == column {
		return data.Query.Dir
	}
	return ""
}

func eventPageURL(data EventsData, page int) string {
	if page < 1 {
		page = 1
	}
	if page > data.Pager.TotalPages {
		page = data.Pager.TotalPages
	}
	return utils.BuildTableQueryURLWith(eventIndexPath(data.GroupID), data.Query, utils.TableQueryPatch{Page: &page})
}

func eventPageSizeURL(data EventsData, pageSize int) string {
	page := 1
	return utils.BuildTableQueryURLWith(eventIndexPath(data.GroupID), data.Query, utils.TableQueryPatch{
		Page:     &page,
		PageSize: &pageSize,
	})
}

func eventSearchAction(data EventsData) string {
	return utils.BuildTableSearchDatastarAction(eventIndexPath(data.GroupID), 50)
}
