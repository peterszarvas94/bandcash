package middleware

import (
	"log/slog"
	"net/http"
	"strings"

	ctxi18nlib "github.com/invopop/ctxi18n"
	ctxi18n "github.com/invopop/ctxi18n/i18n"
	"github.com/labstack/echo/v4"

	"bandcash/internal/db"
	appi18n "bandcash/internal/i18n"
	"bandcash/internal/utils"
)

const (
	UserIDKey       contextKey = "user_id"
	GroupIDKey      contextKey = "group_id"
	GroupRoleKey    contextKey = "group_role"
	IsSuperadminKey contextKey = "is_superadmin"
)

// RequireAuth ensures user is logged in.
func RequireAuth(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		userID := getSessionUserID(c)
		if userID == "" {
			return c.Redirect(http.StatusFound, "/login")
		}

		// Verify user exists
		user, err := db.GetUserByID(c.Request().Context(), userID)
		if err != nil {
			slog.Warn("auth: invalid session user", "user_id", userID)
			clearSession(c)
			return c.Redirect(http.StatusFound, "/login")
		}

		bannedCount, err := db.IsUserBanned(c.Request().Context(), user.ID)
		if err != nil {
			slog.Warn("auth: failed to check user ban", "user_id", user.ID, "err", err)
			clearSession(c)
			return c.Redirect(http.StatusFound, "/login")
		}
		if bannedCount > 0 {
			clearSession(c)
			utils.Notify(c, ctxi18n.T(c.Request().Context(), "auth.banned"))
			return c.Redirect(http.StatusFound, "/login")
		}

		isSuperadmin := false
		superadminEmail := strings.ToLower(strings.TrimSpace(utils.Env().SuperadminEmail))
		if superadminEmail != "" && strings.ToLower(strings.TrimSpace(user.Email)) == superadminEmail {
			isSuperadmin = true
		}

		preferredLang := appi18n.NormalizeLocale(user.PreferredLang)
		if rawLang := strings.TrimSpace(c.QueryParam("lang")); rawLang != "" {
			preferredLang = appi18n.NormalizeLocale(rawLang)
		}
		utils.SetLocaleCookie(c, preferredLang)
		if localizedCtx, localeErr := ctxi18nlib.WithLocale(c.Request().Context(), preferredLang); localeErr == nil {
			c.SetRequest(c.Request().WithContext(localizedCtx))
		}

		c.Set(string(UserIDKey), user.ID)
		c.Set(string(IsSuperadminKey), isSuperadmin)
		return next(c)
	}
}

// RequireGroup ensures user has access to the requested group.
func RequireGroup(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		userID := c.Get(string(UserIDKey)).(string)
		groupID := c.Param("groupId")
		isSuperadmin := IsSuperadmin(c)

		if !utils.IsValidID(groupID, "grp") {
			return c.String(http.StatusBadRequest, "Invalid group ID")
		}

		if isSuperadmin {
			c.Set(string(GroupIDKey), groupID)
			// Superadmin is treated as admin across all groups.
			c.Set(string(GroupRoleKey), "admin")
			return next(c)
		}

		role, err := db.GetGroupAccessRole(c.Request().Context(), db.GetGroupAccessRoleParams{
			UserID:  userID,
			GroupID: groupID,
		})
		if err != nil {
			utils.Notify(c, ctxi18n.T(c.Request().Context(), "groups.errors.access_denied"))
			return c.Redirect(http.StatusFound, "/groups")
		}

		c.Set(string(GroupIDKey), groupID)
		c.Set(string(GroupRoleKey), role)
		return next(c)
	}
}

// RequireAdmin ensures user is admin of the group.
func RequireAdmin(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if !IsAdmin(c) {
			utils.Notify(c, ctxi18n.T(c.Request().Context(), "groups.errors.admin_required"))
			return c.Redirect(http.StatusFound, "/groups")
		}
		return next(c)
	}
}

// RequireOwner ensures user is owner of the group.
func RequireOwner(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if !IsOwner(c) {
			utils.Notify(c, ctxi18n.T(c.Request().Context(), "groups.errors.owner_required"))
			return c.Redirect(http.StatusFound, "/groups")
		}
		return next(c)
	}
}

// RequireOwnerOrSuperadmin ensures user is owner or superadmin.
func RequireOwnerOrSuperadmin(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if IsSuperadmin(c) {
			return next(c)
		}
		return RequireOwner(next)(c)
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
	role := GetGroupRole(c)
	return role == "owner" || role == "admin"
}

// GetGroupRole retrieves current user's role in active group from context.
func GetGroupRole(c echo.Context) string {
	if role, ok := c.Get(string(GroupRoleKey)).(string); ok {
		return role
	}
	return ""
}

// IsOwner checks if current user is owner in active group.
func IsOwner(c echo.Context) bool {
	return GetGroupRole(c) == "owner"
}

func IsSuperadmin(c echo.Context) bool {
	if isSuperadmin, ok := c.Get(string(IsSuperadminKey)).(bool); ok {
		return isSuperadmin
	}
	return false
}

func getSessionUserID(c echo.Context) string {
	cookie, err := c.Cookie(utils.SessionCookieName)
	if err != nil {
		return ""
	}

	session, err := db.GetUserSessionByToken(c.Request().Context(), cookie.Value)
	if err != nil {
		return ""
	}

	return session.UserID
}

func clearSession(c echo.Context) {
	cookie, err := c.Cookie(utils.SessionCookieName)
	if err == nil {
		// Try to get session and delete it from DB
		session, err := db.GetUserSessionByToken(c.Request().Context(), cookie.Value)
		if err == nil {
			userID := GetUserID(c)
			if userID != "" {
				_ = db.DeleteUserSession(c.Request().Context(), db.DeleteUserSessionParams{
					ID:     session.ID,
					UserID: userID,
				})
			}
		}
	}
	utils.ClearSessionCookie(c)
}
