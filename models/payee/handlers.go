package payee

import (
	"context"
	"errors"
	"log/slog"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/starfederation/datastar-go/datastar"

	"bandcash/internal/utils"
)

type payeeParams struct {
	Name        string `json:"name" validate:"required,min=1,max=255"`
	Description string `json:"description" validate:"max=1000"`
}

type payeeTableParams struct {
	FormData payeeParams `json:"formData"`
}

// Default signal state for resetting payee forms on success
var (
	defaultPayeeSignals = map[string]any{
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
		return c.String(500, "Internal Server Error")
	}

	slog.Debug("payee.index", "payee_count", len(data.(PayeesData).Payees))
	return utils.RenderTemplate(c.Response().Writer, p.tmpl, "index", data)
}

func (p *Payees) Show(c echo.Context) error {
	utils.EnsureClientID(c)

	id, err := utils.ParamInt(c, "id")
	if err != nil {
		return c.String(400, "Invalid ID")
	}

	data, err := p.GetShowData(c.Request().Context(), id)
	if err != nil {
		slog.Error("payee.show: failed to get data", "err", err)
		return c.String(500, "Internal Server Error")
	}

	return utils.RenderTemplate(c.Response().Writer, p.tmpl, "show", data)
}

func (p *Payees) Edit(c echo.Context) error {
	utils.EnsureClientID(c)

	id, err := utils.ParamInt(c, "id")
	if err != nil {
		return c.String(400, "Invalid ID")
	}

	data, err := p.GetEditData(c.Request().Context(), id)
	if err != nil {
		slog.Error("payee.edit: failed to get data", "err", err)
		return c.String(500, "Internal Server Error")
	}

	return utils.RenderTemplate(c.Response().Writer, p.tmpl, "edit", data)
}

func (p *Payees) Create(c echo.Context) error {
	var signals payeeTableParams
	if err := datastar.ReadSignals(c.Request(), &signals); err != nil {
		slog.Warn("payee.create.table: failed to read signals", "err", err)
		return c.NoContent(400)
	}

	// Validate
	if errs := utils.Validate(signals.FormData); errs != nil {
		utils.SSEHub.PatchSignals(c, map[string]any{"errors": utils.WithErrors(payeeErrorFields, errs)})
		return c.NoContent(422)
	}

	payee, err := p.CreatePayee(c.Request().Context(), signals.FormData.Name, signals.FormData.Description)
	if err != nil {
		slog.Error("payee.create.table: failed to create payee", "err", err)
		return c.String(500, "Internal Server Error")
	}

	slog.Debug("payee.create.table", "id", payee.ID, "name", payee.Name)

	utils.SSEHub.PatchSignals(c, defaultPayeeSignals)
	data, err := p.GetIndexData(c.Request().Context())
	if err != nil {
		slog.Error("payee.create.table: failed to get data", "err", err)
		return c.NoContent(500)
	}
	html, err := utils.RenderBlock(p.tmpl, "payee-index", data)
	if err != nil {
		slog.Error("payee.create.table: failed to render", "err", err)
	} else {
		utils.SSEHub.PatchHTML(c, html)
	}

	return c.NoContent(200)
}

func (p *Payees) Update(c echo.Context) error {
	id, err := utils.ParamInt(c, "id")
	if err != nil {
		return c.String(400, "Invalid ID")
	}

	var signals payeeTableParams
	if err := datastar.ReadSignals(c.Request(), &signals); err != nil {
		slog.Warn("payee.update: failed to read signals", "err", err)
		return c.NoContent(400)
	}

	// Validate
	if errs := utils.Validate(signals.FormData); errs != nil {
		utils.SSEHub.PatchSignals(c, map[string]any{"errors": utils.WithErrors(payeeErrorFields, errs)})
		return c.NoContent(422)
	}

	_, err = p.UpdatePayee(c.Request().Context(), id, signals.FormData.Name, signals.FormData.Description)
	if err != nil {
		slog.Error("payee.update: failed to update payee", "err", err)
		return c.String(500, "Internal Server Error")
	}

	slog.Debug("payee.update", "id", id)

	if err := utils.SSEHub.Redirect(c, "/payee/"+strconv.Itoa(id)); err != nil {
		if errors.Is(err, context.Canceled) {
			slog.Debug("payee.update: redirect cancelled", "err", err)
		} else {
			slog.Warn("payee.update: failed to redirect", "err", err)
		}
	}

	return c.NoContent(200)
}

func (p *Payees) UpdateSingle(c echo.Context) error {
	id, err := utils.ParamInt(c, "id")
	if err != nil {
		return c.String(400, "Invalid ID")
	}

	var signals payeeTableParams
	if err := datastar.ReadSignals(c.Request(), &signals); err != nil {
		slog.Warn("payee.update.single: failed to read signals", "err", err)
		return c.NoContent(400)
	}

	// Validate
	if errs := utils.Validate(signals.FormData); errs != nil {
		utils.SSEHub.PatchSignals(c, map[string]any{"errors": utils.WithErrors(payeeErrorFields, errs)})
		return c.NoContent(422)
	}

	_, err = p.UpdatePayee(c.Request().Context(), id, signals.FormData.Name, signals.FormData.Description)
	if err != nil {
		slog.Error("payee.update.single: failed to update payee", "err", err)
		return c.String(500, "Internal Server Error")
	}

	slog.Debug("payee.update.single", "id", id)

	utils.SSEHub.PatchSignals(c, defaultPayeeSignals)
	data, err := p.GetIndexData(c.Request().Context())
	if err != nil {
		slog.Error("payee.update.single: failed to get data", "err", err)
		return c.NoContent(500)
	}
	html, err := utils.RenderBlock(p.tmpl, "payee-index", data)
	if err != nil {
		slog.Error("payee.update.single: failed to render", "err", err)
	} else {
		utils.SSEHub.PatchHTML(c, html)
	}

	return c.NoContent(200)
}

func (p *Payees) Destroy(c echo.Context) error {
	id, err := utils.ParamInt(c, "id")
	if err != nil {
		slog.Warn("payee.destroy: invalid id", "err", err)
		return c.NoContent(400)
	}

	if err := p.DeletePayee(c.Request().Context(), id); err != nil {
		slog.Error("payee.destroy: failed to delete payee", "err", err)
		return c.String(500, "Internal Server Error")
	}

	slog.Debug("payee.destroy", "id", id)

	if err := utils.SSEHub.Redirect(c, "/payee"); err != nil {
		slog.Warn("payee.destroy: failed to redirect", "err", err)
	}

	return c.NoContent(200)
}
