package settings

import (
	"context"

	ctxi18n "github.com/invopop/ctxi18n/i18n"

	appi18n "bandcash/internal/i18n"
	"bandcash/internal/utils"
)

type Settings struct{}

func New() *Settings {
	return &Settings{}
}

func (s *Settings) Data(ctx context.Context) SettingsData {
	return SettingsData{
		Title:       ctxi18n.T(ctx, "settings.title"),
		CurrentLang: appi18n.LocaleCode(ctx),
		Breadcrumbs: []utils.Crumb{
			{Label: ctxi18n.T(ctx, "settings.title")},
		},
	}
}
