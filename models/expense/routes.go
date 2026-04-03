package expense

import (
	"github.com/labstack/echo/v4"

	"bandcash/internal/middleware"
)

func RegisterRoutes(e *echo.Echo) *Expenses {
	expenses := New()

	g := e.Group("/groups/:groupId", middleware.RequireAuth, middleware.WithDetailState, middleware.RequireGroup)
	g.GET("/expenses", expenses.IndexPage)
	g.GET("/expenses/:id", expenses.ShowPage)

	admin := g.Group("", middleware.RequireAdmin)
	admin.GET("/expenses/new", expenses.NewExpensePage)
	admin.GET("/expenses/:id/edit", expenses.EditExpensePage)
	admin.POST("/expenses", expenses.Create)
	admin.PUT("/expenses/:id", expenses.Update)
	admin.DELETE("/expenses/:id", expenses.Destroy)
	admin.PUT("/expenses/:id/toggle-paid", expenses.TogglePaid)

	return expenses
}
