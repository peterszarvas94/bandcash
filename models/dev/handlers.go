package dev

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	ctxi18n "github.com/invopop/ctxi18n/i18n"
	"github.com/labstack/echo/v4"
	"github.com/starfederation/datastar-go/datastar"

	"bandcash/internal/email"
	appi18n "bandcash/internal/i18n"
	"bandcash/internal/utils"
	authstore "bandcash/models/auth/data"
	shared "bandcash/models/shared"
	icons "bandcash/models/shared/icons"
)

type devSignals struct {
	TabID    string `json:"tab_id"`
	FormData struct {
		Name string `json:"name" validate:"required,min=1,max=255"`
	} `json:"formData"`
}

type devTabSignals struct {
	TabID string `json:"tab_id"`
}

var devErrorFields = []string{"name"}

func devSessionUser(c echo.Context) (bool, string) {
	if cookie, err := c.Cookie(utils.SessionCookieName); err == nil && cookie.Value != "" {
		if session, err := authstore.GetUserSessionByToken(c.Request().Context(), cookie.Value); err == nil {
			if user, err := authstore.GetUserByID(c.Request().Context(), session.UserID); err == nil {
				return true, user.Email
			}

			return true, ""
		}
	}

	return false, ""
}

func DevPageHandler(c echo.Context) error {
	utils.EnsureTabID(c)
	selector := strings.TrimSpace(c.QueryParam("selector"))
	switch selector {
	case "1", "2", "3":
	default:
		selector = "1"
	}
	isAuthenticated, userEmail := devSessionUser(c)
	data := DevPageData{
		Title:           "bandcash - Dev tools",
		Breadcrumbs:     []utils.Crumb{{Label: "Dev tools"}},
		Signals:         DevPageSignals(),
		LinkSelector:    selector,
		IsAuthenticated: isAuthenticated,
		UserEmail:       userEmail,
	}
	return utils.RenderPage(c, DevPage(data))
}

func TestInline(c echo.Context) error {
	signals := devSignals{}
	if err := datastar.ReadSignals(c.Request(), &signals); err != nil {
		return c.NoContent(http.StatusBadRequest)
	}
	if !utils.SetTabID(c, signals.TabID) {
		return c.NoContent(http.StatusBadRequest)
	}
	signals.FormData.Name = strings.TrimSpace(signals.FormData.Name)

	if errs := utils.ValidateWithLocale(c.Request().Context(), signals.FormData); errs != nil {
		utils.SSEHub.PatchSignals(c, map[string]any{
			"errors": utils.WithErrors(devErrorFields, errs),
		})
		return c.NoContent(http.StatusUnprocessableEntity)
	}

	utils.SSEHub.PatchSignals(c, map[string]any{
		"errors":   utils.GetEmptyErrors(devErrorFields),
		"formData": map[string]any{"name": ""},
	})
	utils.Notify(c, "Inline validation passed")
	if err := patchNotifications(c); err != nil {
		return c.NoContent(http.StatusInternalServerError)
	}

	return c.NoContent(http.StatusOK)
}

func Test(c echo.Context) error {
	signals := devTabSignals{}
	if err := datastar.ReadSignals(c.Request(), &signals); err != nil {
		return c.NoContent(http.StatusBadRequest)
	}
	if !utils.SetTabID(c, signals.TabID) {
		return c.NoContent(http.StatusBadRequest)
	}

	utils.Notify(c, "Notification test")
	if err := patchNotifications(c); err != nil {
		return c.NoContent(http.StatusInternalServerError)
	}
	return c.NoContent(http.StatusOK)
}

func TestSpinner(c echo.Context) error {
	delay := 500
	if raw := strings.TrimSpace(c.QueryParam("ms")); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err == nil && parsed >= 0 && parsed <= 10000 {
			delay = parsed
		}
	}
	time.Sleep(time.Duration(delay) * time.Millisecond)
	return c.NoContent(http.StatusOK)
}

func TestMultiAction(c echo.Context) error {
	signals := devTabSignals{}
	if err := datastar.ReadSignals(c.Request(), &signals); err != nil {
		return c.NoContent(http.StatusBadRequest)
	}
	if !utils.SetTabID(c, signals.TabID) {
		return c.NoContent(http.StatusBadRequest)
	}

	action := c.Param("action")
	time.Sleep(1200 * time.Millisecond)

	utils.SSEHub.PatchSignals(c, map[string]any{
		"multiActionBusy":   false,
		"multiActionActive": "",
	})

	utils.Notify(c, "Completed: "+action)
	if err := patchNotifications(c); err != nil {
		return c.NoContent(http.StatusInternalServerError)
	}

	return c.NoContent(http.StatusOK)
}

func PreviewLoginEmail(c echo.Context) error {
	subject, html, err := email.Email().PreviewMagicLinkHTML(c.Request().Context(), "tok_12345678901234567890", devBaseURL(c))
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return renderEmailPreview(c, EmailPreviewData{
		Title:    "Login email preview",
		From:     utils.Env().EmailFrom,
		To:       "user@example.com",
		Subject:  subject,
		BodyHTML: html,
	})
}

