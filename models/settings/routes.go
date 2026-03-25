package settings

import (
	"github.com/labstack/echo/v4"

	"bandcash/internal/middleware"
)

func RegisterRoutes(e *echo.Echo) *Settings {
	settings := New()

	e.GET("/language", settings.LanguagePage, middleware.RequireAuth, middleware.WithDetailState)
	e.GET("/settings", settings.LegacySettingsRedirect, middleware.RequireAuth, middleware.WithDetailState)
	e.GET("/account", settings.Index, middleware.RequireAuth, middleware.WithDetailState)
	e.POST("/account/language", settings.UpdateLanguage, middleware.RequireAuth, middleware.WithDetailState)
	e.POST("/account/details-state", settings.UpdateDetailsState, middleware.RequireAuth, middleware.WithDetailState)
	e.DELETE("/account/sessions/:id", settings.LogoutSession, middleware.RequireAuth, middleware.WithDetailState)
	e.DELETE("/account/sessions", settings.LogoutAllOtherSessions, middleware.RequireAuth, middleware.WithDetailState)

	return settings
}
