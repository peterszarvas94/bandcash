package member

import (
	"log/slog"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/starfederation/datastar-go/datastar"

	"bandcash/internal/db"
	"bandcash/internal/utils"
)

type memberParams struct {
	Name        string `json:"name" validate:"required,min=1,max=255"`
	Description string `json:"description" validate:"max=1000"`
}

type memberTableParams struct {
	FormData memberParams `json:"formData"`
	Mode     string       `json:"mode"`
}

type modeParams struct {
	Mode string `json:"mode"`
}

// Default signal state for resetting member forms on success
var (
	defaultMemberSignals = map[string]any{
		"mode":      "",
		"formState": "",
		"editingId": 0,
		"formData":  map[string]any{"name": "", "description": ""},
	}
	memberErrorFields = []string{"name", "description"}
)

func (p *Members) Index(c echo.Context) error {
	utils.EnsureClientID(c)

	data, err := p.GetIndexData(c.Request().Context())
	if err != nil {
		slog.Error("member.list: failed to get data", "err", err)
		return c.NoContent(500)
	}

	slog.Debug("member.index", "member_count", len(data.Members))
	return utils.RenderComponent(c, MemberIndex(data))
}

func (p *Members) Show(c echo.Context) error {
	utils.EnsureClientID(c)

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		slog.Warn("member.show: invalid id", "err", err)
		return c.NoContent(400)
	}

	data, err := p.GetShowData(c.Request().Context(), id)
	if err != nil {
		slog.Error("member.show: failed to get data", "err", err)
		return c.NoContent(500)
	}

	return utils.RenderComponent(c, MemberShow(data))
}

func (p *Members) Create(c echo.Context) error {
	var signals memberTableParams
	err := datastar.ReadSignals(c.Request(), &signals)
	if err != nil {
		slog.Warn("member.create.table: failed to read signals", "err", err)
		return c.NoContent(400)
	}

	// Validate
	if errs := utils.ValidateWithLocale(c.Request().Context(), signals.FormData); errs != nil {
		utils.SSEHub.PatchSignals(c, map[string]any{"errors": utils.WithErrors(memberErrorFields, errs)})
		return c.NoContent(422)
	}

	member, err := db.Qry.CreateMember(c.Request().Context(), db.CreateMemberParams{
		Name:        signals.FormData.Name,
		Description: signals.FormData.Description,
	})
	if err != nil {
		slog.Error("member.create.table: failed to create member", "err", err)
		return c.NoContent(500)
	}

	slog.Debug("member.create.table", "id", member.ID, "name", member.Name)

	utils.SSEHub.PatchSignals(c, defaultMemberSignals)
	data, err := p.GetIndexData(c.Request().Context())
	if err != nil {
		slog.Error("member.create.table: failed to get data", "err", err)
		return c.NoContent(500)
	}
	html, err := utils.RenderComponentString(c.Request().Context(), MemberIndex(data))
	if err != nil {
		slog.Error("member.create.table: failed to render", "err", err)
		return c.NoContent(500)
	}

	utils.SSEHub.PatchHTML(c, html)

	return c.NoContent(200)
}

func (p *Members) Update(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		slog.Warn("member.update: invalid id", "err", err)
		return c.NoContent(400)
	}

	var signals memberTableParams
	err = datastar.ReadSignals(c.Request(), &signals)
	if err != nil {
		slog.Warn("member.update: failed to read signals", "err", err)
		return c.NoContent(400)
	}

	// Validate
	if errs := utils.ValidateWithLocale(c.Request().Context(), signals.FormData); errs != nil {
		utils.SSEHub.PatchSignals(c, map[string]any{"errors": utils.WithErrors(memberErrorFields, errs)})
		return c.NoContent(422)
	}

	_, err = db.Qry.UpdateMember(c.Request().Context(), db.UpdateMemberParams{
		Name:        signals.FormData.Name,
		Description: signals.FormData.Description,
		ID:          int64(id),
	})
	if err != nil {
		slog.Error("member.update: failed to update member", "err", err)
		return c.NoContent(500)
	}

	slog.Debug("member.update", "id", id)

	if signals.Mode == "single" {
		err = utils.SSEHub.Redirect(c, "/member/"+strconv.Itoa(id))
		if err != nil {
			slog.Warn("member.update: failed to redirect", "err", err)
		}
		return c.NoContent(200)
	}

	utils.SSEHub.PatchSignals(c, defaultMemberSignals)
	data, err := p.GetIndexData(c.Request().Context())
	if err != nil {
		slog.Error("member.update: failed to get data", "err", err)
		return c.NoContent(500)
	}
	html, err := utils.RenderComponentString(c.Request().Context(), MemberIndex(data))
	if err != nil {
		slog.Error("member.update: failed to render", "err", err)
		return c.NoContent(500)
	}

	utils.SSEHub.PatchHTML(c, html)

	return c.NoContent(200)
}

func (p *Members) Destroy(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		slog.Warn("member.destroy: invalid id", "err", err)
		return c.NoContent(400)
	}

	var signals modeParams
	err = datastar.ReadSignals(c.Request(), &signals)
	if err != nil {
		slog.Warn("member.destroy: failed to read signals", "err", err)
		return c.NoContent(400)
	}

	err = db.Qry.DeleteMember(c.Request().Context(), int64(id))
	if err != nil {
		slog.Error("member.destroy: failed to delete member", "err", err)
		return c.NoContent(500)
	}

	slog.Debug("member.destroy", "id", id)

	if signals.Mode == "single" {
		err = utils.SSEHub.Redirect(c, "/member")
		if err != nil {
			slog.Warn("member.destroy: failed to redirect", "err", err)
		}
		return c.NoContent(200)
	}

	utils.SSEHub.PatchSignals(c, defaultMemberSignals)
	data, err := p.GetIndexData(c.Request().Context())
	if err != nil {
		slog.Error("member.destroy: failed to get data", "err", err)
		return c.NoContent(500)
	}
	html, err := utils.RenderComponentString(c.Request().Context(), MemberIndex(data))
	if err != nil {
		slog.Error("member.destroy: failed to render", "err", err)
		return c.NoContent(500)
	}

	utils.SSEHub.PatchHTML(c, html)

	return c.NoContent(200)
}
