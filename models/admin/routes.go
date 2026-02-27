package admin

import (
	"github.com/labstack/echo/v4"

	"bandcash/internal/middleware"
)

func RegisterRoutes(e *echo.Echo) *Admin {
	a := New()

	g := e.Group("/admin", middleware.RequireAuth(), middleware.RequireSuperadmin())
	g.GET("", a.Dashboard)
	g.POST("/flags/signup", a.UpdateSignupFlag)

	return a
}
