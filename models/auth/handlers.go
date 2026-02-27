package auth

import (
	"database/sql"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	ctxi18n "github.com/invopop/ctxi18n/i18n"
	"github.com/labstack/echo/v4"
	"github.com/starfederation/datastar-go/datastar"

	"bandcash/internal/db"
	"bandcash/internal/email"
	appi18n "bandcash/internal/i18n"
	"bandcash/internal/middleware"
	"bandcash/internal/utils"
)

type Auth struct {
}

const resendCooldown = 30 * time.Second

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
	signals.FormData.Email = utils.NormalizeEmail(signals.FormData.Email)
	if errs := utils.ValidateWithLocale(c.Request().Context(), signals.FormData); errs != nil {
		_ = utils.SSEHub.PatchSignals(c, map[string]any{"authError": errs["email"], "authServerError": ""})
		return c.NoContent(http.StatusUnprocessableEntity)
	}
	emailAddress := signals.FormData.Email

	_, err := db.Qry.GetUserByEmail(c.Request().Context(), emailAddress)
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

		_, err = db.Qry.CreateUser(c.Request().Context(), db.CreateUserParams{
			ID:    utils.GenerateID("usr"),
			Email: emailAddress,
		})
		if err != nil {
			slog.Error("auth.login: failed to create user", "err", err)
			a.patchLoginSentState(c, emailAddress)
			return c.NoContent(http.StatusOK)
		}
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

	err = email.Email().SendMagicLink(c.Request().Context(), emailAddress, token, utils.Env().URL)
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
		return c.String(http.StatusBadRequest, "Invalid token")
	}

	locale := appi18n.NormalizeLocale(c.QueryParam("lang"))
	c.SetCookie(&http.Cookie{
		Name:     appi18n.CookieName,
		Value:    locale,
		Path:     "/",
		MaxAge:   60 * 60 * 24 * 365,
		SameSite: http.SameSiteLaxMode,
		HttpOnly: true,
	})

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
			return c.Redirect(http.StatusFound, "/groups/"+groupID+"/events")
		}

		_, err = db.Qry.CreateGroupReader(c.Request().Context(), db.CreateGroupReaderParams{
			ID:      utils.GenerateID("grd"),
			UserID:  user.ID,
			GroupID: groupID,
		})
		if err != nil {
			slog.Warn("auth: failed to add group reader", "err", err)
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
			return c.Redirect(http.StatusFound, "/groups/"+adminGroups[0].ID+"/events")
		}
		return c.Redirect(http.StatusFound, "/groups/"+filteredReaders[0].ID+"/events")
	}

	return c.Redirect(http.StatusFound, "/dashboard")
}
