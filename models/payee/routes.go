package payee

import (
	"html/template"

	"github.com/labstack/echo/v4"
)

func Register(e *echo.Echo) *Payees {
	payees := New()

	payees.tmpl = template.Must(template.ParseFiles(
		"models/shared/templates/head.html",
		"models/shared/templates/breadcrumbs.html",
		"models/payee/templates/index.html",
		"models/payee/templates/new.html",
		"models/payee/templates/show.html",
		"models/payee/templates/edit.html",
	))

	e.GET("/payee", payees.Index)
	e.POST("/payee", payees.Create)
	e.GET("/payee/:id", payees.Show)
	e.GET("/payee/:id/edit", payees.Edit)
	e.PUT("/payee/:id", payees.Update)
	e.PUT("/payee/:id/single", payees.UpdateSingle)
	e.DELETE("/payee/:id", payees.Destroy)

	return payees
}
