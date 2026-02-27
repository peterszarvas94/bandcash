package middleware

import (
	"time"

	"github.com/labstack/echo/v4"

	"bandcash/internal/utils"
)

func GlobalDelay() echo.MiddlewareFunc {
	delayMS := utils.Env().DevGlobalDelayMS
	if delayMS <= 0 {
		return func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(c echo.Context) error {
				return next(c)
			}
		}
	}

	delay := time.Duration(delayMS) * time.Millisecond

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			time.Sleep(delay)
			return next(c)
		}
	}
}
