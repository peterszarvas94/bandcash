package event

import "github.com/labstack/echo/v4"

func Register(e *echo.Echo) *Events {
	events := New()

	e.GET("/event", events.Index)
	e.GET("/event/:id", events.Show)

	e.POST("/event", events.Create)
	e.POST("/event/:id/participant", events.CreateParticipant)

	e.PUT("/event/:id", events.Update)
	e.PUT("/event/:id/participant/:memberId", events.UpdateParticipant)

	e.DELETE("/event/:id", events.Destroy)
	e.DELETE("/event/:id/participant/:memberId", events.DeleteParticipantTable)

	return events
}
