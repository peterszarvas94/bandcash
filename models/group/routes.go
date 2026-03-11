package group

import (
	"github.com/labstack/echo/v4"

	"bandcash/internal/middleware"
)

func RegisterRoutes(e *echo.Echo) *Group {
	grp := New()

	// Group creation (requires auth)
	e.GET("/dashboard", grp.GroupsPage, middleware.RequireAuth())
	e.GET("/groups/new", grp.NewGroupPage, middleware.RequireAuth())
	e.POST("/groups", grp.CreateGroup, middleware.RequireAuth())

	// Group access pages for any group member
	accessRoutes := e.Group("/groups/:groupId", middleware.RequireAuth(), middleware.RequireGroup())
	accessRoutes.GET("", grp.GroupPage)
	accessRoutes.GET("/access", grp.ViewersPage)
	accessRoutes.GET("/access/viewers", grp.ViewersPage)
	accessRoutes.GET("/access/pending", grp.ViewersPendingPage)
	accessRoutes.GET("/access/admins", grp.ViewersAdminsPage)
	accessRoutes.POST("/leave", grp.LeaveGroup)

	// Access management (admin only)
	adminAccessRoutes := e.Group("/groups/:groupId", middleware.RequireAuth(), middleware.RequireGroup(), middleware.RequireAdmin())
	adminAccessRoutes.PUT("", grp.UpdateGroup)
	adminAccessRoutes.DELETE("", grp.DeleteGroup)
	adminAccessRoutes.POST("/access/viewers", grp.AddViewer)
	adminAccessRoutes.POST("/access/pending", grp.AddViewer)
	adminAccessRoutes.DELETE("/access/viewers/:userId", grp.RemoveViewer)
	adminAccessRoutes.DELETE("/invites/:inviteId", grp.CancelInvite)

	return grp
}
