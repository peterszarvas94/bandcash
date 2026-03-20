package settings

import (
	"github.com/labstack/echo/v4"

	"bandcash/internal/middleware"
)

func RegisterRoutes(e *echo.Echo) *Settings {
	settings := New()

	e.GET("/language", settings.LanguagePage, middleware.RequireAuth, middleware.WithDetailState)
	e.GET("/settings", settings.Index, middleware.RequireAuth, middleware.WithDetailState)
	e.POST("/settings/language", settings.UpdateLanguage, middleware.RequireAuth, middleware.WithDetailState)
	e.POST("/settings/details-state", settings.UpdateDetailsState, middleware.RequireAuth, middleware.WithDetailState)

	return settings
}
