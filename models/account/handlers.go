package account

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	ctxi18nlib "github.com/invopop/ctxi18n"
	ctxi18n "github.com/invopop/ctxi18n/i18n"
	"github.com/labstack/echo/v4"
	"github.com/starfederation/datastar-go/datastar"

	internalbilling "bandcash/internal/billing"
	"bandcash/internal/db"
	appi18n "bandcash/internal/i18n"
	"bandcash/internal/utils"
	authstore "bandcash/models/auth/data"
	shared "bandcash/models/shared"
)

type accountSignals struct {
	TabID    string `json:"tab_id"`
	FormData struct {
		Lang string `json:"lang"`
	} `json:"formData"`
}

type flexInt int

func (v *flexInt) UnmarshalJSON(data []byte) error {
	var asInt int
	if err := json.Unmarshal(data, &asInt); err == nil {
		*v = flexInt(asInt)
		return nil
	}
	var asString string
	if err := json.Unmarshal(data, &asString); err == nil {
		parsed, parseErr := strconv.Atoi(strings.TrimSpace(asString))
		if parseErr != nil {
			return parseErr
		}
		*v = flexInt(parsed)
		return nil
	}
	return strconv.ErrSyntax
}

type accountSeatSignals struct {
	TabID    string `json:"tab_id"`
	FormData struct {
		SeatQuantity flexInt `json:"seatQuantity"`
	} `json:"formData"`
}

type accountTabSignals struct {
	TabID string `json:"tab_id"`
}

func Index(c echo.Context) error {
	utils.EnsureTabID(c)
	data := Data(c.Request().Context())
	userID := utils.GetUserID(c)
	if user, err := authstore.GetUserByID(c.Request().Context(), userID); err == nil {
		data.UserID = user.ID
		data.UserEmail = user.Email
		data.CurrentLang = appi18n.NormalizeLocale(user.PreferredLang)
	}

	if state, err := internalbilling.CurrentAccessState(c.Request().Context(), userID); err == nil {
		data.SubscriptionSlots = state.SubscriptionCount
		data.UsedSlots = state.OwnedGroupCount
		data.RemainingSlots = state.RemainingSlots
	}
	if sub, exists, err := internalbilling.GetUserSubscription(c.Request().Context(), userID); err == nil && exists {
		if strings.TrimSpace(sub.ProviderSubscriptionID) != "" && (strings.TrimSpace(sub.ProviderSubscriptionItemID) == "" || strings.TrimSpace(sub.ProviderPortalURL) == "") {
			if synced, syncedExists, syncErr := internalbilling.SyncSubscriptionFromProvider(c.Request().Context(), userID); syncErr == nil && syncedExists {
				sub = synced
			}
		}
		data.SeatQuantity = sub.SeatQuantity
		if data.SeatQuantity < 1 {
			data.SeatQuantity = 1
		}
		data.CanUpdateSeats = strings.TrimSpace(sub.ProviderSubscriptionItemID) != ""
		if sub.ProviderPortalURL != "" {
			data.SubscriptionPortalURL = sub.ProviderPortalURL
		}
	}
	if data.SeatQuantity < 1 {
		data.SeatQuantity = data.SubscriptionSlots
		if data.SeatQuantity < 1 {
			data.SeatQuantity = 1
		}
	}

	data.Signals = map[string]any{"formData": map[string]any{"lang": data.CurrentLang, "seatQuantity": data.SeatQuantity}}
	data.IsAuthenticated = true
	data.IsSuperAdmin = utils.IsSuperadmin(c)
	return utils.RenderPage(c, AccountIndex(data))
}

