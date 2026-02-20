package event

import (
	"github.com/labstack/echo/v4"

	"bandcash/internal/middleware"
)

func Register(e *echo.Echo) *Events {
	events := New()

	// Group routes under /groups/:groupId with auth middleware
	g := e.Group("/groups/:groupId", middleware.RequireAuth(), middleware.RequireGroup())

	g.GET("", events.RedirectIndex)
	g.GET("/events", events.Index)
	g.GET("/events/:id", events.Show)

	// Admin only routes
	admin := g.Group("", middleware.RequireAdmin())
	admin.PUT("", events.UpdateGroup)
	admin.POST("/events", events.Create)
	admin.POST("/events/:id/participants", events.CreateParticipant)
	admin.PUT("/events/:id", events.Update)
	admin.PUT("/events/:id/participants/:memberId", events.UpdateParticipant)
	admin.DELETE("/events/:id", events.Destroy)
	admin.DELETE("/events/:id/participants/:memberId", events.DeleteParticipantTable)

	return events
}
