package event

import (
	"github.com/labstack/echo/v4"

	"bandcash/internal/middleware"
)

func RegisterRoutes(e *echo.Echo) *Events {
	events := New()

	// Group routes under /groups/:groupId with auth middleware
	g := e.Group("/groups/:groupId", middleware.RequireAuth(), middleware.RequireGroup())

	g.GET("/events", events.Index)
	g.GET("/events/:id", events.Show)

	// Admin only routes
	admin := g.Group("", middleware.RequireAdmin())
	admin.POST("/events", events.Create)
	admin.POST("/events/:id/participants", events.CreateParticipant)
	admin.POST("/events/:id/participants/draft", events.OpenParticipantsDraft)
	admin.POST("/events/:id/participants/draft/:memberId", events.IncludeParticipantsDraftMember)
	admin.PUT("/events/:id", events.Update)
	admin.PUT("/events/:id/participants", events.SaveParticipantsBulk)
	admin.PUT("/events/:id/participants/:memberId", events.UpdateParticipant)
	admin.DELETE("/events/:id/participants/draft", events.CancelParticipantsDraft)
	admin.DELETE("/events/:id/participants/draft/:memberId", events.ExcludeParticipantsDraftMember)
	admin.DELETE("/events/:id", events.Destroy)
	admin.DELETE("/events/:id/participants/:memberId", events.DeleteParticipantTable)
	admin.PUT("/events/:id/toggle-paid", events.TogglePaid)
	admin.PUT("/events/:id/participants/:memberId/toggle-paid", events.ToggleParticipantPaid)

	return events
}
