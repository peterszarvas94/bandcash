package utils

import (
	"context"
	"strings"

	"github.com/labstack/echo/v4"
)

type tabIDContextKey string

const tabIDKey tabIDContextKey = "tab_id"

func WithTabID(ctx context.Context, tabID string) context.Context {
	return context.WithValue(ctx, tabIDKey, tabID)
}

func TabIDFromContext(ctx context.Context) string {
	if id, ok := ctx.Value(tabIDKey).(string); ok {
		return id
	}
	return ""
}

// EnsureTabIDFromContext returns existing tab ID or creates a new one.
func EnsureTabIDFromContext(ctx context.Context) string {
	if tabID := TabIDFromContext(ctx); tabID != "" {
		return tabID
	}
	return GenerateID("tab")
}

// SetTabID updates request context with a valid tab ID.
func SetTabID(c echo.Context, tabID string) bool {
	tabID = strings.TrimSpace(tabID)
	if !IsValidID(tabID, "tab") {
		return false
	}
	c.SetRequest(c.Request().WithContext(WithTabID(c.Request().Context(), tabID)))
	return true
}

// EnsureTabID returns existing tab ID or creates a new one.
func EnsureTabID(c echo.Context) string {
	if tabID := TabIDFromContext(c.Request().Context()); tabID != "" {
		return tabID
	}
	if requestedTabID := strings.TrimSpace(c.QueryParam("tab_id")); IsValidID(requestedTabID, "tab") {
		c.SetRequest(c.Request().WithContext(WithTabID(c.Request().Context(), requestedTabID)))
		return requestedTabID
	}
	tabID := GenerateID("tab")
	c.SetRequest(c.Request().WithContext(WithTabID(c.Request().Context(), tabID)))
	return tabID
}
