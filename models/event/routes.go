package event

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"bandcash/internal/middleware"
)

func RegisterRoutes(e *echo.Echo) {
	// Group routes under /groups/:groupId with auth middleware
	g := e.Group("/groups/:groupId", middleware.RequireAuth, middleware.WithDetailState, middleware.RequireGroup)

	g.GET("/events", IndexPage)
	g.GET("/overview", func(c echo.Context) error {
		groupID := c.Param("groupId")
		return c.Redirect(http.StatusMovedPermanently, "/groups/"+groupID+"/events")
	})
	g.GET("/events/:id", ShowPage)
	g.GET("/events/:id/members/:memberId/note", OpenParticipantNoteDialog)

	// Admin only routes
	admin := g.Group("", middleware.RequireAdmin)
	admin.GET("/events/new", NewEventPage)
	admin.GET("/events/:id/edit", EditEventPage)
	admin.POST("/events", Create)
	admin.POST("/events/:id", Update)
	admin.POST("/events/:id/paid", TogglePaid)
	admin.GET("/events/:id/paid_at", OpenPaidAtPrompt)
	admin.POST("/events/:id/paid_at", UpdatePaidAt)
	admin.GET("/events/:id/members/:memberId/paid_at", OpenParticipantPaidAtDialog)
	admin.POST("/events/:id/members/:memberId/paid_at", UpdateParticipantPaidAt)
	admin.POST("/events/:id/members/:memberId/note", UpdateParticipantNote)
	admin.POST("/events/:id/participants/draft", OpenParticipantsDraft)
	admin.POST("/events/:id/participants/draft/rows", UpdateParticipantsDraftRows)
	admin.PUT("/events/:id/participants", SaveParticipantsBulk)
	admin.DELETE("/events/:id/participants/draft", CancelParticipantsDraft)
	admin.DELETE("/events/:id", Destroy)
	admin.POST("/events/:id/members/:memberId/paid", ToggleParticipantPaid)

}
