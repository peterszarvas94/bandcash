package account

import (
	"database/sql"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	ctxi18nlib "github.com/invopop/ctxi18n"
	ctxi18n "github.com/invopop/ctxi18n/i18n"
	"github.com/labstack/echo/v4"
	"github.com/starfederation/datastar-go/datastar"

	internalbilling "bandcash/internal/billing"
	"bandcash/internal/db"
	"bandcash/internal/flags"
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

type accountTabSignals struct {
	TabID string `json:"tab_id"`
}

func Index(c echo.Context) error {
	return c.Redirect(http.StatusFound, "/account/subscription")
}

func SubscriptionPageHandler(c echo.Context) error {
	utils.EnsureTabID(c)
	data := Data(c.Request().Context())
	data.Title = ctxi18n.T(c.Request().Context(), "account.page_title")
	data.Breadcrumbs = []utils.Crumb{{Label: ctxi18n.T(c.Request().Context(), "account.subscription")}}
	userID := utils.GetUserID(c)
	if user, err := authstore.GetUserByID(c.Request().Context(), userID); err == nil {
		data.UserID = user.ID
		data.UserEmail = user.Email
		data.CurrentLang = appi18n.NormalizeLocale(user.PreferredLang)
	}

	data.CheckoutQuantity = 1
	if state, err := internalbilling.CurrentAccessState(c.Request().Context(), userID); err == nil {
		data.SubscriptionSlots = state.SubscriptionCount
		data.UsedSlots = state.OwnedGroupCount
		data.RemainingSlots = internalbilling.RemainingGroupSlots(state)
		data.IsLimitExceeded = internalbilling.IsLimitExceeded(state)
		data.HasAvailableGroupSlot = internalbilling.HasAvailableGroupSlot(state)
		if state.OwnedGroupCount > data.CheckoutQuantity {
			data.CheckoutQuantity = state.OwnedGroupCount
		}
	}
	if sub, exists, err := internalbilling.GetUserSubscription(c.Request().Context(), userID); err == nil && exists {
		data.HasActiveSubscription = strings.TrimSpace(sub.ProviderSubscriptionID) != "" &&
			internalbilling.IsSubscriptionActive(sub.Status, sub.GraceUntil, time.Now().UTC())
	}
	if paymentsEnabled, err := flags.IsPaymentEnabled(c.Request().Context()); err == nil {
		data.PaymentsEnabled = paymentsEnabled
	}
	data.ActiveTab = "subscription"
	data.Signals = map[string]any{"formData": map[string]any{"lang": data.CurrentLang}}
	data.IsAuthenticated = true
	data.IsSuperAdmin = utils.IsSuperadmin(c)
	return utils.RenderPage(c, AccountIndex(data))
}

func ManageSubscription(c echo.Context) error {
	ctx := c.Request().Context()
	userID := utils.GetUserID(c)
	paymentsEnabled, paymentsErr := flags.IsPaymentEnabled(ctx)
	if paymentsErr != nil {
		slog.Error("account.billing: failed to read payment flag", "user_id", userID, "err", paymentsErr)
		return c.Redirect(http.StatusFound, "/account")
	}
	if !paymentsEnabled {
		slog.Info("account.billing: payments disabled, blocking manage subscription", "user_id", userID)
		return c.Redirect(http.StatusFound, "/account")
	}
	quantity := 1
	if raw := strings.TrimSpace(c.QueryParam("quantity")); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 {
			quantity = parsed
		}
	}

	sub, exists, err := internalbilling.GetUserSubscription(ctx, userID)
	if err != nil {
		slog.Error("account.billing: failed to load subscription", "user_id", userID, "err", err)
		return c.Redirect(http.StatusFound, "/account")
	}

	if exists && strings.TrimSpace(sub.ProviderSubscriptionID) != "" &&
		internalbilling.IsSubscriptionActive(sub.Status, sub.GraceUntil, time.Now().UTC()) {
		portalURL, portalErr := internalbilling.GetSignedCustomerPortalURL(ctx, userID)
		if portalErr == nil && strings.TrimSpace(portalURL) != "" {
			slog.Info("account.billing: redirecting to signed portal", "user_id", userID, "url", portalURL)
			return c.Redirect(http.StatusFound, portalURL)
		}

		slog.Warn("account.billing: signed portal unavailable, trying sync + retry", "user_id", userID, "err", portalErr)
		if _, syncedExists, syncErr := internalbilling.SyncSubscriptionFromProvider(ctx, userID); syncErr == nil && syncedExists {
			portalURL, retryErr := internalbilling.GetSignedCustomerPortalURL(ctx, userID)
			if retryErr == nil && strings.TrimSpace(portalURL) != "" {
				slog.Info("account.billing: redirecting to signed portal after sync", "user_id", userID, "url", portalURL)
				return c.Redirect(http.StatusFound, portalURL)
			}
		}

		storedPortalURL := strings.TrimSpace(sub.ProviderPortalURL)
		if storedPortalURL != "" && strings.Contains(storedPortalURL, "expires=") && strings.Contains(storedPortalURL, "signature=") {
			slog.Warn("account.billing: using stored signed portal url fallback", "user_id", userID)
			slog.Info("account.billing: redirecting to stored signed portal", "user_id", userID, "url", storedPortalURL)
			return c.Redirect(http.StatusFound, storedPortalURL)
		}

		slog.Warn("account.billing: signed portal unavailable, returning to account", "user_id", userID)
		return c.Redirect(http.StatusFound, "/account")
	}

	checkoutURL := "https://bandcash.lemonsqueezy.com/checkout/buy/d5eda8a8-44ee-46db-812d-cf84370c01ef"
	separator := "?"
	if strings.Contains(checkoutURL, "?") {
		separator = "&"
	}
	redirectURL := checkoutURL + separator + "quantity=" + strconv.Itoa(quantity)
	slog.Info("account.billing: redirecting to checkout", "user_id", userID, "url", redirectURL, "quantity", quantity)
	return c.Redirect(http.StatusFound, redirectURL)
}

