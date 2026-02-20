package devnotifications

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/starfederation/datastar-go/datastar"

	"bandcash/internal/utils"
	shared "bandcash/models/shared"
)

type devSignals struct {
	FormData struct {
		Name string `json:"name"`
	} `json:"formData"`
}

var devErrorFields = []string{"name"}

func (h *DevNotifications) Index(c echo.Context) error {
	utils.EnsureClientID(c)
	return utils.RenderComponent(c, Index())
}

func (h *DevNotifications) TestInline(c echo.Context) error {
	signals := devSignals{}
	if err := datastar.ReadSignals(c.Request(), &signals); err != nil {
		return c.NoContent(http.StatusBadRequest)
	}

	if strings.TrimSpace(signals.FormData.Name) == "" {
		utils.SSEHub.PatchSignals(c, map[string]any{
			"errors": utils.WithErrors(devErrorFields, map[string]string{"name": "Field is required."}),
		})
		return c.NoContent(http.StatusUnprocessableEntity)
	}

	utils.SSEHub.PatchSignals(c, map[string]any{
		"errors":   utils.GetEmptyErrors(devErrorFields),
		"formData": map[string]any{"name": ""},
	})
	utils.Notify(c, "success", "Inline validation passed.")
	if err := h.patchNotifications(c); err != nil {
		return c.NoContent(http.StatusInternalServerError)
	}

	return c.NoContent(http.StatusOK)
}

func (h *DevNotifications) TestSuccess(c echo.Context) error {
	utils.Notify(c, "success", "Success notification test.")
	if err := h.patchNotifications(c); err != nil {
		return c.NoContent(http.StatusInternalServerError)
	}
	return c.NoContent(http.StatusOK)
}

func (h *DevNotifications) TestError(c echo.Context) error {
	utils.Notify(c, "error", "Error notification test.")
	if err := h.patchNotifications(c); err != nil {
		return c.NoContent(http.StatusInternalServerError)
	}
	return c.NoContent(http.StatusOK)
}

func (h *DevNotifications) TestInfo(c echo.Context) error {
	utils.Notify(c, "info", "Info notification test.")
	if err := h.patchNotifications(c); err != nil {
		return c.NoContent(http.StatusInternalServerError)
	}
	return c.NoContent(http.StatusOK)
}

func (h *DevNotifications) TestWarning(c echo.Context) error {
	utils.Notify(c, "warning", "Warning notification test.")
	if err := h.patchNotifications(c); err != nil {
		return c.NoContent(http.StatusInternalServerError)
	}
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
