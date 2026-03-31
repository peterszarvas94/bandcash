package auth

import "bandcash/internal/utils"

type AuthPageData struct {
	Title           string
	Breadcrumbs     []utils.Crumb
	CurrentLang     string
	SignupEnabled   bool
	IsAuthenticated bool
	IsSuperAdmin    bool
	Signals         map[string]any
}
