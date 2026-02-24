package home

import (
	"context"

	ctxi18n "github.com/invopop/ctxi18n/i18n"
	"github.com/labstack/echo/v4"

	appi18n "bandcash/internal/i18n"
	"bandcash/internal/utils"
)

type Home struct {
}

// Data returns data for rendering.
func (h *Home) Data(ctx context.Context) HomeData {
	return HomeData{
		Title:       ctxi18n.T(ctx, "app.title"),
		CurrentLang: appi18n.LocaleCode(ctx),
		Breadcrumbs: []utils.Crumb{},
	}
}

// RegisterRoutes registers home routes.
func RegisterRoutes(e *echo.Echo) {
	h := &Home{}
	e.GET("/", h.Index)
}
