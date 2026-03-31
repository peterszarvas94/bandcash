package admin

import (
	"github.com/labstack/echo/v4"

	"bandcash/internal/middleware"
)

func RegisterRoutes(e *echo.Echo) *Admin {
	a := New()

	g := e.Group("/admin", middleware.RequireAuth, middleware.WithDetailState, middleware.RequireSuperadmin)
	g.GET("", a.Dashboard)
	g.GET("/flags", a.FlagsPage)
	g.GET("/users", a.UsersPage)
	g.GET("/groups", a.GroupsPage)
	g.GET("/sessions", a.SessionsPage)
	g.POST("/flags/signup", a.UpdateSignupFlag)
	g.POST("/users/:userId/ban", a.BanUser)
	g.POST("/users/:userId/unban", a.UnbanUser)
	g.DELETE("/users/:id/sessions/:sessionid", a.LogoutSession)
	g.DELETE("/users/:id/sessions/", a.LogoutAllUserSessions)

	return a
}
