package admin

import (
	"database/sql"
	"time"

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
	Title           string
	Breadcrumbs     []utils.Crumb
	Tab             string
	Signals         map[string]any
	IsAuthenticated bool
	IsSuperAdmin    bool

	// Overview stats (always present)
	UsersCount                      int64
	GroupsCount                     int64
	EventsCount                     int64
	MembersCount                    int64
	SignupEnabled                   bool
	PaymentsEnabled                 bool
	BypassLimitForSuperadminEnabled bool

	// Users tab data
	Users     []RecentUserRow
	UserPager utils.TablePagination
	UserQuery utils.TableQuery

	// Groups tab data
	Groups     []db.Group
	GroupPager utils.TablePagination
	GroupQuery utils.TableQuery

	// Sessions tab data
	Sessions     []AdminSessionRow
	SessionPager utils.TablePagination
	SessionQuery utils.TableQuery

	UsersTable    utils.TableLayout
	GroupsTable   utils.TableLayout
	SessionsTable utils.TableLayout
}

type AdminSessionRow struct {
	ID        string
	UserID    string
	UserEmail string
	CreatedAt sql.NullTime
	ExpiresAt time.Time
}
