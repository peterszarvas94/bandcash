package home

import "bandcash/internal/utils"

type HomeData struct {
	Title           string
	Breadcrumbs     []utils.Crumb
	CurrentLang     string
	IsAuthenticated bool
	IsSuperAdmin    bool
	UserID          string
	UserEmail       string
}
