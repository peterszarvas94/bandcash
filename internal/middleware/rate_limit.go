package middleware

import (
	"net/http"
	"time"

	echoMiddleware "github.com/labstack/echo/v4/middleware"
	"golang.org/x/time/rate"

	"github.com/labstack/echo/v4"

	"bandcash/internal/utils"
)

func GlobalRateLimit() echo.MiddlewareFunc {
	if utils.Env().DisableRateLimit {
		return func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(c echo.Context) error {
				return next(c)
			}
		}
	}

	store := echoMiddleware.NewRateLimiterMemoryStoreWithConfig(echoMiddleware.RateLimiterMemoryStoreConfig{
		Rate:      rate.Limit(4),
		Burst:     80,
		ExpiresIn: 5 * time.Minute,
	})

	return echoMiddleware.RateLimiterWithConfig(echoMiddleware.RateLimiterConfig{
		Store:               store,
		IdentifierExtractor: rateLimitIdentifier,
		DenyHandler:         rateLimitDenyHandler,
	})
}

func AuthRateLimit() echo.MiddlewareFunc {
	if utils.Env().DisableRateLimit {
		return func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(c echo.Context) error {
				return next(c)
			}
		}
	}

	store := echoMiddleware.NewRateLimiterMemoryStoreWithConfig(echoMiddleware.RateLimiterMemoryStoreConfig{
		Rate:      rate.Limit(5.0 / 60.0),
		Burst:     3,
		ExpiresIn: 10 * time.Minute,
	})

	return echoMiddleware.RateLimiterWithConfig(echoMiddleware.RateLimiterConfig{
		Store:               store,
		IdentifierExtractor: rateLimitIdentifier,
		DenyHandler:         rateLimitDenyHandler,
	})
}

func rateLimitIdentifier(c echo.Context) (string, error) {
	ip := c.RealIP()
	if ip == "" {
		ip = c.Request().RemoteAddr
	}
	return ip, nil
}

func rateLimitDenyHandler(c echo.Context, _ string, _ error) error {
	c.Response().Header().Set("Retry-After", "60")
	return c.NoContent(http.StatusTooManyRequests)
}
