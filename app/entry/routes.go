package entry

import (
	"html/template"

	"github.com/labstack/echo/v4"
)

func Register(e *echo.Echo) *Entries {
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
	e.GET("/entry/:id", entries.Show)
	e.GET("/entry/:id/edit", entries.Edit)

	e.POST("/entry", entries.Create)
	e.POST("/entry/:id/participant", entries.CreateParticipant)

	e.PUT("/entry/:id", entries.Update)
	e.PUT("/entry/:id/single", entries.UpdateSingle)
	e.PUT("/entry/:id/participant/:payeeId", entries.UpdateParticipant)

	e.DELETE("/entry/:id", entries.Destroy)
	e.DELETE("/entry/:id/single", entries.DestroySingle)
	e.DELETE("/entry/:id/participant/:payeeId", entries.DeleteParticipantTable)

	return entries
}
