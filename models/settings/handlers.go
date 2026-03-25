package settings

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
)

type settingsSignals struct {
	TabID    string `json:"tab_id"`
	FormData struct {
		Lang string `json:"lang"`
	} `json:"formData"`
}

type settingsTabSignals struct {
	TabID string `json:"tab_id"`
}

func (s *Settings) Index(c echo.Context) error {
	utils.EnsureTabID(c)
	data := s.Data(c.Request().Context())
	userID := middleware.GetUserID(c)
	if user, err := db.Qry.GetUserByID(c.Request().Context(), userID); err == nil {
		data.UserEmail = user.Email
		data.CurrentLang = appi18n.NormalizeLocale(user.PreferredLang)
	}
	return utils.RenderPage(c, SettingsIndex(data))
}

func (s *Settings) LanguagePage(c echo.Context) error {
	utils.EnsureTabID(c)
	data := s.Data(c.Request().Context())
	data.Title = ctxi18n.T(c.Request().Context(), "settings.language")
	data.Breadcrumbs = []utils.Crumb{{Label: ctxi18n.T(c.Request().Context(), "settings.language")}}
	userID := middleware.GetUserID(c)
	if user, err := db.Qry.GetUserByID(c.Request().Context(), userID); err == nil {
		data.UserEmail = user.Email
		data.CurrentLang = appi18n.NormalizeLocale(user.PreferredLang)
	}
	return utils.RenderPage(c, LanguagePage(data))
}

func (s *Settings) UpdateLanguage(c echo.Context) error {
	signals := settingsSignals{}
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

func (s *Settings) UpdateDetailsState(c echo.Context) error {
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
		slog.Error("settings.details_state: failed to upsert", "user_id", userID, "key", key, "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}
	slog.Debug("settings.details_state: updated", "user_id", userID, "key", key, "open", open)

	return c.NoContent(http.StatusNoContent)
}

func (s *Settings) SessionsPage(c echo.Context) error {
	utils.EnsureTabID(c)
	userID := middleware.GetUserID(c)

	sessions, err := db.Qry.ListUserSessions(c.Request().Context(), userID)
	if err != nil {
		slog.Error("settings.sessions: failed to list sessions", "user_id", userID, "err", err)
		sessions = []db.UserSession{}
	}

	data := SessionsData{
		Title:            ctxi18n.T(c.Request().Context(), "settings.sessions"),
		Breadcrumbs:      []utils.Crumb{{Label: ctxi18n.T(c.Request().Context(), "settings.sessions")}},
		CurrentSessionID: "",
		Sessions:         sessions,
	}

	if cookie, err := c.Cookie(utils.SessionCookieName); err == nil {
		if session, err := db.Qry.GetUserSessionByToken(c.Request().Context(), cookie.Value); err == nil {
			data.CurrentSessionID = session.ID
		}
	}

	if user, err := db.Qry.GetUserByID(c.Request().Context(), userID); err == nil {
		data.UserEmail = user.Email
	}

	return utils.RenderPage(c, SessionsPage(data))
}

func (s *Settings) LogoutSession(c echo.Context) error {
	signals := settingsTabSignals{}
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
		slog.Error("settings.sessions: failed to delete session", "session_id", sessionID, "user_id", userID, "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	utils.Notify(c, "success", ctxi18n.T(c.Request().Context(), "settings.notifications.session_logged_out"))
	return c.NoContent(http.StatusOK)
}

func (s *Settings) LogoutAllOtherSessions(c echo.Context) error {
	signals := settingsTabSignals{}
	if err := datastar.ReadSignals(c.Request(), &signals); err != nil {
		return c.NoContent(http.StatusBadRequest)
	}
	if !utils.SetTabID(c, signals.TabID) {
		return c.NoContent(http.StatusBadRequest)
	}

	userID := middleware.GetUserID(c)

	// Get current session from cookie
	cookie, err := c.Cookie(utils.SessionCookieName)
	if err != nil {
		return c.NoContent(http.StatusBadRequest)
	}

	currentSession, err := db.Qry.GetUserSessionByToken(c.Request().Context(), cookie.Value)
	if err != nil {
		return c.NoContent(http.StatusBadRequest)
	}

	err = db.Qry.DeleteOtherUserSessions(c.Request().Context(), db.DeleteOtherUserSessionsParams{
		UserID: userID,
		ID:     currentSession.ID,
	})
	if err != nil {
		slog.Error("settings.sessions: failed to delete other sessions", "user_id", userID, "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	utils.Notify(c, "success", ctxi18n.T(c.Request().Context(), "settings.notifications.other_sessions_logged_out"))
	return c.NoContent(http.StatusOK)
}
