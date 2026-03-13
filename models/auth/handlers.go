package auth

import (
	"database/sql"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	ctxi18nlib "github.com/invopop/ctxi18n"
	ctxi18n "github.com/invopop/ctxi18n/i18n"
	"github.com/labstack/echo/v4"
	"github.com/starfederation/datastar-go/datastar"

	"bandcash/internal/db"
	"bandcash/internal/email"
	appi18n "bandcash/internal/i18n"
	"bandcash/internal/middleware"
	"bandcash/internal/utils"
	shared "bandcash/models/shared"
	icons "bandcash/models/shared/icons"
)

type Auth struct {
}

const resendCooldown = 30 * time.Second

const (
	loginEmailCooldown     = 60 * time.Second
	loginIPEmailWindow     = 15 * time.Minute
	loginEmailGlobalWindow = 1 * time.Hour
	loginIPEmailLimit      = 5
	loginEmailGlobalLimit  = 10
)

type loginThrottle struct {
	mu            sync.Mutex
	emailLastSent map[string]time.Time
	ipEmailHits   map[string][]time.Time
	emailHits     map[string][]time.Time
	ops           int
}

func newLoginThrottle() *loginThrottle {
	return &loginThrottle{
		emailLastSent: make(map[string]time.Time),
		ipEmailHits:   make(map[string][]time.Time),
		emailHits:     make(map[string][]time.Time),
	}
}

func (t *loginThrottle) allow(ip string, emailAddress string, now time.Time) bool {
	if ip == "" || emailAddress == "" {
		return true
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	t.ops++
	if t.ops%100 == 0 {
		t.cleanup(now)
	}

	if lastSentAt, ok := t.emailLastSent[emailAddress]; ok && now.Sub(lastSentAt) < loginEmailCooldown {
		return false
	}

	ipEmailKey := ip + "|" + emailAddress
	ipHits := pruneHits(t.ipEmailHits[ipEmailKey], now, loginIPEmailWindow)
	if len(ipHits) >= loginIPEmailLimit {
		t.ipEmailHits[ipEmailKey] = ipHits
		return false
	}

	emailHits := pruneHits(t.emailHits[emailAddress], now, loginEmailGlobalWindow)
	if len(emailHits) >= loginEmailGlobalLimit {
		t.emailHits[emailAddress] = emailHits
		return false
	}

	ipHits = append(ipHits, now)
	emailHits = append(emailHits, now)
	t.ipEmailHits[ipEmailKey] = ipHits
	t.emailHits[emailAddress] = emailHits
	t.emailLastSent[emailAddress] = now

	return true
}

func (t *loginThrottle) cleanup(now time.Time) {
	for key, lastSentAt := range t.emailLastSent {
		if now.Sub(lastSentAt) >= loginEmailGlobalWindow {
			delete(t.emailLastSent, key)
		}
	}

	for key, hits := range t.ipEmailHits {
		pruned := pruneHits(hits, now, loginIPEmailWindow)
		if len(pruned) == 0 {
			delete(t.ipEmailHits, key)
			continue
		}
		t.ipEmailHits[key] = pruned
	}

	for key, hits := range t.emailHits {
		pruned := pruneHits(hits, now, loginEmailGlobalWindow)
		if len(pruned) == 0 {
			delete(t.emailHits, key)
			continue
		}
		t.emailHits[key] = pruned
	}
}

func pruneHits(hits []time.Time, now time.Time, window time.Duration) []time.Time {
	if len(hits) == 0 {
		return hits
	}

	cutoff := now.Add(-window)
	idx := 0
	for idx < len(hits) && hits[idx].Before(cutoff) {
		idx++
	}

	if idx == 0 {
		return hits
	}
	if idx >= len(hits) {
		return nil
	}

	return hits[idx:]
}

func requestIP(c echo.Context) string {
	ip := strings.TrimSpace(c.RealIP())
	if ip != "" {
		return ip
	}

	remoteAddr := strings.TrimSpace(c.Request().RemoteAddr)
	if host, _, err := net.SplitHostPort(remoteAddr); err == nil {
		return host
	}

	return remoteAddr
}

var authLoginThrottle = newLoginThrottle()

type authSignals struct {
	FormData struct {
		Email string `json:"email" validate:"required,email,max=320"`
	} `json:"formData"`
}

func New() *Auth {
	return &Auth{}
}

func maskEmail(email string) string {
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return email
	}

	local := parts[0]
	domain := parts[1]

	maskedLocal := "***"
	if len(local) == 1 {
		maskedLocal = local + "***"
	} else if len(local) > 1 {
		maskedLocal = local[:1] + "***" + local[len(local)-1:]
	}

	maskedDomain := "***"
	if dot := strings.LastIndex(domain, "."); dot > 1 {
		maskedDomain = domain[:1] + "***" + domain[dot:]
	} else if len(domain) > 0 {
		maskedDomain = domain[:1] + "***"
	}

	return maskedLocal + "@" + maskedDomain
}

