package payee

import (
	"log/slog"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/starfederation/datastar-go/datastar"

	"bandcash/internal/hub"
	"bandcash/internal/utils"
	"bandcash/internal/view"
)

type payeeParams struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type payeeTableParams struct {
	FormData payeeParams `json:"formData"`
}

// Default signal state for resetting payee forms
var defaultPayeeSignals = map[string]any{
	"formState": "",
	"editingId": 0,
	"formData":  map[string]any{"name": "", "description": ""},
}

func (p *Payees) Index(c echo.Context) error {
	utils.EnsureClientID(c)

	data, err := p.GetIndexData(c.Request().Context())
	if err != nil {
		slog.Error("payee.list: failed to get data", "err", err)
		return c.String(500, "Internal Server Error")
	}

	slog.Debug("payee.index", "payee_count", len(data.(PayeesData).Payees))
	return p.tmpl.ExecuteTemplate(c.Response().Writer, "index", data)
}

func (p *Payees) New(c echo.Context) error {
	utils.EnsureClientID(c)
	return p.tmpl.ExecuteTemplate(c.Response().Writer, "new", PayeeData{
		Title: "New Payee",
		Breadcrumbs: []view.Crumb{
			{Label: "Payees", Href: "/payee"},
			{Label: "New"},
		},
	})
}

func (p *Payees) Show(c echo.Context) error {
	utils.EnsureClientID(c)

	id, err := utils.ParamInt(c, "id")
	if err != nil {
		return c.String(400, "Invalid ID")
	}

	payee, err := p.GetPayee(c.Request().Context(), id)
	if err != nil {
		slog.Error("payee.show: failed to get payee", "err", err)
		return c.String(500, "Internal Server Error")
	}

	entries, err := p.GetEntries(c.Request().Context(), id)
	if err != nil {
		slog.Error("payee.show: failed to get entries", "err", err)
		return c.String(500, "Internal Server Error")
	}

	return p.tmpl.ExecuteTemplate(c.Response().Writer, "show", PayeeData{
		Title:   payee.Name,
		Payee:   payee,
		Entries: entries,
		Breadcrumbs: []view.Crumb{
			{Label: "Payees", Href: "/payee"},
			{Label: payee.Name},
		},
	})
}

func (p *Payees) Edit(c echo.Context) error {
	utils.EnsureClientID(c)

	id, err := utils.ParamInt(c, "id")
	if err != nil {
		return c.String(400, "Invalid ID")
	}

	payee, err := p.GetPayee(c.Request().Context(), id)
	if err != nil {
		slog.Error("payee.edit: failed to get payee", "err", err)
		return c.String(500, "Internal Server Error")
	}

	return p.tmpl.ExecuteTemplate(c.Response().Writer, "edit", PayeeData{
		Title: "Edit Payee",
		Payee: payee,
		Breadcrumbs: []view.Crumb{
			{Label: "Payees", Href: "/payee"},
			{Label: payee.Name, Href: "/payee/" + strconv.Itoa(id)},
			{Label: "Edit"},
		},
	})
}

func (p *Payees) Create(c echo.Context) error {
	var signals payeeTableParams
	if err := datastar.ReadSignals(c.Request(), &signals); err != nil {
		slog.Warn("payee.create: failed to read signals", "err", err)
		return c.NoContent(400)
	}

	if signals.FormData.Name == "" {
		slog.Debug("payee.create: empty name")
		return c.NoContent(200)
	}

	payee, err := p.CreatePayee(c.Request().Context(), signals.FormData.Name, signals.FormData.Description)
	if err != nil {
		slog.Error("payee.create: failed to create payee", "err", err)
		return c.String(500, "Internal Server Error")
	}

	slog.Debug("payee.create", "id", payee.ID, "name", payee.Name)

	if err := hub.Hub.Redirect(c, "/payee/"+strconv.FormatInt(payee.ID, 10)); err != nil {
		slog.Warn("payee.create: failed to redirect", "err", err)
	}

	return c.NoContent(200)
}

func (p *Payees) CreateTable(c echo.Context) error {
	var signals payeeTableParams
	if err := datastar.ReadSignals(c.Request(), &signals); err != nil {
		slog.Warn("payee.create.table: failed to read signals", "err", err)
		return c.NoContent(400)
	}

	if signals.FormData.Name == "" {
		slog.Debug("payee.create.table: empty name")
		return c.NoContent(200)
	}

	payee, err := p.CreatePayee(c.Request().Context(), signals.FormData.Name, signals.FormData.Description)
	if err != nil {
		slog.Error("payee.create.table: failed to create payee", "err", err)
		return c.String(500, "Internal Server Error")
	}

	slog.Debug("payee.create.table", "id", payee.ID, "name", payee.Name)

	if err := hub.Hub.PatchSignals(c, defaultPayeeSignals); err != nil {
		slog.Warn("payee.create.table: failed to patch signals", "err", err)
	}

	if err := hub.Hub.Refresh(c); err != nil {
		slog.Warn("payee.create.table: failed to signal client", "err", err)
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

	_, err = p.UpdatePayee(c.Request().Context(), id, signals.FormData.Name, signals.FormData.Description)
	if err != nil {
		slog.Error("payee.update: failed to update payee", "err", err)
		return c.String(500, "Internal Server Error")
	}

	slog.Debug("payee.update", "id", id)

	if err := hub.Hub.Redirect(c, "/payee/"+strconv.Itoa(id)); err != nil {
		slog.Warn("payee.update: failed to redirect", "err", err)
	}

	return c.NoContent(200)
}

func (p *Payees) UpdateTable(c echo.Context) error {
	id, err := utils.ParamInt(c, "id")
	if err != nil {
		return c.String(400, "Invalid ID")
	}

	var signals payeeTableParams
	if err := datastar.ReadSignals(c.Request(), &signals); err != nil {
		slog.Warn("payee.update.table: failed to read signals", "err", err)
		return c.NoContent(400)
	}

	_, err = p.UpdatePayee(c.Request().Context(), id, signals.FormData.Name, signals.FormData.Description)
	if err != nil {
		slog.Error("payee.update.table: failed to update payee", "err", err)
		return c.String(500, "Internal Server Error")
	}

	slog.Debug("payee.update.table", "id", id)

	if err := hub.Hub.PatchSignals(c, defaultPayeeSignals); err != nil {
		slog.Warn("payee.update.table: failed to patch signals", "err", err)
	}

	if err := hub.Hub.Refresh(c); err != nil {
		slog.Warn("payee.update.table: failed to signal client", "err", err)
	}

	return c.NoContent(200)
}

func (p *Payees) Destroy(c echo.Context) error {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		slog.Warn("payee.destroy: invalid id", "id", idStr)
		return c.NoContent(400)
	}

	if err := p.DeletePayee(c.Request().Context(), id); err != nil {
		slog.Error("payee.destroy: failed to delete payee", "err", err)
		return c.String(500, "Internal Server Error")
	}

	slog.Debug("payee.destroy", "id", id)

	if err := hub.Hub.Redirect(c, "/payee"); err != nil {
		slog.Warn("payee.destroy: failed to redirect", "err", err)
	}

	return c.NoContent(200)
}
