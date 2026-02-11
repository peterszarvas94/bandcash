package counter

import (
	"html/template"

	"github.com/labstack/echo/v4"

	"webapp/internal/sse"
)

// Register registers counter routes.
func Register(e *echo.Echo) {
	ctr := &Counter{}

	// Parse shared head + feature template (contains "index" and "app" blocks)
	ctr.tmpl = template.Must(template.ParseFiles(
		"web/templates/head.html",
		"app/counter/templates/index.html",
	))

	e.GET("/counter", ctr.Index)
	e.PATCH("/counter", ctr.Update)
	e.GET("/counter/sse", sse.Handler(ctr.tmpl, "app", ctr.Data))
}
