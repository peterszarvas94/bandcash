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
	entryTmpl := entry.Register(e)
	payeeTmpl := payee.Register(e)
	appSSE.Register(e, entryTmpl, payeeTmpl)
}
