package server

import (
	"github.com/labstack/echo/v4"

	"bandcash/app/entry"
	"bandcash/app/health"
	"bandcash/app/home"
	"bandcash/app/payee"
	appSSE "bandcash/app/sse"
)

func RegisterRoutes(e *echo.Echo) {
	e.Static("/static", "web/static")

	health.Register(e)
	home.Register(e)
	entry.Register(e)
	payee.Register(e)
	appSSE.Register(e)
}
