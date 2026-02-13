package payee

import (
	"html/template"

	"github.com/labstack/echo/v4"
)

func Register(e *echo.Echo) {
	payees := New()

	payees.tmpl = template.Must(template.ParseFiles(
		"web/templates/head.html",
		"web/templates/breadcrumbs.html",
		"app/payee/templates/index.html",
		"app/payee/templates/new.html",
		"app/payee/templates/show.html",
		"app/payee/templates/edit.html",
	))

	e.GET("/payee", payees.Index)
	e.POST("/payee", payees.Create)
	e.GET("/payee/:id", payees.Show)
	e.GET("/payee/:id/edit", payees.Edit)
	e.PUT("/payee/:id", payees.Update)
	e.PUT("/payee/:id/single", payees.UpdateSingle)
	e.DELETE("/payee/:id", payees.Destroy)
}
