package devnotifications

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/starfederation/datastar-go/datastar"

	"bandcash/internal/utils"
)

type devSignals struct {
	FormData struct {
		Name string `json:"name"`
	} `json:"formData"`
}

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
			"errors": map[string]any{"name": "Field is required."},
		})
		return c.NoContent(http.StatusUnprocessableEntity)
	}

	utils.SSEHub.PatchSignals(c, map[string]any{
		"errors":   map[string]any{},
		"formData": map[string]any{"name": ""},
	})
	utils.Notify(c, "success", "Inline validation passed.")
	if err := h.patchPage(c); err != nil {
		return c.NoContent(http.StatusInternalServerError)
	}

	return c.NoContent(http.StatusOK)
}

func (h *DevNotifications) TestSuccess(c echo.Context) error {
	utils.Notify(c, "success", "Success notification test.")
	if err := h.patchPage(c); err != nil {
		return c.NoContent(http.StatusInternalServerError)
	}
	return c.NoContent(http.StatusOK)
}

func (h *DevNotifications) TestError(c echo.Context) error {
	utils.Notify(c, "error", "Error notification test.")
	if err := h.patchPage(c); err != nil {
		return c.NoContent(http.StatusInternalServerError)
	}
	return c.NoContent(http.StatusOK)
}

func (h *DevNotifications) TestInfo(c echo.Context) error {
	utils.Notify(c, "info", "Info notification test.")
	if err := h.patchPage(c); err != nil {
		return c.NoContent(http.StatusInternalServerError)
	}
	return c.NoContent(http.StatusOK)
}

func (h *DevNotifications) TestWarning(c echo.Context) error {
	utils.Notify(c, "warning", "Warning notification test.")
	if err := h.patchPage(c); err != nil {
		return c.NoContent(http.StatusInternalServerError)
	}
	return c.NoContent(http.StatusOK)
}

func (h *DevNotifications) patchPage(c echo.Context) error {
	html, err := utils.RenderComponentStringFor(c, Index())
	if err != nil {
		return err
	}
	utils.SSEHub.PatchHTML(c, html)
	return nil
}
