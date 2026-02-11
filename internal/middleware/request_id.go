package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log/slog"

	"github.com/labstack/echo/v4"
)

type contextKey string

const RequestIDKey contextKey = "request_id"

// RequestID generates a unique request ID and adds it to context and logs
func RequestID() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			requestID := generateID()
			c.Set(string(RequestIDKey), requestID)
			c.Response().Header().Set("X-Request-ID", requestID)

			// Store in request context for handlers
			ctx := context.WithValue(c.Request().Context(), RequestIDKey, requestID)
			c.SetRequest(c.Request().WithContext(ctx))

			return next(c)
		}
	}
}

// GetRequestID retrieves the request ID from context
func GetRequestID(ctx context.Context) string {
	if id, ok := ctx.Value(RequestIDKey).(string); ok {
		return id
	}
	return ""
}

// Logger returns a slog logger with request_id and client_id if available
func Logger(c echo.Context) *slog.Logger {
	attrs := []any{}

	if id := GetRequestID(c.Request().Context()); id != "" {
		attrs = append(attrs, "request_id", id)
	}

	if cookie, err := c.Cookie("client_id"); err == nil {
		attrs = append(attrs, "client_id", cookie.Value)
	}

	return slog.With(attrs...)
}

func generateID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