func UpdatePaymentMethod(c echo.Context) error {
	ctx := c.Request().Context()
	userID := utils.GetUserID(c)
	paymentsEnabled, paymentsErr := flags.IsPaymentEnabled(ctx)
	if paymentsErr != nil {
		slog.Error("account.billing: failed to read payment flag for update payment", "user_id", userID, "err", paymentsErr)
		return c.Redirect(http.StatusFound, "/account")
	}
	if !paymentsEnabled {
		slog.Info("account.billing: payments disabled, blocking update payment", "user_id", userID)
		return c.Redirect(http.StatusFound, "/account")
	}

	sub, exists, err := internalbilling.GetUserSubscription(ctx, userID)
	if err != nil {
		slog.Error("account.billing: failed to load subscription for update payment", "user_id", userID, "err", err)
		return c.Redirect(http.StatusFound, "/account")
	}

	if !exists || strings.TrimSpace(sub.ProviderSubscriptionID) == "" ||
		!internalbilling.IsSubscriptionActive(sub.Status, sub.GraceUntil, time.Now().UTC()) {
		slog.Warn("account.billing: update payment requested without active subscription", "user_id", userID)
		return c.Redirect(http.StatusFound, "/account")
	}

	updatePaymentURL, updateErr := internalbilling.GetSignedUpdatePaymentMethodURL(ctx, userID)
	if updateErr == nil && strings.TrimSpace(updatePaymentURL) != "" {
		slog.Info("account.billing: redirecting to signed update payment", "user_id", userID, "url", updatePaymentURL)
		return c.Redirect(http.StatusFound, updatePaymentURL)
	}

	slog.Warn("account.billing: signed update payment unavailable, trying sync + retry", "user_id", userID, "err", updateErr)
	if _, syncedExists, syncErr := internalbilling.SyncSubscriptionFromProvider(ctx, userID); syncErr == nil && syncedExists {
		updatePaymentURL, retryErr := internalbilling.GetSignedUpdatePaymentMethodURL(ctx, userID)
		if retryErr == nil && strings.TrimSpace(updatePaymentURL) != "" {
			slog.Info("account.billing: redirecting to signed update payment after sync", "user_id", userID, "url", updatePaymentURL)
			return c.Redirect(http.StatusFound, updatePaymentURL)
		}
	}

	storedUpdateURL := strings.TrimSpace(sub.ProviderUpdatePaymentURL)
	if storedUpdateURL != "" && strings.Contains(storedUpdateURL, "expires=") && strings.Contains(storedUpdateURL, "signature=") {
		slog.Warn("account.billing: using stored signed update payment url fallback", "user_id", userID)
		slog.Info("account.billing: redirecting to stored signed update payment", "user_id", userID, "url", storedUpdateURL)
		return c.Redirect(http.StatusFound, storedUpdateURL)
	}

	slog.Warn("account.billing: signed update payment unavailable, returning to account", "user_id", userID)
	return c.Redirect(http.StatusFound, "/account")
}