func (a *Auth) patchLoginSentState(c echo.Context, emailAddress string) {
	_ = utils.SSEHub.PatchSignals(c, map[string]any{
		"authError":            "",
		"authServerError":      "",
		"authState":            "sent",
		"submittedEmail":       emailAddress,
		"submittedEmailMasked": maskEmail(emailAddress),
		"resendRemaining":      int(resendCooldown.Seconds()),
	})
}

func (a *Auth) renderVerifyLinkError(c echo.Context, status int) error {
	ctx := c.Request().Context()
	return utils.RenderPage(c, shared.ErrorPage(shared.ErrorPageData{
		Title:      ctxi18n.T(ctx, "error_pages.link.invalid_title"),
		StatusCode: status,
		IconName:   icons.IconLink2Off,
		Heading:    ctxi18n.T(ctx, "error_pages.link.invalid_title"),
		Message:    ctxi18n.T(ctx, "error_pages.link.invalid_body"),
		HomeLabel:  ctxi18n.T(ctx, "error_pages.home_action"),
		HomeHref:   appi18n.LocalizedHomePath(ctx),
	}))
}

// LoginPage shows the login form
func (a *Auth) LoginPage(c echo.Context) error {
	return c.Redirect(http.StatusFound, "/")
}

// LoginRequest handles login form submission (sends magic link)
func (a *Auth) LoginRequest(c echo.Context) error {
	signals := authSignals{}
	if err := datastar.ReadSignals(c.Request(), &signals); err != nil {
		return c.NoContent(http.StatusBadRequest)
	}
	signals.FormData.Email = strings.ToLower(strings.TrimSpace(signals.FormData.Email))
	if errs := utils.ValidateWithLocale(c.Request().Context(), signals.FormData); errs != nil {
		_ = utils.SSEHub.PatchSignals(c, map[string]any{"authError": errs["email"], "authServerError": ""})
		return c.NoContent(http.StatusUnprocessableEntity)
	}
	emailAddress := signals.FormData.Email
	clientIP := requestIP(c)
	if !authLoginThrottle.allow(clientIP, emailAddress, time.Now()) {
		slog.Warn("auth.login: throttled login request", "ip", clientIP, "email", maskEmail(emailAddress))
		a.patchLoginSentState(c, emailAddress)
		return c.NoContent(http.StatusOK)
	}

	loginUser, err := db.Qry.GetUserByEmail(c.Request().Context(), emailAddress)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			slog.Error("auth.login: failed to load user by email", "err", err)
			_ = utils.SSEHub.PatchSignals(c, map[string]any{
				"authError":       "",
				"authServerError": ctxi18n.T(c.Request().Context(), "auth.generic_server_error"),
			})
			return c.NoContent(http.StatusInternalServerError)
		}

		signupEnabled, err := utils.IsSignupEnabled(c.Request().Context())
		if err != nil {
			slog.Error("auth.login: failed to read signup flag", "err", err)
			a.patchLoginSentState(c, emailAddress)
			return c.NoContent(http.StatusOK)
		}

		if !signupEnabled {
			a.patchLoginSentState(c, emailAddress)
			return c.NoContent(http.StatusOK)
		}

		loginUser, err = db.Qry.CreateUser(c.Request().Context(), db.CreateUserParams{
			ID:            utils.GenerateID("usr"),
			Email:         emailAddress,
			PreferredLang: appi18n.LocaleCode(c.Request().Context()),
		})
		if err != nil {
			slog.Error("auth.login: failed to create user", "err", err)
			a.patchLoginSentState(c, emailAddress)
			return c.NoContent(http.StatusOK)
		}
	}

	bannedCount, err := db.Qry.IsUserBanned(c.Request().Context(), loginUser.ID)
	if err != nil {
		slog.Error("auth.login: failed to check user ban", "err", err)
		a.patchLoginSentState(c, emailAddress)
		return c.NoContent(http.StatusOK)
	}
	if bannedCount > 0 {
		a.patchLoginSentState(c, emailAddress)
		return c.NoContent(http.StatusOK)
	}

	token := utils.GenerateID("tok")
	expiresAt := time.Now().Add(1 * time.Hour)

	_, err = db.Qry.CreateMagicLink(c.Request().Context(), db.CreateMagicLinkParams{
		ID:        utils.GenerateID("mag"),
		Token:     token,
		Email:     emailAddress,
		Action:    "login",
		ExpiresAt: expiresAt,
	})
	if err != nil {
		slog.Error("auth.login: failed to create magic link", "err", err)
		a.patchLoginSentState(c, emailAddress)
		return c.NoContent(http.StatusOK)
	}

	mailCtx := c.Request().Context()
	if loginUser.PreferredLang != "" {
		if localizedCtx, localeErr := ctxi18nlib.WithLocale(mailCtx, loginUser.PreferredLang); localeErr == nil {
			mailCtx = localizedCtx
		}
	}
	err = email.Email().SendMagicLink(mailCtx, emailAddress, token, utils.Env().URL)
	if err != nil {
		slog.Error("auth.login: failed to send email", "err", err)
	}

	a.patchLoginSentState(c, emailAddress)
	return c.NoContent(http.StatusOK)
}

