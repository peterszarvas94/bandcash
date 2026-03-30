package auth

import (
	"database/sql"
	"errors"
	"log/slog"
	"net/http"
	"strings"
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

type authSignals struct {
	TabID    string `json:"tab_id"`
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

func authSessionUser(c echo.Context) (bool, string) {
	cookie, err := c.Cookie(utils.SessionCookieName)
	if err != nil || cookie.Value == "" {
		return false, ""
	}

	session, err := db.Qry.GetUserSessionByToken(c.Request().Context(), cookie.Value)
	if err != nil {
		return false, ""
	}

	user, err := db.Qry.GetUserByID(c.Request().Context(), session.UserID)
	if err != nil {
		return true, ""
	}

	syncPreferredLangFromQuery(c, user.ID, user.PreferredLang)
	return true, user.Email
}

func syncPreferredLangFromQuery(c echo.Context, userID string, currentPreferredLang string) {
	rawLang := strings.TrimSpace(c.QueryParam("lang"))
	if rawLang == "" {
		return
	}

	lang := appi18n.NormalizeLocale(rawLang)
	if appi18n.NormalizeLocale(currentPreferredLang) == lang {
		return
	}

	_ = db.Qry.UpdateUserPreferredLang(c.Request().Context(), db.UpdateUserPreferredLangParams{
		ID:            userID,
		PreferredLang: lang,
	})
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
	utils.EnsureTabID(c)

	ctx := c.Request().Context()
	isAuthenticated, userEmail := authSessionUser(c)
	data := AuthPageData{
		Title:           ctxi18n.T(ctx, "auth.sign_in") + " - Bandcash",
		Breadcrumbs:     []utils.Crumb{{Label: ctxi18n.T(ctx, "auth.sign_in")}},
		CurrentLang:     appi18n.LocaleCode(ctx),
		IsAuthenticated: isAuthenticated,
		UserEmail:       userEmail,
	}

	return utils.RenderPage(c, LoginPage(data))
}

// LoginRequest handles login form submission (sends magic link)
func (a *Auth) LoginRequest(c echo.Context) error {
	signals := authSignals{}
	if err := datastar.ReadSignals(c.Request(), &signals); err != nil {
		return c.NoContent(http.StatusBadRequest)
	}
	if !utils.SetTabID(c, signals.TabID) {
		return c.NoContent(http.StatusBadRequest)
	}
	signals.FormData.Email = strings.ToLower(strings.TrimSpace(signals.FormData.Email))
	if errs := utils.ValidateWithLocale(c.Request().Context(), signals.FormData); errs != nil {
		_ = utils.SSEHub.PatchSignals(c, map[string]any{"authError": errs["email"], "authServerError": ""})
		return c.NoContent(http.StatusUnprocessableEntity)
	}
	emailAddress := signals.FormData.Email

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
	return c.Redirect(http.StatusFound, "/login")
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

	// Invite links do not expire.
	if magicLink.Action != "invite" && time.Now().After(magicLink.ExpiresAt) {
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
		utils.Notify(c, "error", ctxi18n.T(c.Request().Context(), "auth.banned"))
		return c.Redirect(http.StatusFound, "/login")
	}

	// Create session
	expiresAt := time.Now().Add(30 * 24 * time.Hour)
	session, err := db.Qry.CreateUserSession(c.Request().Context(), db.CreateUserSessionParams{
		ID:        utils.GenerateID("ses"),
		UserID:    user.ID,
		Token:     utils.GenerateID("tok"),
		ExpiresAt: expiresAt,
	})
	if err != nil {
		slog.Error("auth.verify: failed to create session", "user_id", user.ID, "err", err)
		return a.renderVerifyLinkError(c, http.StatusInternalServerError)
	}
	utils.SetSessionCookie(c, session.Token)

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

		return c.Redirect(http.StatusFound, "/groups/"+groupID+"/overview")
	}

	// Redirect to groups page
	return c.Redirect(http.StatusFound, "/groups")
}

// Logout clears the session
func (a *Auth) Logout(c echo.Context) error {
	type logoutSignals struct {
		TabID string `json:"tab_id"`
	}

	signals := logoutSignals{}
	if err := datastar.ReadSignals(c.Request(), &signals); err != nil {
		return c.NoContent(http.StatusBadRequest)
	}
	if !utils.SetTabID(c, signals.TabID) {
		return c.NoContent(http.StatusBadRequest)
	}

	// Delete session from DB
	cookie, err := c.Cookie(utils.SessionCookieName)
	if err == nil {
		session, err := db.Qry.GetUserSessionByToken(c.Request().Context(), cookie.Value)
		if err == nil {
			userID := middleware.GetUserID(c)
			if userID != "" {
				_ = db.Qry.DeleteUserSession(c.Request().Context(), db.DeleteUserSessionParams{
					ID:     session.ID,
					UserID: userID,
				})
			}
		}
	}

	utils.ClearSessionCookie(c)

	err = utils.SSEHub.Redirect(c, "/")
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
			return c.Redirect(http.StatusFound, "/groups/"+adminGroups[0].ID+"/overview")
		}
		return c.Redirect(http.StatusFound, "/groups/"+filteredReaders[0].ID+"/overview")
	}

	return c.Redirect(http.StatusFound, "/groups")
}
