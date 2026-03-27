package account

import (
	"bandcash/internal/db"
	"bandcash/internal/utils"
)

type AccountData struct {
	Title       string
	Breadcrumbs []utils.Crumb
	CurrentLang string
	UserEmail   string
}

type SessionsData struct {
	Title            string
	Breadcrumbs      []utils.Crumb
	UserEmail        string
	CurrentSessionID string
	Sessions         []db.UserSession
}
