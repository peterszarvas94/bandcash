package admin

import (
	"database/sql"

	"bandcash/internal/db"
	"bandcash/internal/utils"
)

type RecentUserRow struct {
	ID        string
	Email     string
	CreatedAt sql.NullTime
	IsBanned  bool
}

type DashboardData struct {
	Title       string
	Breadcrumbs []utils.Crumb
	UserEmail   string

	UsersCount   int64
	GroupsCount  int64
	EventsCount  int64
	MembersCount int64

	SignupEnabled bool

	RecentUsers  []RecentUserRow
	RecentGroups []db.Group
}
