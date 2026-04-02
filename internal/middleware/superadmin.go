package middleware

import (
	"net/http"
	"strings"

	ctxi18n "github.com/invopop/ctxi18n/i18n"
	"github.com/labstack/echo/v4"

	"bandcash/internal/db"
	"bandcash/internal/utils"
)

// RequireSuperadmin ensures user email matches configured superadmin email.
func RequireSuperadmin(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		allowedEmail := strings.ToLower(strings.TrimSpace(utils.Env().SuperadminEmail))
		if allowedEmail == "" {
			utils.Notify(c, ctxi18n.T(c.Request().Context(), "admin.access_denied"))
			return c.Redirect(http.StatusFound, "/groups")
		}

		userID := GetUserID(c)
		if userID == "" {
			return c.Redirect(http.StatusFound, "/login")
		}

		user, err := db.Qry.GetUserByID(c.Request().Context(), userID)
		if err != nil {
			return c.Redirect(http.StatusFound, "/login")
		}

		if strings.ToLower(strings.TrimSpace(user.Email)) != allowedEmail {
			utils.Notify(c, ctxi18n.T(c.Request().Context(), "admin.access_denied"))
			return c.Redirect(http.StatusFound, "/groups")
		}

		return next(c)
	}
}
