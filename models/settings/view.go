package settings

import "bandcash/internal/utils"

type SettingsData struct {
	Title       string
	Breadcrumbs []utils.Crumb
	CurrentLang string
	UserEmail   string
}
