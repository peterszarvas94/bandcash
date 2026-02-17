package entry

import "github.com/labstack/echo/v4"

func Register(e *echo.Echo) *Entries {
	entries := New()

	e.GET("/entry", entries.Index)
	e.GET("/entry/:id", entries.Show)

	e.POST("/entry", entries.Create)
	e.POST("/entry/:id/participant", entries.CreateParticipant)

	e.PUT("/entry/:id", entries.Update)
	e.PUT("/entry/:id/participant/:payeeId", entries.UpdateParticipant)

	e.DELETE("/entry/:id", entries.Destroy)
	e.DELETE("/entry/:id/participant/:payeeId", entries.DeleteParticipantTable)

	return entries
}
