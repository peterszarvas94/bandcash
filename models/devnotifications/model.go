package devnotifications

import (
	"github.com/labstack/echo/v4"

	"bandcash/internal/utils"
)

type DevNotifications struct{}

func Register(e *echo.Echo) {
	if utils.Env().AppEnv != "development" {
		return
	}

	h := &DevNotifications{}
	e.GET("/dev", h.Redirect)
	e.GET("/dev/notifications", h.Index)
	e.GET("/dev/rate-limit", h.RateLimitPage)
	e.POST("/dev/notifications/inline", h.TestInline)
	e.POST("/dev/notifications/success", h.TestSuccess)
	e.POST("/dev/notifications/error", h.TestError)
	e.POST("/dev/notifications/info", h.TestInfo)
	e.POST("/dev/notifications/warning", h.TestWarning)
}