func UpdateSeats(c echo.Context) error {
	signals := accountSeatSignals{}
	if err := datastar.ReadSignals(c.Request(), &signals); err != nil {
		return c.NoContent(http.StatusBadRequest)
	}
	if !utils.SetTabID(c, signals.TabID) {
		return c.NoContent(http.StatusBadRequest)
	}

	quantity := int(signals.FormData.SeatQuantity)
	if quantity < 1 {
		utils.Notify(c, ctxi18n.T(c.Request().Context(), "account.notifications.seats_invalid"))
		return c.NoContent(http.StatusBadRequest)
	}

	userID := utils.GetUserID(c)
	state, err := internalbilling.CurrentAccessState(c.Request().Context(), userID)
	if err != nil {
		slog.Error("account.seats: failed to load access state", "user_id", userID, "err", err)
		utils.Notify(c, ctxi18n.T(c.Request().Context(), "account.notifications.seats_update_failed"))
		return c.NoContent(http.StatusInternalServerError)
	}
	if quantity < state.OwnedGroupCount {
		utils.Notify(c, ctxi18n.T(c.Request().Context(), "account.notifications.seats_below_used"))
		return c.NoContent(http.StatusBadRequest)
	}

	subscription, exists, err := internalbilling.GetUserSubscription(c.Request().Context(), userID)
	if err != nil {
		slog.Error("account.seats: failed to load subscription", "user_id", userID, "err", err)
		utils.Notify(c, ctxi18n.T(c.Request().Context(), "account.notifications.seats_update_failed"))
		return c.NoContent(http.StatusInternalServerError)
	}
	if !exists {
		utils.Notify(c, ctxi18n.T(c.Request().Context(), "account.notifications.subscription_missing"))
		return c.NoContent(http.StatusBadRequest)
	}
	if strings.TrimSpace(subscription.ProviderSubscriptionItemID) == "" {
		synced, syncedExists, syncErr := internalbilling.SyncSubscriptionFromProvider(c.Request().Context(), userID)
		if syncErr != nil || !syncedExists {
			utils.Notify(c, ctxi18n.T(c.Request().Context(), "account.seats_update_unavailable"))
			return c.NoContent(http.StatusBadRequest)
		}
		subscription = synced
		if strings.TrimSpace(subscription.ProviderSubscriptionItemID) == "" {
			utils.Notify(c, ctxi18n.T(c.Request().Context(), "account.seats_update_unavailable"))
			return c.NoContent(http.StatusBadRequest)
		}
	}

	if err := internalbilling.UpdateSubscriptionItemQuantity(c.Request().Context(), subscription.ProviderSubscriptionItemID, quantity); err != nil {
		slog.Error("account.seats: lemon update failed", "user_id", userID, "subscription_item_id", subscription.ProviderSubscriptionItemID, "quantity", quantity, "err", err)
		utils.Notify(c, ctxi18n.T(c.Request().Context(), "account.notifications.seats_update_failed"))
		return c.NoContent(http.StatusBadGateway)
	}

	utils.Notify(c, ctxi18n.T(c.Request().Context(), "account.notifications.seats_update_requested"))
	return c.NoContent(http.StatusOK)
}
func LanguagePageHandler(c echo.Context) error {
	utils.EnsureTabID(c)
	data := Data(c.Request().Context())
	data.Title = ctxi18n.T(c.Request().Context(), "account.language")
	data.Breadcrumbs = []utils.Crumb{{Label: ctxi18n.T(c.Request().Context(), "account.language")}}
	userID := utils.GetUserID(c)
	if user, err := authstore.GetUserByID(c.Request().Context(), userID); err == nil {
		data.UserEmail = user.Email
		data.CurrentLang = appi18n.NormalizeLocale(user.PreferredLang)
	}
	data.Signals = map[string]any{"formData": map[string]any{"lang": data.CurrentLang}}
	data.IsAuthenticated = true
	data.IsSuperAdmin = utils.IsSuperadmin(c)
	return utils.RenderPage(c, LanguagePage(data))
}

func UpdateLanguage(c echo.Context) error {
	signals := accountSignals{}
	if err := datastar.ReadSignals(c.Request(), &signals); err != nil {
		return c.NoContent(http.StatusBadRequest)
	}
	if !utils.SetTabID(c, signals.TabID) {
		return c.NoContent(http.StatusBadRequest)
	}

	locale := appi18n.NormalizeLocale(signals.FormData.Lang)
	if err := authstore.UpdateUserPreferredLang(c.Request().Context(), authstore.UpdateUserPreferredLangParams{PreferredLang: locale, ID: utils.GetUserID(c)}); err != nil {
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

func SessionsPageHandler(c echo.Context) error {
	utils.EnsureTabID(c)
	userID := utils.GetUserID(c)

	sessions, err := authstore.ListUserSessions(c.Request().Context(), userID)
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
		if session, err := authstore.GetUserSessionByToken(c.Request().Context(), cookie.Value); err == nil {
			data.CurrentSessionID = session.ID
		}
	}

	data.Signals = nil
	data.IsAuthenticated = true
	data.IsSuperAdmin = utils.IsSuperadmin(c)

	return utils.RenderPage(c, SessionsPage(data))
}

func LogoutSession(c echo.Context) error {
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

	userID := utils.GetUserID(c)
	err := authstore.DeleteUserSession(c.Request().Context(), authstore.DeleteUserSessionParams{
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

func LogoutAllOtherSessions(c echo.Context) error {
	signals := accountTabSignals{}
	if err := datastar.ReadSignals(c.Request(), &signals); err != nil {
		return c.NoContent(http.StatusBadRequest)
	}
	if !utils.SetTabID(c, signals.TabID) {
		return c.NoContent(http.StatusBadRequest)
	}

	userID := utils.GetUserID(c)

	if err := authstore.DeleteAllUserSessions(c.Request().Context(), userID); err != nil {
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
