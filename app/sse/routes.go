package sse

import (
	"html/template"

	"github.com/labstack/echo/v4"

	"bandcash/app/entry"
	"bandcash/app/payee"
	"bandcash/internal/sse"
)

func Register(e *echo.Echo, entryTmpl, payeeTmpl *template.Template) {
	registry := sse.NewRegistry()
	registry.Add(entry.NewSSERenderer(entryTmpl).Render)
	registry.Add(payee.NewSSERenderer(payeeTmpl).Render)

	e.GET("/sse", sse.HandlerWithView(registry.Render))
}
