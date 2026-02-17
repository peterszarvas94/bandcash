package settings

import "github.com/labstack/echo/v4"

func Register(e *echo.Echo) *Settings {
	settings := New()

	e.GET("/settings", settings.Index)
	e.POST("/settings/language", settings.UpdateLanguage)

	return settings
}
