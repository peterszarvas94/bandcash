package payee

import "github.com/labstack/echo/v4"

func Register(e *echo.Echo) *Payees {
	payees := New()

	e.GET("/payee", payees.Index)
	e.POST("/payee", payees.Create)
	e.GET("/payee/:id", payees.Show)
	e.PUT("/payee/:id", payees.Update)
	e.DELETE("/payee/:id", payees.Destroy)

	return payees
}
