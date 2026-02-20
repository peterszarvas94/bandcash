package auth

import (
	"log/slog"
	"net/http"
	"net/url"
	"time"

	ctxi18n "github.com/invopop/ctxi18n/i18n"
	"github.com/labstack/echo/v4"
	"github.com/starfederation/datastar-go/datastar"

	"bandcash/internal/db"
	"bandcash/internal/email"
	"bandcash/internal/middleware"
	"bandcash/internal/utils"
)

type Auth struct {
}

type authSignals struct {
	FormData struct {
		Email string `json:"email"`
	} `json:"formData"`
}

func New() *Auth {
	return &Auth{}
}

// LoginPage shows the login form
func (a *Auth) LoginPage(c echo.Context) error {
	utils.EnsureClientID(c)
	data := AuthPageData{
		Title:       ctxi18n.T(c.Request().Context(), "auth.login_title"),
		Breadcrumbs: []utils.Crumb{{Label: ctxi18n.T(c.Request().Context(), "auth.login")}},
	}
	return utils.RenderComponent(c, LoginPage(data))
}

// LoginRequest handles login form submission (sends magic link)
func (a *Auth) LoginRequest(c echo.Context) error {
	signals := authSignals{}
	if err := datastar.ReadSignals(c.Request(), &signals); err != nil {
		return c.NoContent(http.StatusBadRequest)
	}
	emailAdress := signals.FormData.Email
	if emailAdress == "" {
		return c.NoContent(http.StatusBadRequest)
	}

	// Check if user exists
	_, err := db.Qry.GetUserByEmail(c.Request().Context(), emailAdress)
	if err != nil {
		// User doesn't exist - offer to sign up instead
		err = utils.SSEHub.Redirect(c, "/auth/signup?email="+url.QueryEscape(emailAdress))
		if err != nil {
			return c.NoContent(http.StatusInternalServerError)
		}
		return c.NoContent(http.StatusOK)
	}

	// Create magic link
	token := utils.GenerateID("tok")
	expiresAt := time.Now().Add(1 * time.Hour)

	_, err = db.Qry.CreateMagicLink(c.Request().Context(), db.CreateMagicLinkParams{
		ID:        utils.GenerateID("mag"),
		Token:     token,
		Email:     emailAdress,
		Action:    "login",
		ExpiresAt: expiresAt,
	})
	if err != nil {
		slog.Error("auth: failed to create magic link", "err", err)
		utils.Notify(c, "error", ctxi18n.T(c.Request().Context(), "auth.notifications.create_link_failed"))
		return c.NoContent(http.StatusInternalServerError)
	}

	// Send email
	err = email.Email().SendMagicLink(c.Request().Context(), emailAdress, token, utils.Env().URL)
	if err != nil {
		slog.Error("auth: failed to send email", "err", err)
		utils.Notify(c, "error", ctxi18n.T(c.Request().Context(), "auth.notifications.send_email_failed"))
		return c.NoContent(http.StatusInternalServerError)
	}
	utils.Notify(c, "info", ctxi18n.T(c.Request().Context(), "auth.notifications.link_sent"))

	err = utils.SSEHub.Redirect(c, "/auth/login-sent")
	if err != nil {
		return c.NoContent(http.StatusInternalServerError)
	}
	return c.NoContent(http.StatusOK)
}

// SignupPage shows the signup form
func (a *Auth) SignupPage(c echo.Context) error {
	utils.EnsureClientID(c)
	emailAdress := c.QueryParam("email")
	data := AuthPageData{
		Title:       ctxi18n.T(c.Request().Context(), "auth.signup_title"),
		Breadcrumbs: []utils.Crumb{{Label: ctxi18n.T(c.Request().Context(), "auth.signup")}},
		Email:       emailAdress,
	}
	return utils.RenderComponent(c, SignupPage(data))
}

