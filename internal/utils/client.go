package utils

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

type clientIDContextKey string

const clientIDKey clientIDContextKey = "client_id"

func withClientID(ctx context.Context, clientID string) context.Context {
	return context.WithValue(ctx, clientIDKey, clientID)
}

func ClientIDFromContext(ctx context.Context) string {
	if id, ok := ctx.Value(clientIDKey).(string); ok {
		return id
	}
	return ""
}

// EnsureClientID returns the client ID from the cookie, or generates a new one and sets the cookie.
func EnsureClientID(c echo.Context) string {
	cookie, err := c.Cookie("client_id")
	if err == nil {
		c.SetRequest(c.Request().WithContext(withClientID(c.Request().Context(), cookie.Value)))
		return cookie.Value
	}

	clientID := fmt.Sprintf("client_%d", time.Now().UnixNano())
	c.SetCookie(&http.Cookie{
		Name:     "client_id",
		Value:    clientID,
		Path:     "/",
		HttpOnly: true,
		MaxAge:   86400,
	})
	c.SetRequest(c.Request().WithContext(withClientID(c.Request().Context(), clientID)))
	return clientID
}

// GetClientID returns the client ID from the cookie or an error if not set.
func GetClientID(c echo.Context) (string, error) {
	cookie, err := c.Cookie("client_id")
	if err != nil {
		return "", errors.New("no client_id cookie")
	}
	return cookie.Value, nil
}
