package group

import (
	"bandcash/internal/db"
	"bandcash/internal/utils"
)

type NewGroupPageData struct {
	Title       string
	Breadcrumbs []utils.Crumb
	UserEmail   string
}

type GroupWithRole struct {
	Group       db.Group
	Role        string
	ViewerCount int
	AdminEmail  string
}

type GroupsPageData struct {
	Title        string
	Breadcrumbs  []utils.Crumb
	UserEmail    string
	AllGroups    []GroupWithRole
	AdminGroups  []GroupSummary
	ReaderGroups []GroupSummary
	Query        utils.TableQuery
	Pagination   utils.TablePagination
}

type GroupSummary struct {
	Group       db.Group
	ViewerCount int
	AdminEmail  string
}

type ViewersPageData struct {
	Title       string
	Breadcrumbs []utils.Crumb
	UserEmail   string
	Group       db.Group
	Admin       db.User
	Viewers     []db.User
	Invites     []db.MagicLink
	IsAdmin     bool
}

type GroupPageData struct {
	Title       string
	Breadcrumbs []utils.Crumb
	UserEmail   string
	Group       db.Group
	Admin       db.User
	Income      int64
	Payouts     int64
	Expenses    int64
	Leftover    int64
	IsAdmin     bool
}
