package entry

import (
	"html/template"

	"github.com/labstack/echo/v4"
)

func Register(e *echo.Echo) {
	entries := New()

	entries.tmpl = template.Must(template.ParseFiles(
		"web/templates/head.html",
		"web/templates/breadcrumbs.html",
		"app/entry/templates/index.html",
		"app/entry/templates/new.html",
		"app/entry/templates/show.html",
		"app/entry/templates/edit.html",
	))

	e.GET("/entry", entries.Index)
	e.GET("/entry/new", entries.New)
	e.POST("/entry", entries.Create)
	e.GET("/entry/:id", entries.Show)
	e.GET("/entry/:id/edit", entries.Edit)
	e.POST("/entry/:id/participants", entries.AddParticipant)
	e.PUT("/entry/:id/participants/:payeeId", entries.UpdateParticipant)
	e.PUT("/entry/:id", entries.Update)
	e.DELETE("/entry/:id", entries.Destroy)
}
