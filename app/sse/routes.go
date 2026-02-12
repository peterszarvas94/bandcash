package sse

import (
	"github.com/labstack/echo/v4"

	"bandcash/app/entry"
	"bandcash/app/payee"
	"bandcash/internal/sse"
)

func Register(e *echo.Echo) {
	registry := sse.NewRegistry()
	registry.Add(entry.NewSSERenderer().Render)
	registry.Add(payee.NewSSERenderer().Render)

	e.GET("/sse", sse.HandlerWithView(registry.Render))
}
