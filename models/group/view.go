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
	AdminGroups  []db.Group
	ReaderGroups []db.Group
	MessageKey   string
	ErrorKey     string
}

type ViewersPageData struct {
	Title       string
	Breadcrumbs []utils.Crumb
	UserEmail   string
	Group       db.Group
	Viewers     []db.User
	MessageKey  string
	ErrorKey    string
}
