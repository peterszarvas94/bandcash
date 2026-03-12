package settings

import (
	"github.com/labstack/echo/v4"

	"bandcash/internal/middleware"
)

func RegisterRoutes(e *echo.Echo) *Settings {
	settings := New()

	e.GET("/language", settings.LanguagePage, middleware.RequireAuth())
	e.GET("/settings", settings.Index, middleware.RequireAuth())
	e.POST("/settings/language", settings.UpdateLanguage, middleware.RequireAuth())

	return settings
}
