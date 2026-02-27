package expense

import (
	"github.com/labstack/echo/v4"

	"bandcash/internal/middleware"
)

func RegisterRoutes(e *echo.Echo) *Expenses {
	expenses := New()

	g := e.Group("/groups/:groupId", middleware.RequireAuth(), middleware.RequireGroup())
	g.GET("/expenses", expenses.Index)

	admin := g.Group("", middleware.RequireAdmin())
	admin.POST("/expenses", expenses.Create)
	admin.PUT("/expenses/:id", expenses.Update)
	admin.DELETE("/expenses/:id", expenses.Destroy)

	return expenses
}
