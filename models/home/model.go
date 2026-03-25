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
	return h.DataWithTitle(ctx, ctxi18n.T(ctx, "app.title"))
}

func (h *Home) DataWithTitle(ctx context.Context, title string) HomeData {
	return HomeData{
		Title:       title,
		CurrentLang: appi18n.LocaleCode(ctx),
		Breadcrumbs: []utils.Crumb{},
	}
}

func (h *Home) LegalDataWithTitle(ctx context.Context, title string, currentLabel string) HomeData {
	data := h.DataWithTitle(ctx, title)
	data.Breadcrumbs = []utils.Crumb{{Label: currentLabel, Href: ""}}
	return data
}

// RegisterRoutes registers home routes.
func RegisterRoutes(e *echo.Echo) {
	h := &Home{}
	e.GET("/", h.Index)
	e.GET("/pricing", h.Pricing)
	e.GET("/terms-and-conditions", h.TermsAndConditions)
	e.GET("/privacy-policy", h.PrivacyPolicy)
	e.GET("/refund-policy", h.RefundPolicy)
}
