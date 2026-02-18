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

	// Viewer management (admin only)
	g := e.Group("/groups/:groupId", middleware.RequireAuth(), middleware.RequireGroup(), middleware.RequireAdmin())
	g.GET("/viewers", grp.ViewersPage)
	g.POST("/viewers", grp.AddViewer)
	g.POST("/viewers/:userId/remove", grp.RemoveViewer)

	return grp
}
