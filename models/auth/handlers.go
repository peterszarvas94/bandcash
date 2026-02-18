package auth

import (
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/labstack/echo/v4"

	"bandcash/internal/db"
	"bandcash/internal/email"
	"bandcash/internal/middleware"
	"bandcash/internal/utils"
)

type Auth struct {
	emailService *email.Service
}

func New() *Auth {
	return &Auth{
		emailService: email.NewFromEnv(),
	}
}

// LoginPage shows the login form
func (a *Auth) LoginPage(c echo.Context) error {
	return c.HTML(http.StatusOK, loginPageHTML)
}

// LoginRequest handles login form submission (sends magic link)
func (a *Auth) LoginRequest(c echo.Context) error {
	email := c.FormValue("email")
	if email == "" {
		return c.String(http.StatusBadRequest, "Email required")
	}

	// Check if user exists
	_, err := db.Qry.GetUserByEmail(c.Request().Context(), email)
	if err != nil {
		// User doesn't exist - offer to sign up instead
		return c.Redirect(http.StatusFound, "/auth/signup?email="+email)
	}

	// Create magic link
	token := utils.GenerateID("tok")
	expiresAt := time.Now().Add(1 * time.Hour)

	_, err = db.Qry.CreateMagicLink(c.Request().Context(), db.CreateMagicLinkParams{
		ID:        utils.GenerateID("mag"),
		Token:     token,
		Email:     email,
		Action:    "login",
		ExpiresAt: expiresAt,
	})
	if err != nil {
		slog.Error("auth: failed to create magic link", "err", err)
		return c.String(http.StatusInternalServerError, "Failed to create login link")
	}

	// Send email
	baseURL := os.Getenv("APP_BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}

	err = a.emailService.SendMagicLink(email, token, baseURL)
	if err != nil {
		slog.Error("auth: failed to send email", "err", err)
		return c.String(http.StatusInternalServerError, "Failed to send email")
	}

	return c.Redirect(http.StatusFound, "/auth/login-sent")
}

// SignupPage shows the signup form
func (a *Auth) SignupPage(c echo.Context) error {
	email := c.QueryParam("email")
	return c.HTML(http.StatusOK, signupPageHTML(email))
}

// SignupRequest handles signup form submission
func (a *Auth) SignupRequest(c echo.Context) error {
	email := c.FormValue("email")
	if email == "" {
		return c.String(http.StatusBadRequest, "Email required")
	}

	// Check if user already exists
	_, err := db.Qry.GetUserByEmail(c.Request().Context(), email)
	if err == nil {
		return c.String(http.StatusConflict, "User already exists. Please login instead.")
	}

	// Create user
	user, err := db.Qry.CreateUser(c.Request().Context(), db.CreateUserParams{
		ID:    utils.GenerateID("usr"),
		Email: email,
	})
	if err != nil {
		slog.Error("auth: failed to create user", "err", err)
		return c.String(http.StatusInternalServerError, "Failed to create user")
	}

	// Create magic link for login
	token := utils.GenerateID("tok")
	expiresAt := time.Now().Add(1 * time.Hour)

	_, err = db.Qry.CreateMagicLink(c.Request().Context(), db.CreateMagicLinkParams{
		ID:        utils.GenerateID("mag"),
		Token:     token,
		Email:     email,
		Action:    "login",
		ExpiresAt: expiresAt,
	})
	if err != nil {
		slog.Error("auth: failed to create magic link", "err", err)
		return c.String(http.StatusInternalServerError, "Failed to create login link")
	}

	slog.Info("auth: created user and magic link", "user_id", user.ID, "email", email)

	// Send welcome/login email
	baseURL := os.Getenv("APP_BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}

	err = a.emailService.SendMagicLink(email, token, baseURL)
	if err != nil {
		slog.Error("auth: failed to send email", "err", err)
		return c.String(http.StatusInternalServerError, "Failed to send email")
	}

	return c.Redirect(http.StatusFound, "/auth/login-sent")
}

// LoginSentPage shows confirmation that email was sent
func (a *Auth) LoginSentPage(c echo.Context) error {
	return c.HTML(http.StatusOK, loginSentPageHTML)
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
		slog.Error("auth: user not found for email", "email", magicLink.Email)
		return c.String(http.StatusNotFound, "User not found")
	}

	// Create session cookie
	c.SetCookie(&http.Cookie{
		Name:     "session",
		Value:    user.ID,
		Path:     "/",
		MaxAge:   86400 * 30, // 30 days
		HttpOnly: true,
		Secure:   os.Getenv("APP_ENV") == "production",
		SameSite: http.SameSiteStrictMode,
	})

	// Redirect based on whether user has a group
	group, _ := db.Qry.GetGroupByAdmin(c.Request().Context(), user.ID)
	if group.ID != "" {
		return c.Redirect(http.StatusFound, "/groups/"+group.ID+"/events")
	}

	// Check if user is a reader in any group
	// For now, redirect to create group page
	return c.Redirect(http.StatusFound, "/groups/new")
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
	return c.Redirect(http.StatusFound, "/")
}

// Dashboard shows user's groups or create group page
func (a *Auth) Dashboard(c echo.Context) error {
	userID := middleware.GetUserID(c)

	// Check if user is admin of a group
	group, err := db.Qry.GetGroupByAdmin(c.Request().Context(), userID)
	if err == nil && group.ID != "" {
		return c.Redirect(http.StatusFound, "/groups/"+group.ID+"/events")
	}

	// Check if user is reader in any groups
	// TODO: implement list of groups where user is reader

	// Show create group page
	return c.Redirect(http.StatusFound, "/groups/new")
}
