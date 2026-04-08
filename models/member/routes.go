package member

import (
	"github.com/labstack/echo/v4"

	"bandcash/internal/middleware"
)

func RegisterRoutes(e *echo.Echo) {
	// Group routes under /groups/:groupId with auth middleware
	g := e.Group("/groups/:groupId", middleware.RequireAuth, middleware.WithDetailState, middleware.RequireGroup)

	g.GET("/members", Index)
	g.GET("/members/:id", Show)

	// Admin only routes
	admin := g.Group("", middleware.RequireAdmin)
	admin.GET("/members/new", NewMemberPage)
	admin.GET("/members/:id/edit", EditMemberPage)
	admin.POST("/members", Create)
	admin.PUT("/members/:id", Update)
	admin.GET("/members/:id/events/:eventId/paid_at", OpenParticipantPaidAtDialog)
	admin.POST("/members/:id/events/:eventId/paid_at", UpdateParticipantPaidAt)
	admin.PUT("/members/:id/events/:eventId/toggle-paid", ToggleParticipantPaid)
	admin.DELETE("/members/:id", Destroy)
}
