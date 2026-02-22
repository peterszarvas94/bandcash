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
	e.GET("/dev", h.Redirect)
	e.GET("/dev/notifications", h.Index)
	e.GET("/dev/rate-limit", h.RateLimitPage)
	e.GET("/dev/body-limit", h.BodyLimitPage)
	e.POST("/dev/body-limit/global", h.BodyLimitGlobalTest)
	e.POST("/dev/body-limit/auth", h.BodyLimitAuthTest, middleware.AuthBodyLimit())
	e.POST("/dev/notifications/inline", h.TestInline)
	e.POST("/dev/notifications/success", h.TestSuccess)
	e.POST("/dev/notifications/error", h.TestError)
	e.POST("/dev/notifications/info", h.TestInfo)
	e.POST("/dev/notifications/warning", h.TestWarning)
}
