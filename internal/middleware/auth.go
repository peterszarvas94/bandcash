package middleware

import (
	"log/slog"
	"net/http"
	"strings"

	ctxi18nlib "github.com/invopop/ctxi18n"
	ctxi18n "github.com/invopop/ctxi18n/i18n"
	"github.com/labstack/echo/v4"

	appi18n "bandcash/internal/i18n"
	"bandcash/internal/utils"
	authstore "bandcash/models/auth/data"
	groupstore "bandcash/models/group/data"
)

// RequireAuth ensures user is logged in.
func RequireAuth(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		userID := getSessionUserID(c)
		if userID == "" {
			return c.Redirect(http.StatusFound, "/login")
		}

		// Verify user exists
		user, err := authstore.GetUserByID(c.Request().Context(), userID)
		if err != nil {
			slog.Warn("auth: invalid session user", "user_id", userID)
			clearSession(c)
			return c.Redirect(http.StatusFound, "/login")
		}

		bannedCount, err := authstore.IsUserBanned(c.Request().Context(), user.ID)
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

		isSuperadmin := utils.EmailMatchesSuperadmin(user.Email)

		preferredLang := appi18n.NormalizeLocale(user.PreferredLang)
		if rawLang := strings.TrimSpace(c.QueryParam("lang")); rawLang != "" {
			preferredLang = appi18n.NormalizeLocale(rawLang)
		}
		utils.SetLocaleCookie(c, preferredLang)
		if localizedCtx, localeErr := ctxi18nlib.WithLocale(c.Request().Context(), preferredLang); localeErr == nil {
			c.SetRequest(c.Request().WithContext(localizedCtx))
		}

		c.Set(utils.CtxUserIDKey, user.ID)
		c.Set(utils.CtxIsSuperadminKey, isSuperadmin)
		return next(c)
	}
}

// RequireGroup ensures user has access to the requested group.
func RequireGroup(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		userID := utils.GetUserID(c)
		groupID := c.Param("groupId")
		isSuperadmin := utils.IsSuperadmin(c)

		if !utils.IsValidID(groupID, "grp") {
			return c.String(http.StatusBadRequest, "Invalid group ID")
		}

		if isSuperadmin {
			c.Set(utils.CtxGroupIDKey, groupID)
			// Superadmin is treated as admin across all groups.
			c.Set(utils.CtxGroupRoleKey, "admin")
			return next(c)
		}

		role, err := groupstore.GetGroupAccessRole(c.Request().Context(), groupstore.GetGroupAccessRoleParams{
			UserID:  userID,
			GroupID: groupID,
		})
		if err != nil {
			utils.Notify(c, ctxi18n.T(c.Request().Context(), "groups.errors.access_denied"))
			return c.Redirect(http.StatusFound, "/groups")
		}

		c.Set(utils.CtxGroupIDKey, groupID)
		c.Set(utils.CtxGroupRoleKey, role)
		return next(c)
	}
}

// RequireAdmin ensures user is admin of the group.
func RequireAdmin(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if !utils.IsAdmin(c) {
			utils.Notify(c, ctxi18n.T(c.Request().Context(), "groups.errors.admin_required"))
			return c.Redirect(http.StatusFound, "/groups")
		}
		return next(c)
	}
}

// RequireOwner ensures user is owner of the group.
func RequireOwner(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if !utils.IsOwner(c) {
			utils.Notify(c, ctxi18n.T(c.Request().Context(), "groups.errors.owner_required"))
			return c.Redirect(http.StatusFound, "/groups")
		}
		return next(c)
	}
}

// RequireOwnerOrSuperadmin ensures user is owner or superadmin.
func RequireOwnerOrSuperadmin(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if utils.IsSuperadmin(c) {
			return next(c)
		}
		return RequireOwner(next)(c)
	}
}

func getSessionUserID(c echo.Context) string {
	cookie, err := c.Cookie(utils.SessionCookieName)
	if err != nil {
		return ""
	}

	session, err := authstore.GetUserSessionByToken(c.Request().Context(), cookie.Value)
	if err != nil {
		return ""
	}

	return session.UserID
}

func clearSession(c echo.Context) {
	cookie, err := c.Cookie(utils.SessionCookieName)
	if err == nil {
		// Try to get session and delete it from DB
		session, err := authstore.GetUserSessionByToken(c.Request().Context(), cookie.Value)
		if err == nil {
			userID := utils.GetUserID(c)
			if userID != "" {
				_ = authstore.DeleteUserSession(c.Request().Context(), authstore.DeleteUserSessionParams{
					ID:     session.ID,
					UserID: userID,
				})
			}
		}
	}
	utils.ClearSessionCookie(c)
}
