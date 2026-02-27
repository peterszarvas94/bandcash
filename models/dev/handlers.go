package dev

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/starfederation/datastar-go/datastar"
	"net/http"
	"strconv"
	"strings"
	"time"

	"bandcash/internal/email"
	"bandcash/internal/utils"
	shared "bandcash/models/shared"
)

type devSignals struct {
	FormData struct {
		Name string `json:"name" validate:"required,min=1,max=255"`
	} `json:"formData"`
}

var devErrorFields = []string{"name"}

func (h *DevNotifications) DevPageHandler(c echo.Context) error {
	utils.EnsureClientID(c)
	return utils.RenderComponent(c, DevPage())
}

func (h *DevNotifications) TestInline(c echo.Context) error {
	signals := devSignals{}
	if err := datastar.ReadSignals(c.Request(), &signals); err != nil {
		return c.NoContent(http.StatusBadRequest)
	}
	signals.FormData.Name = utils.NormalizeText(signals.FormData.Name)

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
	utils.Notify(c, "success", "Inline validation passed")
	if err := h.patchNotifications(c); err != nil {
		return c.NoContent(http.StatusInternalServerError)
	}

	return c.NoContent(http.StatusOK)
}

func (h *DevNotifications) TestSuccess(c echo.Context) error {
	utils.Notify(c, "success", "Success notification test")
	if err := h.patchNotifications(c); err != nil {
		return c.NoContent(http.StatusInternalServerError)
	}
	return c.NoContent(http.StatusOK)
}

func (h *DevNotifications) TestError(c echo.Context) error {
	utils.Notify(c, "error", "Error notification test")
	if err := h.patchNotifications(c); err != nil {
		return c.NoContent(http.StatusInternalServerError)
	}
	return c.NoContent(http.StatusOK)
}

func (h *DevNotifications) TestInfo(c echo.Context) error {
	utils.Notify(c, "info", "Info notification test")
	if err := h.patchNotifications(c); err != nil {
		return c.NoContent(http.StatusInternalServerError)
	}
	return c.NoContent(http.StatusOK)
}

func (h *DevNotifications) TestWarning(c echo.Context) error {
	utils.Notify(c, "warning", "Warning notification test")
	if err := h.patchNotifications(c); err != nil {
		return c.NoContent(http.StatusInternalServerError)
	}
	return c.NoContent(http.StatusOK)
}

func (h *DevNotifications) TestBodyLimitGlobal(c echo.Context) error {
	return c.NoContent(http.StatusOK)
}

func (h *DevNotifications) TestBodyLimitAuth(c echo.Context) error {
	return c.NoContent(http.StatusOK)
}

func (h *DevNotifications) TestSpinner(c echo.Context) error {
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

func (h *DevNotifications) TestMultiAction(c echo.Context) error {
	action := c.Param("action")
	time.Sleep(1200 * time.Millisecond)

	utils.SSEHub.PatchSignals(c, map[string]any{
		"multiActionBusy":   false,
		"multiActionActive": "",
	})

	utils.Notify(c, "info", "Completed: "+action)
	if err := h.patchNotifications(c); err != nil {
		return c.NoContent(http.StatusInternalServerError)
	}

	return c.NoContent(http.StatusOK)
}

func (h *DevNotifications) PreviewLoginEmail(c echo.Context) error {
	html, err := email.Email().PreviewMagicLinkHTML(c.Request().Context(), "tok_12345678901234567890", devBaseURL(c))
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return c.HTML(http.StatusOK, html)
}

func (h *DevNotifications) PreviewInviteEmail(c echo.Context) error {
	html, err := email.Email().PreviewGroupInvitationHTML(c.Request().Context(), "Preview Group", "tok_ABCDEFGHIJ1234567890", devBaseURL(c))
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return c.HTML(http.StatusOK, html)
}

func devBaseURL(c echo.Context) string {
	configured := strings.TrimSpace(utils.Env().URL)
	if configured != "" {
		return configured
	}
	return fmt.Sprintf("%s://%s", c.Scheme(), c.Request().Host)
}

func (h *DevNotifications) patchNotifications(c echo.Context) error {
	html, err := utils.RenderComponentStringFor(c, shared.Notifications())
	if err != nil {
		return err
	}
	utils.SSEHub.PatchHTML(c, html)
	return nil
}
