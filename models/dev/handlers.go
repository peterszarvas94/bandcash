package dev

import (
	"net/http"

	ctxi18n "github.com/invopop/ctxi18n/i18n"
	"github.com/labstack/echo/v4"
	"github.com/starfederation/datastar-go/datastar"

	"bandcash/internal/utils"
	shared "bandcash/models/shared"
)

type devSignals struct {
	FormData struct {
		Name string `json:"name" validate:"required,min=1,max=255"`
	} `json:"formData"`
}

var devErrorFields = []string{"name"}

func (h *DevNotifications) Index(c echo.Context) error {
	utils.EnsureClientID(c)
	return utils.RenderComponent(c, Index())
}

func (h *DevNotifications) RateLimitPage(c echo.Context) error {
	utils.EnsureClientID(c)
	return utils.RenderComponent(c, RateLimit())
}

func (h *DevNotifications) BodyLimitPage(c echo.Context) error {
	utils.EnsureClientID(c)
	return utils.RenderComponent(c, BodyLimit())
}

func (h *DevNotifications) Redirect(c echo.Context) error {
	return c.Redirect(http.StatusFound, "/dev/notifications")
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
	utils.Notify(c, "success", ctxi18n.T(c.Request().Context(), "dev.notifications.inline_passed"))
	if err := h.patchNotifications(c); err != nil {
		return c.NoContent(http.StatusInternalServerError)
	}

	return c.NoContent(http.StatusOK)
}

func (h *DevNotifications) TestSuccess(c echo.Context) error {
	utils.Notify(c, "success", ctxi18n.T(c.Request().Context(), "dev.notifications.success_test"))
	if err := h.patchNotifications(c); err != nil {
		return c.NoContent(http.StatusInternalServerError)
	}
	return c.NoContent(http.StatusOK)
}

func (h *DevNotifications) TestError(c echo.Context) error {
	utils.Notify(c, "error", ctxi18n.T(c.Request().Context(), "dev.notifications.error_test"))
	if err := h.patchNotifications(c); err != nil {
		return c.NoContent(http.StatusInternalServerError)
	}
	return c.NoContent(http.StatusOK)
}

func (h *DevNotifications) TestInfo(c echo.Context) error {
	utils.Notify(c, "info", ctxi18n.T(c.Request().Context(), "dev.notifications.info_test"))
	if err := h.patchNotifications(c); err != nil {
		return c.NoContent(http.StatusInternalServerError)
	}
	return c.NoContent(http.StatusOK)
}

func (h *DevNotifications) TestWarning(c echo.Context) error {
	utils.Notify(c, "warning", ctxi18n.T(c.Request().Context(), "dev.notifications.warning_test"))
	if err := h.patchNotifications(c); err != nil {
		return c.NoContent(http.StatusInternalServerError)
	}
	return c.NoContent(http.StatusOK)
}

func (h *DevNotifications) BodyLimitGlobalTest(c echo.Context) error {
	return c.NoContent(http.StatusOK)
}

func (h *DevNotifications) BodyLimitAuthTest(c echo.Context) error {
	return c.NoContent(http.StatusOK)
}

func (h *DevNotifications) patchNotifications(c echo.Context) error {
	html, err := utils.RenderComponentStringFor(c, shared.Notifications())
	if err != nil {
		return err
	}
	utils.SSEHub.PatchHTML(c, html)
	return nil
}
