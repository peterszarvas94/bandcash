package utils

import "github.com/labstack/echo/v4"

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
