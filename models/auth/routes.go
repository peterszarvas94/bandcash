package auth

import (
	"github.com/labstack/echo/v4"

	"bandcash/internal/middleware"
)

func RegisterRoutes(e *echo.Echo) *Auth {
	auth := New()

	// Public auth routes
	e.GET("/auth/login", auth.LoginPage)
	e.POST("/auth/login", auth.LoginRequest, middleware.AuthBodyLimit, middleware.AuthRateLimit)
	e.GET("/auth/verify", auth.VerifyMagicLink)
	e.DELETE("/auth/session", auth.Logout)

	return auth
}
