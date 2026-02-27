package admin

import (
	"bandcash/internal/db"
	"bandcash/internal/utils"
)

type DashboardData struct {
	Title       string
	Breadcrumbs []utils.Crumb
	UserEmail   string

	UsersCount   int64
	GroupsCount  int64
	EventsCount  int64
	MembersCount int64

	RecentUsers  []db.User
	RecentGroups []db.Group
}