// SignupRequest handles signup form submission
func (a *Auth) SignupRequest(c echo.Context) error {
	signals := authSignals{}
	if err := datastar.ReadSignals(c.Request(), &signals); err != nil {
		return c.NoContent(http.StatusBadRequest)
	}
	emailAdress := signals.FormData.Email
	if emailAdress == "" {
		return c.NoContent(http.StatusBadRequest)
	}

	// Check if user already exists
	_, err := db.Qry.GetUserByEmail(c.Request().Context(), emailAdress)
	if err == nil {
		err = utils.SSEHub.PatchSignals(c, map[string]any{
			"authError": ctxi18n.T(c.Request().Context(), "auth.email_in_use"),
		})
		if err != nil {
			return c.NoContent(http.StatusInternalServerError)
		}
		return c.NoContent(http.StatusOK)
	}

	// Create user
	user, err := db.Qry.CreateUser(c.Request().Context(), db.CreateUserParams{
		ID:    utils.GenerateID("usr"),
		Email: emailAdress,
	})
	if err != nil {
		slog.Error("auth: failed to create user", "err", err)
		utils.Notify(c, "error", ctxi18n.T(c.Request().Context(), "auth.notifications.create_user_failed"))
		return c.NoContent(http.StatusInternalServerError)
	}

	// Create magic link for login
	token := utils.GenerateID("tok")
	expiresAt := time.Now().Add(1 * time.Hour)

	_, err = db.Qry.CreateMagicLink(c.Request().Context(), db.CreateMagicLinkParams{
		ID:        utils.GenerateID("mag"),
		Token:     token,
		Email:     emailAdress,
		Action:    "login",
		ExpiresAt: expiresAt,
	})
	if err != nil {
		slog.Error("auth: failed to create magic link", "err", err)
		utils.Notify(c, "error", ctxi18n.T(c.Request().Context(), "auth.notifications.create_link_failed"))
		return c.NoContent(http.StatusInternalServerError)
	}

	slog.Info("auth: created user and magic link", "user_id", user.ID, "email", emailAdress)

	// Send welcome/login email
	err = email.Email().SendMagicLink(c.Request().Context(), emailAdress, token, utils.Env().URL)
	if err != nil {
		slog.Error("auth: failed to send email", "err", err)
		utils.Notify(c, "error", ctxi18n.T(c.Request().Context(), "auth.notifications.send_email_failed"))
		return c.NoContent(http.StatusInternalServerError)
	}
	utils.Notify(c, "success", ctxi18n.T(c.Request().Context(), "auth.notifications.account_created"))

	err = utils.SSEHub.Redirect(c, "/auth/login-sent")
	if err != nil {
		return c.NoContent(http.StatusInternalServerError)
	}
	return c.NoContent(http.StatusOK)
}

// LoginSentPage shows confirmation that email was sent
func (a *Auth) LoginSentPage(c echo.Context) error {
	data := AuthPageData{
		Title:       ctxi18n.T(c.Request().Context(), "auth.check_email_title"),
		Breadcrumbs: []utils.Crumb{{Label: ctxi18n.T(c.Request().Context(), "auth.check_email")}},
	}
	return utils.RenderComponent(c, LoginSentPage(data))
}

// VerifyMagicLink handles the magic link verification
func (a *Auth) VerifyMagicLink(c echo.Context) error {
	token := c.QueryParam("token")
	if token == "" {
		return c.String(http.StatusBadRequest, "Invalid token")
	}

	// Get magic link
	magicLink, err := db.Qry.GetMagicLinkByToken(c.Request().Context(), token)
	if err != nil {
		return c.String(http.StatusNotFound, "Invalid or expired link")
	}

	// Check if already used
	if magicLink.UsedAt.Valid {
		return c.String(http.StatusBadRequest, "Link already used")
	}

	// Check if expired
	if time.Now().After(magicLink.ExpiresAt) {
		return c.String(http.StatusBadRequest, "Link expired")
	}

	// Mark as used
	err = db.Qry.UseMagicLink(c.Request().Context(), magicLink.ID)
	if err != nil {
		slog.Error("auth: failed to mark magic link used", "err", err)
		return c.String(http.StatusInternalServerError, "Failed to process link")
	}

	// Get user
	user, err := db.Qry.GetUserByEmail(c.Request().Context(), magicLink.Email)
	if err != nil {
		// Create user on invite accept
		user, err = db.Qry.CreateUser(c.Request().Context(), db.CreateUserParams{
			ID:    utils.GenerateID("usr"),
			Email: magicLink.Email,
		})
		if err != nil {
			slog.Error("auth: failed to create user", "err", err)
			return c.String(http.StatusInternalServerError, "Failed to create user")
		}
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
			return c.String(http.StatusBadRequest, "Invalid invitation")
		}

		groupID := magicLink.GroupID.String
		group, err := db.Qry.GetGroupByID(c.Request().Context(), groupID)
		if err == nil && group.AdminUserID == user.ID {
			return c.Redirect(http.StatusFound, "/groups/"+groupID)
		}

		_, err = db.Qry.CreateGroupReader(c.Request().Context(), db.CreateGroupReaderParams{
			ID:      utils.GenerateID("grd"),
			UserID:  user.ID,
			GroupID: groupID,
		})
		if err != nil {
			slog.Warn("auth: failed to add group reader", "err", err)
		}

		return c.Redirect(http.StatusFound, "/groups/"+groupID)
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

	adminGroups, err := db.Qry.ListGroupsByAdmin(c.Request().Context(), userID)
	if err != nil {
		slog.Error("auth: failed to load admin groups", "err", err)
		return c.Redirect(http.StatusFound, "/groups/new")
	}

	readerGroups, err := db.Qry.ListGroupsByReader(c.Request().Context(), userID)
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
			return c.Redirect(http.StatusFound, "/groups/"+adminGroups[0].ID)
		}
		return c.Redirect(http.StatusFound, "/groups/"+filteredReaders[0].ID)
	}

	return c.Redirect(http.StatusFound, "/dashboard")
}
