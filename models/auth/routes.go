package auth

import (
	"github.com/labstack/echo/v4"

	"bandcash/internal/middleware"
)

func RegisterRoutes(e *echo.Echo) *Auth {
	auth := New()

	// Public auth routes
	e.GET("/login", auth.LoginPage)
	e.POST("/login", auth.LoginRequest, middleware.AuthBodyLimit, middleware.AuthRateLimit)
	e.GET("/login/verify", auth.VerifyMagicLink)
	e.DELETE("/session", auth.Logout)

	return auth
}