func PreviewInviteEmail(c echo.Context) error {
	subject, html, err := email.Email().PreviewGroupInvitationHTML(c.Request().Context(), "Preview Group", "tok_ABCDEFGHIJ1234567890", devBaseURL(c))
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return renderEmailPreview(c, EmailPreviewData{
		Title:    "Invite email preview",
		From:     utils.Env().EmailFrom,
		To:       "invitee@example.com",
		Subject:  subject,
		BodyHTML: html,
	})
}

func PreviewInviteAcceptedEmail(c echo.Context) error {
	subject, html, err := email.Email().PreviewInviteAcceptedHTML(c.Request().Context(), "Preview Group", "grp_preview1234567890", devBaseURL(c))
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return renderEmailPreview(c, EmailPreviewData{
		Title:    "Invite accepted email preview",
		From:     utils.Env().EmailFrom,
		To:       "admin@example.com",
		Subject:  subject,
		BodyHTML: html,
	})
}

func PreviewGroupCreatedEmail(c echo.Context) error {
	subject, html, err := email.Email().PreviewGroupCreatedHTML(c.Request().Context(), "Preview Group", "grp_preview1234567890", devBaseURL(c))
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return renderEmailPreview(c, EmailPreviewData{
		Title:    "Group created email preview",
		From:     utils.Env().EmailFrom,
		To:       "owner@example.com",
		Subject:  subject,
		BodyHTML: html,
	})
}

func PreviewRoleUpgradedEmail(c echo.Context) error {
	subject, html, err := email.Email().PreviewRoleUpgradedToAdminHTML(c.Request().Context(), "Preview Group", "grp_preview1234567890", devBaseURL(c))
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return renderEmailPreview(c, EmailPreviewData{
		Title:    "Role upgraded email preview",
		From:     utils.Env().EmailFrom,
		To:       "member@example.com",
		Subject:  subject,
		BodyHTML: html,
	})
}

func PreviewRoleDowngradedEmail(c echo.Context) error {
	subject, html, err := email.Email().PreviewRoleDowngradedToViewerHTML(c.Request().Context(), "Preview Group", "grp_preview1234567890", devBaseURL(c))
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return renderEmailPreview(c, EmailPreviewData{
		Title:    "Role downgraded email preview",
		From:     utils.Env().EmailFrom,
		To:       "member@example.com",
		Subject:  subject,
		BodyHTML: html,
	})
}

func PreviewAccessRemovedEmail(c echo.Context) error {
	subject, html, err := email.Email().PreviewAccessRemovedHTML(c.Request().Context(), "Preview Group", []string{"admin.one@example.com", "admin.two@example.com"}, devBaseURL(c))
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return renderEmailPreview(c, EmailPreviewData{
		Title:    "Access removed email preview",
		From:     utils.Env().EmailFrom,
		To:       "member@example.com",
		Subject:  subject,
		BodyHTML: html,
	})
}

func PreviewInvalidLinkErrorPage(c echo.Context) error {
	return renderDevErrorPage(c, http.StatusBadRequest, icons.IconLink2Off, "error_pages.link.invalid_title", "error_pages.link.invalid_body")
}

func PreviewBadRequestErrorPage(c echo.Context) error {
	return echo.NewHTTPError(http.StatusBadRequest)
}

func PreviewForbiddenErrorPage(c echo.Context) error {
	return echo.NewHTTPError(http.StatusForbidden)
}

func PreviewNotFoundErrorPage(c echo.Context) error {
	return echo.NewHTTPError(http.StatusNotFound)
}

func PreviewInternalErrorPage(c echo.Context) error {
	return echo.NewHTTPError(http.StatusInternalServerError)
}

func renderEmailPreview(c echo.Context, data EmailPreviewData) error {
	return utils.RenderPage(c, DevEmailPreviewPage(data))
}

func renderDevErrorPage(c echo.Context, status int, iconName icons.IconName, titleKey, bodyKey string) error {
	ctx := c.Request().Context()
	isAuthenticated, isSuperAdmin := utils.ResolveAuthState(c)
	return utils.RenderPage(c, shared.ErrorPage(shared.ErrorPageData{
		Title:           ctxi18n.T(ctx, titleKey),
		StatusCode:      status,
		IconName:        iconName,
		Heading:         ctxi18n.T(ctx, titleKey),
		Message:         ctxi18n.T(ctx, bodyKey),
		HomeLabel:       ctxi18n.T(ctx, "error_pages.home_action"),
		HomeHref:        appi18n.LocalizedHomePath(ctx),
		IsAuthenticated: isAuthenticated,
		IsSuperAdmin:    isSuperAdmin,
	}))
}

func devBaseURL(c echo.Context) string {
	configured := strings.TrimSpace(utils.Env().URL)
	if configured != "" {
		return configured
	}
	return fmt.Sprintf("%s://%s", c.Scheme(), c.Request().Host)
}

func patchNotifications(c echo.Context) error {
	html, err := utils.RenderHTMLForRequest(c, shared.Notifications())
	if err != nil {
		return err
	}
	utils.SSEHub.PatchHTML(c, html)
	return nil
}
