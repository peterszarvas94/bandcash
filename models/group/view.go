package group

import (
	"time"

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
	GroupsTable  utils.TableLayout
}

type GroupSummary struct {
	Group       db.Group
	ViewerCount int
	AdminEmail  string
}

type AccessPageData struct {
	Title         string
	Breadcrumbs   []utils.Crumb
	UserEmail     string
	CurrentUserID string
	Group         db.Group
	AccessRows    []GroupAccessRow
	IsAdmin       bool
	Query         utils.TableQuery
	Pager         utils.TablePagination
	GroupID       string
	AccessTable   utils.TableLayout
}

type GroupAccessRow struct {
	Kind      string
	Status    string
	Role      string
	Email     string
	UserID    string
	InviteID  string
	CreatedAt time.Time
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