func OverLimitPageHandler(c echo.Context) error {
	utils.EnsureTabID(c)
	userID := utils.GetUserID(c)
	state, err := internalbilling.CurrentAccessState(c.Request().Context(), userID)
	if err != nil {
		slog.Error("account.over_limit: failed to load access state", "user_id", userID, "err", err)
		return c.Redirect(http.StatusFound, "/account")
	}
	if !internalbilling.IsLimitExceeded(state) {
		return c.Redirect(http.StatusFound, "/account")
	}

	type groupRow struct {
		Name string `bun:"name"`
	}
	rows := make([]groupRow, 0)
	if scanErr := db.BunDB.NewSelect().
		TableExpr("groups").
		Column("name").
		Where("admin_user_id = ?", userID).
		OrderExpr("created_at DESC").
		Scan(c.Request().Context(), &rows); scanErr != nil && !errors.Is(scanErr, sql.ErrNoRows) {
		slog.Error("account.over_limit: failed to list owned groups", "user_id", userID, "err", scanErr)
	}

	groups := make([]OverLimitGroup, 0, len(rows))
	for _, row := range rows {
		groups = append(groups, OverLimitGroup{Name: strings.TrimSpace(row.Name)})
	}

	data := OverLimitData{
		Title:                 ctxi18n.T(c.Request().Context(), "account.over_limit_page_title"),
		Breadcrumbs:           []utils.Crumb{{Label: ctxi18n.T(c.Request().Context(), "account.over_limit_title")}},
		SubscriptionCap:       state.SubscriptionCount,
		OwnedGroups:           state.OwnedGroupCount,
		ExcessGroups:          state.OwnedGroupCount - state.SubscriptionCount,
		Groups:                groups,
		PaymentsEnabled:       false,
		HasActiveSubscription: false,
		CheckoutQuantity:      1,
		IsAuthenticated:       true,
		IsSuperAdmin:          utils.IsSuperadmin(c),
	}
	if state.OwnedGroupCount > data.CheckoutQuantity {
		data.CheckoutQuantity = state.OwnedGroupCount
	}
	if paymentsEnabled, err := flags.IsPaymentEnabled(c.Request().Context()); err == nil {
		data.PaymentsEnabled = paymentsEnabled
	}
	if sub, exists, err := internalbilling.GetUserSubscription(c.Request().Context(), userID); err == nil && exists {
		data.HasActiveSubscription = strings.TrimSpace(sub.ProviderSubscriptionID) != "" &&
			internalbilling.IsSubscriptionActive(sub.Status, sub.GraceUntil, time.Now().UTC())
	}

	return utils.RenderPage(c, OverLimitPage(data))
}
func LanguagePageHandler(c echo.Context) error {
	utils.EnsureTabID(c)
	data := Data(c.Request().Context())
	data.Title = ctxi18n.T(c.Request().Context(), "account.page_title")
	data.Breadcrumbs = []utils.Crumb{{Label: ctxi18n.T(c.Request().Context(), "account.language")}}
	userID := utils.GetUserID(c)
	if user, err := authstore.GetUserByID(c.Request().Context(), userID); err == nil {
		data.UserEmail = user.Email
		data.CurrentLang = appi18n.NormalizeLocale(user.PreferredLang)
	}
	data.ActiveTab = "language"
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
		Title:            ctxi18n.T(c.Request().Context(), "account.page_title"),
		Breadcrumbs:      []utils.Crumb{{Label: ctxi18n.T(c.Request().Context(), "account.sessions")}},
		CurrentSessionID: "",
		UserEmail:        "",
		Sessions:         sessions,
		ActiveTab:        "sessions",
	}

	if user, err := authstore.GetUserByID(c.Request().Context(), userID); err == nil {
		data.UserEmail = user.Email
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
