package payee

import (
	"log/slog"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/starfederation/datastar-go/datastar"

	"bandcash/internal/db"
	"bandcash/internal/utils"
)

type payeeParams struct {
	Name        string `json:"name" validate:"required,min=1,max=255"`
	Description string `json:"description" validate:"max=1000"`
}

type payeeTableParams struct {
	FormData payeeParams `json:"formData"`
	Mode     string      `json:"mode"`
}

type modeParams struct {
	Mode string `json:"mode"`
}

// Default signal state for resetting payee forms on success
var (
	defaultPayeeSignals = map[string]any{
		"mode":      "",
		"formState": "",
		"editingId": 0,
		"formData":  map[string]any{"name": "", "description": ""},
	}
	payeeErrorFields = []string{"name", "description"}
)

func (p *Payees) Index(c echo.Context) error {
	utils.EnsureClientID(c)

	data, err := p.GetIndexData(c.Request().Context())
	if err != nil {
		slog.Error("payee.list: failed to get data", "err", err)
		return c.NoContent(500)
	}

	slog.Debug("payee.index", "payee_count", len(data.Payees))
	return utils.RenderComponent(c, PayeeIndex(data))
}

func (p *Payees) Show(c echo.Context) error {
	utils.EnsureClientID(c)

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		slog.Warn("payee.show: invalid id", "err", err)
		return c.NoContent(400)
	}

	data, err := p.GetShowData(c.Request().Context(), id)
	if err != nil {
		slog.Error("payee.show: failed to get data", "err", err)
		return c.NoContent(500)
	}

	return utils.RenderComponent(c, PayeeShow(data))
}

func (p *Payees) Create(c echo.Context) error {
	var signals payeeTableParams
	err := datastar.ReadSignals(c.Request(), &signals)
	if err != nil {
		slog.Warn("payee.create.table: failed to read signals", "err", err)
		return c.NoContent(400)
	}

	// Validate
	if errs := utils.Validate(signals.FormData); errs != nil {
		utils.SSEHub.PatchSignals(c, map[string]any{"errors": utils.WithErrors(payeeErrorFields, errs)})
		return c.NoContent(422)
	}

	payee, err := db.Qry.CreatePayee(c.Request().Context(), db.CreatePayeeParams{
		Name:        signals.FormData.Name,
		Description: signals.FormData.Description,
	})
	if err != nil {
		slog.Error("payee.create.table: failed to create payee", "err", err)
		return c.NoContent(500)
	}

	slog.Debug("payee.create.table", "id", payee.ID, "name", payee.Name)

	utils.SSEHub.PatchSignals(c, defaultPayeeSignals)
	data, err := p.GetIndexData(c.Request().Context())
	if err != nil {
		slog.Error("payee.create.table: failed to get data", "err", err)
		return c.NoContent(500)
	}
	html, err := utils.RenderComponentString(c.Request().Context(), PayeeIndex(data))
	if err != nil {
		slog.Error("payee.create.table: failed to render", "err", err)
		return c.NoContent(500)
	}

	utils.SSEHub.PatchHTML(c, html)

	return c.NoContent(200)
}

func (p *Payees) Update(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		slog.Warn("payee.update: invalid id", "err", err)
		return c.NoContent(400)
	}

	var signals payeeTableParams
	err = datastar.ReadSignals(c.Request(), &signals)
	if err != nil {
		slog.Warn("payee.update: failed to read signals", "err", err)
		return c.NoContent(400)
	}

	// Validate
	if errs := utils.Validate(signals.FormData); errs != nil {
		utils.SSEHub.PatchSignals(c, map[string]any{"errors": utils.WithErrors(payeeErrorFields, errs)})
		return c.NoContent(422)
	}

	_, err = db.Qry.UpdatePayee(c.Request().Context(), db.UpdatePayeeParams{
		Name:        signals.FormData.Name,
		Description: signals.FormData.Description,
		ID:          int64(id),
	})
	if err != nil {
		slog.Error("payee.update: failed to update payee", "err", err)
		return c.NoContent(500)
	}

	slog.Debug("payee.update", "id", id)

	if signals.Mode == "single" {
		err = utils.SSEHub.Redirect(c, "/payee/"+strconv.Itoa(id))
		if err != nil {
			slog.Warn("payee.update: failed to redirect", "err", err)
		}
		return c.NoContent(200)
	}

	utils.SSEHub.PatchSignals(c, defaultPayeeSignals)
	data, err := p.GetIndexData(c.Request().Context())
	if err != nil {
		slog.Error("payee.update: failed to get data", "err", err)
		return c.NoContent(500)
	}
	html, err := utils.RenderComponentString(c.Request().Context(), PayeeIndex(data))
	if err != nil {
		slog.Error("payee.update: failed to render", "err", err)
		return c.NoContent(500)
	}

	utils.SSEHub.PatchHTML(c, html)

	return c.NoContent(200)
}

func (p *Payees) Destroy(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		slog.Warn("payee.destroy: invalid id", "err", err)
		return c.NoContent(400)
	}

	var signals modeParams
	err = datastar.ReadSignals(c.Request(), &signals)
	if err != nil {
		slog.Warn("payee.destroy: failed to read signals", "err", err)
		return c.NoContent(400)
	}

	err = db.Qry.DeletePayee(c.Request().Context(), int64(id))
	if err != nil {
		slog.Error("payee.destroy: failed to delete payee", "err", err)
		return c.NoContent(500)
	}

	slog.Debug("payee.destroy", "id", id)

	if signals.Mode == "single" {
		err = utils.SSEHub.Redirect(c, "/payee")
		if err != nil {
			slog.Warn("payee.destroy: failed to redirect", "err", err)
		}
		return c.NoContent(200)
	}

	utils.SSEHub.PatchSignals(c, defaultPayeeSignals)
	data, err := p.GetIndexData(c.Request().Context())
	if err != nil {
		slog.Error("payee.destroy: failed to get data", "err", err)
		return c.NoContent(500)
	}
	html, err := utils.RenderComponentString(c.Request().Context(), PayeeIndex(data))
	if err != nil {
		slog.Error("payee.destroy: failed to render", "err", err)
		return c.NoContent(500)
	}

	utils.SSEHub.PatchHTML(c, html)

	return c.NoContent(200)
}
