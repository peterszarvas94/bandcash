package entry

import (
	"html/template"

	"github.com/labstack/echo/v4"

	"webapp/internal/sse"
)

func Register(e *echo.Echo) {
	entries := New()

	entries.tmpl = template.Must(template.ParseFiles(
		"web/templates/head.html",
		"app/entry/templates/list.html",
		"app/entry/templates/new.html",
		"app/entry/templates/show.html",
		"app/entry/templates/edit.html",
	))

	e.GET("/entry", entries.List)
	e.GET("/entry/new", entries.New)
	e.POST("/entry", entries.Create)
	e.GET("/entry/:id", entries.Show)
	e.GET("/entry/:id/edit", entries.Edit)
	e.PUT("/entry/:id", entries.Update)
	e.DELETE("/entry/:id", entries.Delete)
	e.GET("/entry/sse", sse.Handler(entries.tmpl, "app", entries.DataForSSE))
}