// LoginSentPage shows confirmation that email was sent
func (a *Auth) LoginSentPage(c echo.Context) error {
	return c.Redirect(http.StatusFound, "/")
}

// VerifyMagicLink handles the magic link verification
func (a *Auth) VerifyMagicLink(c echo.Context) error {
	token := c.QueryParam("token")
	if !utils.IsValidID(token, "tok") {
		return a.renderVerifyLinkError(c, http.StatusBadRequest)
	}

	locale := appi18n.NormalizeLocale(c.QueryParam("lang"))

	// Get magic link
	magicLink, err := db.Qry.GetMagicLinkByToken(c.Request().Context(), token)
	if err != nil {
		return a.renderVerifyLinkError(c, http.StatusBadRequest)
	}

	// Check if already used
	if magicLink.UsedAt.Valid {
		return a.renderVerifyLinkError(c, http.StatusBadRequest)
	}

	// Check if expired
	if time.Now().After(magicLink.ExpiresAt) {
		return a.renderVerifyLinkError(c, http.StatusBadRequest)
	}

	// Mark as used
	err = db.Qry.UseMagicLink(c.Request().Context(), magicLink.ID)
	if err != nil {
		slog.Error("auth: failed to mark magic link used", "err", err)
		return a.renderVerifyLinkError(c, http.StatusBadRequest)
	}

	// Get user
	user, err := db.Qry.GetUserByEmail(c.Request().Context(), magicLink.Email)
	if err != nil {
		// Create user on invite accept
		user, err = db.Qry.CreateUser(c.Request().Context(), db.CreateUserParams{
			ID:            utils.GenerateID("usr"),
			Email:         magicLink.Email,
			PreferredLang: locale,
		})
		if err != nil {
			slog.Error("auth: failed to create user", "err", err)
			return a.renderVerifyLinkError(c, http.StatusBadRequest)
		}
	} else if user.PreferredLang != locale {
		if err := db.Qry.UpdateUserPreferredLang(c.Request().Context(), db.UpdateUserPreferredLangParams{PreferredLang: locale, ID: user.ID}); err != nil {
			slog.Warn("auth.verify: failed to update user preferred language", "user_id", user.ID, "err", err)
		} else {
			user.PreferredLang = locale
		}
	}

	bannedCount, err := db.Qry.IsUserBanned(c.Request().Context(), user.ID)
	if err != nil {
		slog.Error("auth.verify: failed to check user ban", "user_id", user.ID, "err", err)
		return a.renderVerifyLinkError(c, http.StatusBadRequest)
	}
	if bannedCount > 0 {
		utils.Notify(c, "warning", ctxi18n.T(c.Request().Context(), "auth.banned"))
		return c.Redirect(http.StatusFound, "/auth/login")
	}

	// Create session cookie
	env := utils.Env()
	c.SetCookie(&http.Cookie{
		Name:     "session",
		Value:    user.ID,
		Path:     "/",
		MaxAge:   86400 * 30, // 30 days
		HttpOnly: true,
		Secure:   env.AppEnv == "production",
		SameSite: http.SameSiteLaxMode,
	})

	// If invite, add viewer access
	if magicLink.Action == "invite" {
		if !magicLink.GroupID.Valid {
			return a.renderVerifyLinkError(c, http.StatusBadRequest)
		}

		groupID := magicLink.GroupID.String
		inviteRole := strings.TrimSpace(strings.ToLower(magicLink.InviteRole))
		if inviteRole != "admin" {
			inviteRole = "viewer"
		}
		groupName := groupID
		group, groupErr := db.Qry.GetGroupByID(c.Request().Context(), groupID)
		if groupErr == nil {
			groupName = group.Name
			if group.AdminUserID != user.ID {
				if inviteRole == "admin" {
					_ = db.Qry.RemoveGroupReader(c.Request().Context(), db.RemoveGroupReaderParams{UserID: user.ID, GroupID: groupID})
					_, err = db.Qry.CreateGroupAdmin(c.Request().Context(), db.CreateGroupAdminParams{
						ID:      utils.GenerateID("gad"),
						UserID:  user.ID,
						GroupID: groupID,
					})
					if err != nil {
						slog.Warn("auth: failed to add group admin", "group_id", groupID, "user_id", user.ID, "err", err)
					}
				} else {
					_ = db.Qry.RemoveGroupAdmin(c.Request().Context(), db.RemoveGroupAdminParams{UserID: user.ID, GroupID: groupID})
					_, err = db.Qry.CreateGroupReader(c.Request().Context(), db.CreateGroupReaderParams{
						ID:      utils.GenerateID("grd"),
						UserID:  user.ID,
						GroupID: groupID,
					})
					if err != nil {
						slog.Warn("auth: failed to add group reader", "group_id", groupID, "user_id", user.ID, "err", err)
					}
				}
			}
		} else {
			slog.Warn("auth.verify: failed to load group for invite", "group_id", groupID, "err", groupErr)
			if inviteRole == "admin" {
				_ = db.Qry.RemoveGroupReader(c.Request().Context(), db.RemoveGroupReaderParams{UserID: user.ID, GroupID: groupID})
				_, err = db.Qry.CreateGroupAdmin(c.Request().Context(), db.CreateGroupAdminParams{
					ID:      utils.GenerateID("gad"),
					UserID:  user.ID,
					GroupID: groupID,
				})
				if err != nil {
					slog.Warn("auth: failed to add group admin", "group_id", groupID, "user_id", user.ID, "err", err)
				}
			} else {
				_ = db.Qry.RemoveGroupAdmin(c.Request().Context(), db.RemoveGroupAdminParams{UserID: user.ID, GroupID: groupID})
				_, err = db.Qry.CreateGroupReader(c.Request().Context(), db.CreateGroupReaderParams{
					ID:      utils.GenerateID("grd"),
					UserID:  user.ID,
					GroupID: groupID,
				})
				if err != nil {
					slog.Warn("auth: failed to add group reader", "group_id", groupID, "user_id", user.ID, "err", err)
				}
			}
		}

		notifyCtx := c.Request().Context()
		if user.PreferredLang != "" {
			if localizedCtx, localeErr := ctxi18nlib.WithLocale(notifyCtx, user.PreferredLang); localeErr == nil {
				notifyCtx = localizedCtx
			}
		}
		err = email.Email().SendInviteAccepted(notifyCtx, user.Email, groupName, groupID, utils.Env().URL)
		if err != nil {
			slog.Warn("auth.verify: failed to send invite accepted email", "group_id", groupID, "user_id", user.ID, "err", err)
		}

		return c.Redirect(http.StatusFound, "/groups/"+groupID+"/events")
	}

	// Redirect to group dashboard
	return c.Redirect(http.StatusFound, "/dashboard")
}

