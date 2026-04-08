package expense

import (
	"github.com/labstack/echo/v4"

	"bandcash/internal/middleware"
)

func RegisterRoutes(e *echo.Echo) {
	g := e.Group("/groups/:groupId", middleware.RequireAuth, middleware.WithDetailState, middleware.RequireGroup)
	g.GET("/expenses", IndexPage)
	g.GET("/expenses/:id", ShowPage)

	admin := g.Group("", middleware.RequireAdmin)
	admin.GET("/expenses/new", NewExpensePage)
	admin.GET("/expenses/:id/edit", EditExpensePage)
	admin.POST("/expenses", Create)
	admin.PUT("/expenses/:id", Update)
	admin.DELETE("/expenses/:id", Destroy)
	admin.PUT("/expenses/:id/toggle-paid", TogglePaid)
	admin.GET("/expenses/:id/paid_at", OpenPaidAtPrompt)
	admin.POST("/expenses/:id/paid_at", UpdatePaidAt)

}
