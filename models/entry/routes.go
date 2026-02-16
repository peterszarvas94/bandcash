package entry

import (
	"html/template"

	"github.com/labstack/echo/v4"
)

func Register(e *echo.Echo) *Entries {
	entries := New()

	entries.tmpl = template.Must(template.New("").Funcs(template.FuncMap{
		"add": func(a, b int64) int64 { return a + b },
	}).ParseFiles(
		"models/shared/templates/head.html",
		"models/shared/templates/breadcrumbs.html",
		"models/entry/templates/index.html",
		"models/entry/templates/new.html",
		"models/entry/templates/show.html",
		"models/entry/templates/edit.html",
	))

	e.GET("/entry", entries.Index)
	e.GET("/entry/:id", entries.Show)
	e.GET("/entry/:id/edit", entries.Edit)

	e.POST("/entry", entries.Create)
	e.POST("/entry/:id/participant", entries.CreateParticipant)

	e.PUT("/entry/:id", entries.Update)
	e.PUT("/entry/:id/participant/:payeeId", entries.UpdateParticipant)

	e.DELETE("/entry/:id", entries.Destroy)
	e.DELETE("/entry/:id/participant/:payeeId", entries.DeleteParticipantTable)

	return entries
}
