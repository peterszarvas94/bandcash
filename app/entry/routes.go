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
	e.POST("/entry/table", entries.CreateTable)
	e.GET("/entry/:id", entries.Show)
	e.GET("/entry/:id/edit", entries.Edit)
	e.POST("/entry/:id/participants", entries.AddParticipant)
	e.POST("/entry/:id/participants/table", entries.AddParticipantTable)
	e.PUT("/entry/:id/participants/:payeeId", entries.UpdateParticipant)
	e.PUT("/entry/:id/participants/:payeeId/table", entries.UpdateParticipantTable)
	e.DELETE("/entry/:id/participants/:payeeId", entries.DeleteParticipant)
	e.PUT("/entry/:id", entries.Update)
	e.PUT("/entry/:id/table", entries.UpdateTable)
	e.DELETE("/entry/:id", entries.Destroy)
}
