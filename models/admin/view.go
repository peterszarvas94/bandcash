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
	Tab         string

	// Overview stats (always present)
	UsersCount    int64
	GroupsCount   int64
	EventsCount   int64
	MembersCount  int64
	SignupEnabled bool

	// Users tab data
	Users     []RecentUserRow
	UserPager utils.TablePagination
	UserQuery utils.TableQuery

	// Groups tab data
	Groups     []db.Group
	GroupPager utils.TablePagination
	GroupQuery utils.TableQuery
}
