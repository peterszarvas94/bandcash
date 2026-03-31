package account

import (
	"bandcash/internal/db"
	"bandcash/internal/utils"
)

type AccountData struct {
	Title           string
	Breadcrumbs     []utils.Crumb
	CurrentLang     string
	UserEmail       string
	Signals         map[string]any
	IsAuthenticated bool
	IsSuperAdmin    bool
}

type SessionsData struct {
	Title            string
	Breadcrumbs      []utils.Crumb
	CurrentSessionID string
	Sessions         []db.UserSession
	Signals          map[string]any
	IsAuthenticated  bool
	IsSuperAdmin     bool
}
