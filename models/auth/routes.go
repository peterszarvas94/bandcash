package auth

import "github.com/labstack/echo/v4"

func Register(e *echo.Echo) *Auth {
	auth := New()

	// Public auth routes
	e.GET("/auth/login", auth.LoginPage)
	e.POST("/auth/login", auth.LoginRequest)
	e.GET("/auth/signup", auth.SignupPage)
	e.POST("/auth/signup", auth.SignupRequest)
	e.GET("/auth/login-sent", auth.LoginSentPage)
	e.GET("/auth/verify", auth.VerifyMagicLink)
	e.GET("/auth/logout", auth.Logout)

	return auth
}
