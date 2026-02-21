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

type GroupsPageData struct {
	Title        string
	Breadcrumbs  []utils.Crumb
	UserEmail    string
	AdminGroups  []GroupSummary
	ReaderGroups []GroupSummary
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
	TotalAmount int64
	IsAdmin     bool
}
