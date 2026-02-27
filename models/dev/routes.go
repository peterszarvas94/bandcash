package dev

import (
	"github.com/labstack/echo/v4"

	"bandcash/internal/middleware"
	"bandcash/internal/utils"
)

func RegisterRoutes(e *echo.Echo) {
	if utils.Env().AppEnv != "development" {
		return
	}

	h := &DevNotifications{}
	e.GET("/dev", h.DevPageHandler)
	e.POST("/dev/body-limit/global", h.TestBodyLimitGlobal)
	e.POST("/dev/body-limit/auth", h.TestBodyLimitAuth, middleware.AuthBodyLimit())
	e.POST("/dev/spinner-test", h.TestSpinner)
	e.POST("/dev/notifications/inline", h.TestInline)
	e.POST("/dev/notifications/success", h.TestSuccess)
	e.POST("/dev/notifications/error", h.TestError)
	e.POST("/dev/notifications/info", h.TestInfo)
	e.POST("/dev/notifications/warning", h.TestWarning)
}
