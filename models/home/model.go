package home

import (
	"context"

	ctxi18n "github.com/invopop/ctxi18n/i18n"

	appi18n "bandcash/internal/i18n"
	"bandcash/internal/utils"
)

// Data returns data for rendering.
func Data(ctx context.Context) HomeData {
	return DataWithTitle(ctx, ctxi18n.T(ctx, "app.title"))
}

func DataWithTitle(ctx context.Context, title string) HomeData {
	return HomeData{
		Title:       title,
		CurrentLang: appi18n.LocaleCode(ctx),
		Breadcrumbs: []utils.Crumb{},
	}
}

func LegalDataWithTitle(ctx context.Context, title string, currentLabel string) HomeData {
	data := DataWithTitle(ctx, title)
	data.Breadcrumbs = []utils.Crumb{{Label: currentLabel, Href: ""}}
	return data
}
