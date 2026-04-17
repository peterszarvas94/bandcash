package middleware

import (
	"time"

	"github.com/labstack/echo/v4"
)

func GlobalDelay(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		time.Sleep(300 * time.Millisecond)
		return next(c)
	}
}