// Logout clears the session
func (a *Auth) Logout(c echo.Context) error {
	c.SetCookie(&http.Cookie{
		Name:     "session",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})

	err := utils.SSEHub.Redirect(c, "/")
	if err != nil {
		return c.Redirect(http.StatusFound, "/")
	}

	return c.NoContent(http.StatusOK)
}

// Dashboard shows user's groups or create group page
func (a *Auth) Dashboard(c echo.Context) error {
	userID := middleware.GetUserID(c)

	adminGroups, err := db.Qry.ListGroupsByAdmin(c.Request().Context(), db.ListGroupsByAdminParams{OwnerUserID: userID, UserID: userID})
	if err != nil {
		slog.Error("auth: failed to load admin groups", "err", err)
		return c.Redirect(http.StatusFound, "/groups/new")
	}

	readerGroups, err := db.Qry.ListGroupsByReader(c.Request().Context(), db.ListGroupsByReaderParams{UserID: userID, AdminUserID: userID})
	if err != nil {
		slog.Error("auth: failed to load reader groups", "err", err)
		return c.Redirect(http.StatusFound, "/groups/new")
	}

	// Dedupe reader groups where user is admin
	adminMap := make(map[string]bool, len(adminGroups))
	for _, group := range adminGroups {
		adminMap[group.ID] = true
	}
	filteredReaders := make([]db.Group, 0, len(readerGroups))
	for _, group := range readerGroups {
		if adminMap[group.ID] {
			continue
		}
		filteredReaders = append(filteredReaders, group)
	}

	if len(adminGroups)+len(filteredReaders) == 0 {
		return c.Redirect(http.StatusFound, "/groups/new")
	}

	if len(adminGroups)+len(filteredReaders) == 1 {
		if len(adminGroups) == 1 {
			return c.Redirect(http.StatusFound, "/groups/"+adminGroups[0].ID+"/events")
		}
		return c.Redirect(http.StatusFound, "/groups/"+filteredReaders[0].ID+"/events")
	}

	return c.Redirect(http.StatusFound, "/dashboard")
}
