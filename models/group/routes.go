package group

import (
	"github.com/labstack/echo/v4"

	"bandcash/internal/middleware"
)

func Register(e *echo.Echo) *Group {
	grp := New()

	// Group creation (requires auth but no existing group)
	e.GET("/groups/new", grp.NewGroupPage, middleware.RequireAuth())
	e.POST("/groups", grp.CreateGroup, middleware.RequireAuth())

	return grp
}
