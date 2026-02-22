package member

import (
	"github.com/labstack/echo/v4"

	"bandcash/internal/middleware"
)

func RegisterRoutes(e *echo.Echo) *Members {
	members := New()

	// Group routes under /groups/:groupId with auth middleware
	g := e.Group("/groups/:groupId", middleware.RequireAuth(), middleware.RequireGroup())

	g.GET("/members", members.Index)
	g.GET("/members/:id", members.Show)

	// Admin only routes
	admin := g.Group("", middleware.RequireAdmin())
	admin.POST("/members", members.Create)
	admin.PUT("/members/:id", members.Update)
	admin.DELETE("/members/:id", members.Destroy)

	return members
}
