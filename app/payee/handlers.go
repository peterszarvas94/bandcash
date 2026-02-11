package payee

import (
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/starfederation/datastar-go/datastar"

	appmw "bandcash/internal/middleware"
	"bandcash/internal/utils"
)

type payeeParams struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (p *Payees) Index(c echo.Context) error {
	utils.EnsureClientID(c)

	log := appmw.Logger(c)
	data, err := p.GetIndexData(c.Request().Context())
	if err != nil {
		log.Error("payee.list: failed to get data", "err", err)
		return c.String(500, "Internal Server Error")
	}

	log.Debug("payee.index", "payee_count", len(data.(PayeesData).Payees))
	return p.tmpl.ExecuteTemplate(c.Response().Writer, "index", data)
}

func (p *Payees) New(c echo.Context) error {
	utils.EnsureClientID(c)
	return p.tmpl.ExecuteTemplate(c.Response().Writer, "new", PayeeData{Title: "New Payee"})
}

func (p *Payees) Show(c echo.Context) error {
	utils.EnsureClientID(c)
	log := appmw.Logger(c)

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.String(400, "Invalid ID")
	}

	payee, err := p.GetPayee(c.Request().Context(), id)
	if err != nil {
		log.Error("payee.show: failed to get payee", "err", err)
		return c.String(500, "Internal Server Error")
	}

	entries, err := p.GetEntries(c.Request().Context(), id)
	if err != nil {
		log.Error("payee.show: failed to get entries", "err", err)
		return c.String(500, "Internal Server Error")
	}

	return p.tmpl.ExecuteTemplate(c.Response().Writer, "show", PayeeData{
		Title:   payee.Name,
		Payee:   payee,
		Entries: entries,
	})
}

func (p *Payees) Edit(c echo.Context) error {
	utils.EnsureClientID(c)
	log := appmw.Logger(c)

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.String(400, "Invalid ID")
	}

	payee, err := p.GetPayee(c.Request().Context(), id)
	if err != nil {
		log.Error("payee.edit: failed to get payee", "err", err)
		return c.String(500, "Internal Server Error")
	}

	return p.tmpl.ExecuteTemplate(c.Response().Writer, "edit", PayeeData{Title: "Edit Payee", Payee: payee})
}

func (p *Payees) Create(c echo.Context) error {
	log := appmw.Logger(c)

	var signals payeeParams
	if err := datastar.ReadSignals(c.Request(), &signals); err != nil {
		log.Warn("payee.create: failed to read signals", "err", err)
		return c.NoContent(400)
	}

	if signals.Name == "" {
		log.Debug("payee.create: empty name")
		return c.NoContent(200)
	}

	payee, err := p.CreatePayee(c.Request().Context(), signals.Name, signals.Description)
	if err != nil {
		log.Error("payee.create: failed to create payee", "err", err)
		return c.String(500, "Internal Server Error")
	}

	log.Debug("payee.create", "id", payee.ID, "name", payee.Name)
	return c.Redirect(303, "/payee")
}

func (p *Payees) Update(c echo.Context) error {
	log := appmw.Logger(c)

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.String(400, "Invalid ID")
	}

	var signals payeeParams
	if err := datastar.ReadSignals(c.Request(), &signals); err != nil {
		log.Warn("payee.update: failed to read signals", "err", err)
		return c.NoContent(400)
	}

	_, err = p.UpdatePayee(c.Request().Context(), id, signals.Name, signals.Description)
	if err != nil {
		log.Error("payee.update: failed to update payee", "err", err)
		return c.String(500, "Internal Server Error")
	}

	log.Debug("payee.update", "id", id)
	return c.Redirect(303, "/payee/"+strconv.Itoa(id))
}

func (p *Payees) Destroy(c echo.Context) error {
	log := appmw.Logger(c)

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Warn("payee.destroy: invalid id", "id", idStr)
		return c.NoContent(400)
	}

	if err := p.DeletePayee(c.Request().Context(), id); err != nil {
		log.Error("payee.destroy: failed to delete payee", "err", err)
		return c.String(500, "Internal Server Error")
	}

	log.Debug("payee.destroy", "id", id)
	return c.Redirect(303, "/payee")
}
