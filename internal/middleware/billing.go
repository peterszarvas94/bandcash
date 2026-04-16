package middleware

import (
	"github.com/labstack/echo/v4"
)

func RequireCanCreateGroup(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Temporarily disabled until Lemon Squeezy store approval.
		return next(c)
	}
}

func RequireCanCreateEvent(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		return next(c)
	}
}
