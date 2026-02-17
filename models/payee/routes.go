package payee

import (
	"html/template"

	"github.com/labstack/echo/v4"
)

func Register(e *echo.Echo) *Payees {
	payees := New()

	payees.tmpl = template.Must(template.New("").Funcs(template.FuncMap{
		"add": func(a, b int64) int64 { return a + b },
	}).ParseFiles(
		"models/shared/templates/head.html",
		"models/shared/templates/breadcrumbs.html",
		"models/payee/templates/index.html",
		"models/payee/templates/show.html",
	))

	e.GET("/payee", payees.Index)
	e.POST("/payee", payees.Create)
	e.GET("/payee/:id", payees.Show)
	e.PUT("/payee/:id", payees.Update)
	e.DELETE("/payee/:id", payees.Destroy)

	return payees
}
