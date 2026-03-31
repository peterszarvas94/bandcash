package admin

import (
	"fmt"

	"bandcash/internal/utils"
)

func adminUsersSortURL(data DashboardData, column string) string {
	return utils.BuildTableSortURL("/admin/users", data.UserQuery, column)
}

func adminGroupsSortURL(data DashboardData, column string) string {
	return utils.BuildTableSortURL("/admin/groups", data.GroupQuery, column)
}

func adminPageSizeURL(query utils.TableQuery, tab string, pageSize int) string {
	return utils.BuildTablePageSizeURL(fmt.Sprintf("/admin/%s", tab), query, pageSize)
}

func adminSessionsSortURL(data DashboardData, column string) string {
	return utils.BuildTableSortURL("/admin/sessions", data.SessionQuery, column)
}
