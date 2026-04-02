package account

import (
	"log/slog"
	"net/http"
	"strings"

	ctxi18nlib "github.com/invopop/ctxi18n"
	ctxi18n "github.com/invopop/ctxi18n/i18n"
	"github.com/labstack/echo/v4"
	"github.com/starfederation/datastar-go/datastar"

	"bandcash/internal/db"
	appi18n "bandcash/internal/i18n"
	"bandcash/internal/middleware"
	"bandcash/internal/utils"
	shared "bandcash/models/shared"
)

type accountSignals struct {
	TabID    string `json:"tab_id"`
	FormData struct {
		Lang string `json:"lang"`
	} `json:"formData"`
}

type accountTabSignals struct {
	TabID string `json:"tab_id"`
}

func (s *Account) Index(c echo.Context) error {
	utils.EnsureTabID(c)
	data := s.Data(c.Request().Context())
	userID := middleware.GetUserID(c)
	if user, err := db.Qry.GetUserByID(c.Request().Context(), userID); err == nil {
		data.UserEmail = user.Email
		data.CurrentLang = appi18n.NormalizeLocale(user.PreferredLang)
	}
	data.Signals = map[string]any{"formData": map[string]any{"lang": data.CurrentLang}}
	data.IsAuthenticated = true
	data.IsSuperAdmin = middleware.IsSuperadmin(c)
	return utils.RenderPage(c, AccountIndex(data))
}

func (s *Account) LanguagePage(c echo.Context) error {
	utils.EnsureTabID(c)
	data := s.Data(c.Request().Context())
	data.Title = ctxi18n.T(c.Request().Context(), "account.language")
	data.Breadcrumbs = []utils.Crumb{{Label: ctxi18n.T(c.Request().Context(), "account.language")}}
	userID := middleware.GetUserID(c)
	if user, err := db.Qry.GetUserByID(c.Request().Context(), userID); err == nil {
		data.UserEmail = user.Email
		data.CurrentLang = appi18n.NormalizeLocale(user.PreferredLang)
	}
	data.Signals = map[string]any{"formData": map[string]any{"lang": data.CurrentLang}}
	data.IsAuthenticated = true
	data.IsSuperAdmin = middleware.IsSuperadmin(c)
	return utils.RenderPage(c, LanguagePage(data))
}

func (s *Account) UpdateLanguage(c echo.Context) error {
	signals := accountSignals{}
	if err := datastar.ReadSignals(c.Request(), &signals); err != nil {
		return c.NoContent(http.StatusBadRequest)
	}
	if !utils.SetTabID(c, signals.TabID) {
		return c.NoContent(http.StatusBadRequest)
	}

	locale := appi18n.NormalizeLocale(signals.FormData.Lang)
	if err := db.Qry.UpdateUserPreferredLang(c.Request().Context(), db.UpdateUserPreferredLangParams{PreferredLang: locale, ID: middleware.GetUserID(c)}); err != nil {
		return c.NoContent(http.StatusInternalServerError)
	}
	utils.SetLocaleCookie(c, locale)
	notifyCtx, err := ctxi18nlib.WithLocale(c.Request().Context(), locale)
	if err != nil {
		notifyCtx = c.Request().Context()
	}
	utils.Notify(c, ctxi18n.T(notifyCtx, "account.notifications.language_saved"))
	err = utils.SSEHub.ExecuteScript(c, "window.location.reload()")
	if err != nil {
		return c.NoContent(http.StatusInternalServerError)
	}
	return c.NoContent(http.StatusOK)
}

func (s *Account) UpdateDetailsState(c echo.Context) error {
	key := strings.TrimSpace(c.QueryParam("key"))
	if key == "" {
		return c.NoContent(http.StatusBadRequest)
	}

	open := c.QueryParam("open") == "1"
	userID := middleware.GetUserID(c)
	err := db.Qry.UpsertUserDetailCardState(c.Request().Context(), db.UpsertUserDetailCardStateParams{
		UserID:   userID,
		StateKey: key,
		IsOpen: func() int64 {
			if open {
				return 1
			}
			return 0
		}(),
	})
	if err != nil {
		slog.Error("account.details_state: failed to upsert", "user_id", userID, "key", key, "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}
	slog.Debug("account.details_state: updated", "user_id", userID, "key", key, "open", open)

	return c.NoContent(http.StatusNoContent)
}

func (s *Account) SessionsPage(c echo.Context) error {
	utils.EnsureTabID(c)
	userID := middleware.GetUserID(c)

	sessions, err := db.Qry.ListUserSessions(c.Request().Context(), userID)
	if err != nil {
		slog.Error("account.sessions: failed to list sessions", "user_id", userID, "err", err)
		sessions = []db.UserSession{}
	}

	data := SessionsData{
		Title:            ctxi18n.T(c.Request().Context(), "account.account"),
		Breadcrumbs:      []utils.Crumb{{Label: ctxi18n.T(c.Request().Context(), "account.account")}},
		CurrentSessionID: "",
		Sessions:         sessions,
	}

	if cookie, err := c.Cookie(utils.SessionCookieName); err == nil {
		if session, err := db.Qry.GetUserSessionByToken(c.Request().Context(), cookie.Value); err == nil {
			data.CurrentSessionID = session.ID
		}
	}

	data.Signals = nil
	data.IsAuthenticated = true
	data.IsSuperAdmin = middleware.IsSuperadmin(c)

	return utils.RenderPage(c, SessionsPage(data))
}

func (s *Account) LogoutSession(c echo.Context) error {
	signals := accountTabSignals{}
	if err := datastar.ReadSignals(c.Request(), &signals); err != nil {
		return c.NoContent(http.StatusBadRequest)
	}
	if !utils.SetTabID(c, signals.TabID) {
		return c.NoContent(http.StatusBadRequest)
	}

	sessionID := c.Param("id")
	if !utils.IsValidID(sessionID, "ses") {
		return c.NoContent(http.StatusBadRequest)
	}

	userID := middleware.GetUserID(c)
	err := db.Qry.DeleteUserSession(c.Request().Context(), db.DeleteUserSessionParams{
		ID:     sessionID,
		UserID: userID,
	})
	if err != nil {
		slog.Error("account.sessions: failed to delete session", "session_id", sessionID, "user_id", userID, "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	utils.Notify(c, ctxi18n.T(c.Request().Context(), "account.notifications.session_logged_out"))
	return c.NoContent(http.StatusOK)
}

func (s *Account) LogoutAllOtherSessions(c echo.Context) error {
	signals := accountTabSignals{}
	if err := datastar.ReadSignals(c.Request(), &signals); err != nil {
		return c.NoContent(http.StatusBadRequest)
	}
	if !utils.SetTabID(c, signals.TabID) {
		return c.NoContent(http.StatusBadRequest)
	}

	userID := middleware.GetUserID(c)

	if err := db.Qry.DeleteAllUserSessions(c.Request().Context(), userID); err != nil {
		slog.Error("account.sessions: failed to delete all sessions", "user_id", userID, "err", err)
		utils.Notify(c, ctxi18n.T(c.Request().Context(), "account.notifications.logout_everywhere_failed"))
		if notificationsHTML, renderErr := utils.RenderHTMLForRequest(c, shared.Notifications()); renderErr == nil {
			_ = utils.SSEHub.PatchHTML(c, notificationsHTML)
		}
		return c.NoContent(http.StatusInternalServerError)
	}

	utils.ClearSessionCookie(c)
	if err := utils.SSEHub.Redirect(c, "/"); err != nil {
		return c.Redirect(http.StatusFound, "/")
	}

	return c.NoContent(http.StatusOK)
}
