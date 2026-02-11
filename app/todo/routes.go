package todo

import (
	"html/template"

	"github.com/labstack/echo/v4"

	"webapp/internal/sse"
)

// Register registers todo routes.
func Register(e *echo.Echo) {
	t := &Todo{
		items: []Item{},
	}

	// Parse shared head + feature template (contains "index" and "app" blocks)
	t.tmpl = template.Must(template.ParseFiles(
		"web/templates/head.html",
		"app/todo/templates/index.html",
	))

	e.GET("/todo", t.Index)
	e.POST("/todo", t.Create)
	e.GET("/todo/sse", sse.Handler(t.tmpl, "app", t.Data))
}
