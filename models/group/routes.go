package group

import (
	"github.com/labstack/echo/v4"

	"bandcash/internal/middleware"
)

func RegisterRoutes(e *echo.Echo) *Group {
	grp := New()

	// Group creation (requires auth)
	e.GET("/dashboard", grp.GroupsPage, middleware.RequireAuth, middleware.WithDetailState)
	e.GET("/groups/new", grp.NewGroupPage, middleware.RequireAuth, middleware.WithDetailState)
	e.POST("/groups", grp.CreateGroup, middleware.RequireAuth, middleware.WithDetailState)

	// Group access pages for any group member
	accessRoutes := e.Group("/groups/:groupId", middleware.RequireAuth, middleware.WithDetailState, middleware.RequireGroup)
	accessRoutes.GET("", grp.GroupPage)
	accessRoutes.GET("/access", grp.AccessPage)
	accessRoutes.POST("/leave", grp.LeaveGroup)

	// Access management (admin only)
	adminAccessRoutes := e.Group("/groups/:groupId", middleware.RequireAuth, middleware.WithDetailState, middleware.RequireGroup, middleware.RequireAdmin)
	adminAccessRoutes.PUT("", grp.UpdateGroup)
	adminAccessRoutes.DELETE("", grp.DeleteGroup)
	adminAccessRoutes.POST("/access", grp.AddViewer)
	adminAccessRoutes.DELETE("/access/users/:userId", grp.RemoveViewer)
	adminAccessRoutes.PUT("/access/users/:userId/admin", grp.PromoteViewerToAdmin)
	adminAccessRoutes.PUT("/access/users/:userId/viewer", grp.DemoteAdminToViewer)
	adminAccessRoutes.DELETE("/access/invites/:inviteId", grp.CancelInvite)

	return grp
}
