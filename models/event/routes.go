package event

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"bandcash/internal/middleware"
)

func RegisterRoutes(e *echo.Echo) *Events {
	events := New()

	// Group routes under /groups/:groupId with auth middleware
	g := e.Group("/groups/:groupId", middleware.RequireAuth, middleware.WithDetailState, middleware.RequireGroup)

	g.GET("/events", events.IndexPage)
	g.GET("/overview", func(c echo.Context) error {
		groupID := c.Param("groupId")
		return c.Redirect(http.StatusMovedPermanently, "/groups/"+groupID+"/events")
	})
	g.GET("/events/:id", events.ShowPage)
	g.GET("/events/:id/members/:memberId/note", events.OpenParticipantNoteDialog)

	// Admin only routes
	admin := g.Group("", middleware.RequireAdmin)
	admin.GET("/events/new", events.NewEventPage)
	admin.GET("/events/:id/edit", events.EditEventPage)
	admin.POST("/events", events.Create)
	admin.POST("/events/:id", events.Update)
	admin.POST("/events/:id/paid", events.TogglePaid)
	admin.GET("/events/:id/paid_at", events.OpenPaidAtPrompt)
	admin.POST("/events/:id/paid_at", events.UpdatePaidAt)
	admin.GET("/events/:id/members/:memberId/paid_at", events.OpenParticipantPaidAtDialog)
	admin.POST("/events/:id/members/:memberId/paid_at", events.UpdateParticipantPaidAt)
	admin.POST("/events/:id/members/:memberId/note", events.UpdateParticipantNote)
	admin.POST("/events/:id/participants/draft", events.OpenParticipantsDraft)
	admin.POST("/events/:id/participants/draft/rows", events.UpdateParticipantsDraftRows)
	admin.PUT("/events/:id/participants", events.SaveParticipantsBulk)
	admin.DELETE("/events/:id/participants/draft", events.CancelParticipantsDraft)
	admin.DELETE("/events/:id", events.Destroy)
	admin.POST("/events/:id/members/:memberId/paid", events.ToggleParticipantPaid)

	return events
}
