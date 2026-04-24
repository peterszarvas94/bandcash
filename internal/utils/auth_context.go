package utils

import (
	"strings"

	"github.com/labstack/echo/v4"

	authstore "bandcash/models/auth/data"
)

const (
	CtxUserIDKey       = "user_id"
	CtxGroupIDKey      = "group_id"
	CtxGroupRoleKey    = "group_role"
	CtxIsSuperadminKey = "is_superadmin"
)

func GetUserID(c echo.Context) string {
	if id, ok := c.Get(CtxUserIDKey).(string); ok {
		return id
	}
	return ""
}

func GetGroupID(c echo.Context) string {
	if id, ok := c.Get(CtxGroupIDKey).(string); ok {
		return id
	}
	return ""
}

func GetGroupRole(c echo.Context) string {
	if role, ok := c.Get(CtxGroupRoleKey).(string); ok {
		return role
	}
	return ""
}

func IsOwner(c echo.Context) bool {
	return GetGroupRole(c) == "owner"
}

func IsAdmin(c echo.Context) bool {
	role := GetGroupRole(c)
	return role == "owner" || role == "admin"
}

func IsSuperadmin(c echo.Context) bool {
	if isSuperadmin, ok := c.Get(CtxIsSuperadminKey).(bool); ok {
		return isSuperadmin
	}
	return false
}

// EmailMatchesSuperadmin returns true if email equals SUPERADMIN_EMAIL (case-insensitive). Empty config never matches.
func EmailMatchesSuperadmin(email string) bool {
	want := strings.ToLower(strings.TrimSpace(Env().SuperadminEmail))
	if want == "" {
		return false
	}
	return strings.ToLower(strings.TrimSpace(email)) == want
}

func ResolveAuthState(c echo.Context) (bool, bool) {
	if GetUserID(c) != "" {
		return true, IsSuperadmin(c)
	}

	cookie, err := c.Cookie(SessionCookieName)
	if err != nil || strings.TrimSpace(cookie.Value) == "" {
		return false, false
	}

	session, err := authstore.GetUserSessionByToken(c.Request().Context(), cookie.Value)
	if err != nil {
		return false, false
	}

	user, err := authstore.GetUserByID(c.Request().Context(), session.UserID)
	if err != nil {
		return false, false
	}

	bannedCount, err := authstore.IsUserBanned(c.Request().Context(), user.ID)
	if err != nil || bannedCount > 0 {
		return false, false
	}

	return true, EmailMatchesSuperadmin(user.Email)
}
