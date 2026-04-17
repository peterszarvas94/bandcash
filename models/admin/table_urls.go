package admin

import (
	"bandcash/internal/utils"
)

func adminUsersSortURL(data DashboardData, column string) string {
	return utils.BuildTableSortURL("/admin/users", data.UserQuery, column)
}

func adminGroupsSortURL(data DashboardData, column string) string {
	return utils.BuildTableSortURL("/admin/groups", data.GroupQuery, column)
}

func adminSessionsSortURL(data DashboardData, column string) string {
	return utils.BuildTableSortURL("/admin/sessions", data.SessionQuery, column)
}
