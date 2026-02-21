package group

import (
	"github.com/labstack/echo/v4"

	"bandcash/internal/middleware"
)

func Register(e *echo.Echo) *Group {
	grp := New()

	// Group creation (requires auth)
	e.GET("/dashboard", grp.GroupsPage, middleware.RequireAuth())
	e.GET("/groups/new", grp.NewGroupPage, middleware.RequireAuth())
	e.POST("/groups", grp.CreateGroup, middleware.RequireAuth())

	// Group access pages for any group member
	leave := e.Group("/groups/:groupId", middleware.RequireAuth(), middleware.RequireGroup())
	leave.GET("", grp.GroupPage)
	leave.GET("/viewers", grp.ViewersPage)
	leave.POST("/leave", grp.LeaveGroup)

	// Viewer management (admin only)
	g := e.Group("/groups/:groupId", middleware.RequireAuth(), middleware.RequireGroup(), middleware.RequireAdmin())
	g.PUT("", grp.UpdateGroup)
	g.POST("/viewers", grp.AddViewer)
	g.POST("/viewers/:userId/remove", grp.RemoveViewer)
	g.POST("/invites/:inviteId/remove", grp.CancelInvite)
	g.POST("/delete", grp.DeleteGroup)

	return grp
}
