package settings

import (
	"net/http"

	ctxi18nlib "github.com/invopop/ctxi18n"
	ctxi18n "github.com/invopop/ctxi18n/i18n"
	"github.com/labstack/echo/v4"
	"github.com/starfederation/datastar-go/datastar"

	"bandcash/internal/db"
	appi18n "bandcash/internal/i18n"
	"bandcash/internal/utils"
)

type settingsSignals struct {
	FormData struct {
		Lang string `json:"lang"`
	} `json:"formData"`
}

func (s *Settings) Index(c echo.Context) error {
	utils.EnsureClientID(c)
	data := s.Data(c.Request().Context())
	if cookie, err := c.Cookie("session"); err == nil && cookie.Value != "" {
		if user, err := db.Qry.GetUserByID(c.Request().Context(), cookie.Value); err == nil {
			data.UserEmail = user.Email
		}
	}
	return utils.RenderPage(c, SettingsIndex(data))
}

func (s *Settings) LanguagePage(c echo.Context) error {
	utils.EnsureClientID(c)
	data := s.Data(c.Request().Context())
	data.Title = ctxi18n.T(c.Request().Context(), "settings.language")
	data.Breadcrumbs = []utils.Crumb{{Label: ctxi18n.T(c.Request().Context(), "settings.language")}}
	if cookie, err := c.Cookie("session"); err == nil && cookie.Value != "" {
		if user, err := db.Qry.GetUserByID(c.Request().Context(), cookie.Value); err == nil {
			data.UserEmail = user.Email
		}
	}
	return utils.RenderPage(c, LanguagePage(data))
}

func (s *Settings) UpdateLanguage(c echo.Context) error {
	signals := settingsSignals{}
	if err := datastar.ReadSignals(c.Request(), &signals); err != nil {
		return c.NoContent(http.StatusBadRequest)
	}

	locale := appi18n.NormalizeLocale(signals.FormData.Lang)
	if sessionCookie, err := c.Cookie("session"); err == nil && sessionCookie.Value != "" {
		if err := db.Qry.UpdateUserPreferredLang(c.Request().Context(), db.UpdateUserPreferredLangParams{PreferredLang: locale, ID: sessionCookie.Value}); err != nil {
			return c.NoContent(http.StatusInternalServerError)
		}
	}
	notifyCtx, err := ctxi18nlib.WithLocale(c.Request().Context(), locale)
	if err != nil {
		notifyCtx = c.Request().Context()
	}
	utils.Notify(c, "success", ctxi18n.T(notifyCtx, "settings.notifications.language_saved"))
	err = utils.SSEHub.ExecuteScript(c, "window.location.reload()")
	if err != nil {
		return c.NoContent(http.StatusInternalServerError)
	}
	return c.NoContent(http.StatusOK)
}
