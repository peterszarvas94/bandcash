package middleware

import (
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"

	"bandcash/internal/db"
)

const (
	UserIDKey  contextKey = "user_id"
	GroupIDKey contextKey = "group_id"
	IsAdminKey contextKey = "is_admin"
)

// RequireAuth ensures user is logged in
func RequireAuth() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			userID := getSessionUserID(c)
			if userID == "" {
				return c.Redirect(http.StatusFound, "/auth/login")
			}

			// Verify user exists
			user, err := db.Qry.GetUserByID(c.Request().Context(), userID)
			if err != nil {
				slog.Warn("auth: invalid session user", "user_id", userID)
				clearSession(c)
				return c.Redirect(http.StatusFound, "/auth/login")
			}

			c.Set(string(UserIDKey), user.ID)
			return next(c)
		}
	}
}

// RequireGroup ensures user has access to the requested group
func RequireGroup() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			userID := c.Get(string(UserIDKey)).(string)
			groupID := c.Param("groupId")

			if groupID == "" {
				return c.String(http.StatusBadRequest, "Group ID required")
			}

			// Check if admin
			group, err := db.Qry.GetGroupByID(c.Request().Context(), groupID)
			if err == nil && group.AdminUserID == userID {
				c.Set(string(GroupIDKey), groupID)
				c.Set(string(IsAdminKey), true)
				return next(c)
			}

			// Check if reader
			readerCount, err := db.Qry.IsGroupReader(c.Request().Context(), db.IsGroupReaderParams{
				UserID:  userID,
				GroupID: groupID,
			})
			if err != nil || readerCount == 0 {
				return c.String(http.StatusForbidden, "Access denied")
			}

			c.Set(string(GroupIDKey), groupID)
			c.Set(string(IsAdminKey), false)
			return next(c)
		}
	}
}

// RequireAdmin ensures user is admin of the group
func RequireAdmin() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			isAdmin, ok := c.Get(string(IsAdminKey)).(bool)
			if !ok || !isAdmin {
				return c.String(http.StatusForbidden, "Admin access required")
			}
			return next(c)
		}
	}
}

// GetUserID retrieves user ID from context
func GetUserID(c echo.Context) string {
	if id, ok := c.Get(string(UserIDKey)).(string); ok {
		return id
	}
	return ""
}

// GetGroupID retrieves group ID from context
func GetGroupID(c echo.Context) string {
	if id, ok := c.Get(string(GroupIDKey)).(string); ok {
		return id
	}
	return ""
}

// IsAdmin checks if user is admin
func IsAdmin(c echo.Context) bool {
	if isAdmin, ok := c.Get(string(IsAdminKey)).(bool); ok {
		return isAdmin
	}
	return false
}

func getSessionUserID(c echo.Context) string {
	cookie, err := c.Cookie("session")
	if err != nil {
		return ""
	}
	return cookie.Value
}

func clearSession(c echo.Context) {
	c.SetCookie(&http.Cookie{
		Name:     "session",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	})
}
