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

	// Group users pages for any group member
	accessRoutes := e.Group("/groups/:groupId", middleware.RequireAuth, middleware.WithDetailState, middleware.RequireGroup)
	accessRoutes.GET("", grp.GroupPage)
	accessRoutes.GET("/users", grp.AccessPage)
	accessRoutes.GET("/users/:id", grp.AccessEntryPage)
	accessRoutes.POST("/leave", grp.LeaveGroup)

	// User management (admin only)
	adminAccessRoutes := e.Group("/groups/:groupId", middleware.RequireAuth, middleware.WithDetailState, middleware.RequireGroup, middleware.RequireAdmin)
	adminAccessRoutes.GET("/edit", grp.EditGroupPage)
	adminAccessRoutes.GET("/users/new", grp.AccessNewPage)
	adminAccessRoutes.GET("/users/:id/edit", grp.AccessEditPage)
	adminAccessRoutes.PUT("", grp.UpdateGroup)
	adminAccessRoutes.DELETE("", grp.DeleteGroup)
	adminAccessRoutes.POST("/users", grp.AddViewer)
	adminAccessRoutes.DELETE("/users/:id", grp.DeleteAccessEntry)
	adminAccessRoutes.PUT("/users/:id/admin", grp.PromoteViewerToAdmin)
	adminAccessRoutes.PUT("/users/:id/viewer", grp.DemoteAdminToViewer)

	return grp
}
