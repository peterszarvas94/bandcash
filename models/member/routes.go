package member

import "github.com/labstack/echo/v4"

func Register(e *echo.Echo) *Members {
	members := New()

	e.GET("/member", members.Index)
	e.POST("/member", members.Create)
	e.GET("/member/:id", members.Show)
	e.PUT("/member/:id", members.Update)
	e.DELETE("/member/:id", members.Destroy)

	return members
}
