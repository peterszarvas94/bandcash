package member

import (
	"log/slog"
	"net/http"

	ctxi18n "github.com/invopop/ctxi18n/i18n"
	"github.com/labstack/echo/v4"
	"github.com/starfederation/datastar-go/datastar"

	"bandcash/internal/db"
	"bandcash/internal/middleware"
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
		"editingId": "",
		"formData":  map[string]any{"name": "", "description": ""},
		"errors":    map[string]any{"name": "", "description": ""},
	}
	memberErrorFields = []string{"name", "description"}
)

func getGroupID(c echo.Context) string {
	return middleware.GetGroupID(c)
}

func getUserEmail(c echo.Context) string {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return ""
	}
	user, err := db.Qry.GetUserByID(c.Request().Context(), userID)
	if err != nil {
		return ""
	}
	return user.Email
}

func (p *Members) Index(c echo.Context) error {
	utils.EnsureClientID(c)
	groupID := getGroupID(c)
	userEmail := getUserEmail(c)

	data, err := p.GetIndexData(c.Request().Context(), groupID)
	if err != nil {
		slog.Error("member.list: failed to get data", "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}
	data.IsAdmin = middleware.IsAdmin(c)
	data.UserEmail = userEmail

	slog.Debug("member.index", "member_count", len(data.Members))
	return utils.RenderComponent(c, MemberIndex(data))
}

func (p *Members) Show(c echo.Context) error {
	utils.EnsureClientID(c)
	groupID := getGroupID(c)
	userEmail := getUserEmail(c)

	id := c.Param("id")
	if !utils.IsValidID(id, utils.PrefixMember) {
		slog.Warn("member.show: invalid id")
		return c.NoContent(http.StatusBadRequest)
	}

	data, err := p.GetShowData(c.Request().Context(), groupID, id)
	if err != nil {
		slog.Error("member.show: failed to get data", "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}
	data.IsAdmin = middleware.IsAdmin(c)
	data.UserEmail = userEmail

	return utils.RenderComponent(c, MemberShow(data))
}

func (p *Members) Create(c echo.Context) error {
	groupID := getGroupID(c)
	userEmail := getUserEmail(c)

	var signals memberTableParams
	err := datastar.ReadSignals(c.Request(), &signals)
	if err != nil {
		slog.Warn("member.create.table: failed to read signals", "err", err)
		return c.NoContent(http.StatusBadRequest)
	}
	signals.FormData.Name = utils.NormalizeText(signals.FormData.Name)
	signals.FormData.Description = utils.NormalizeText(signals.FormData.Description)

	// Validate
	if errs := utils.ValidateWithLocale(c.Request().Context(), signals.FormData); errs != nil {
		utils.SSEHub.PatchSignals(c, map[string]any{"errors": utils.WithErrors(memberErrorFields, errs)})
		return c.NoContent(http.StatusUnprocessableEntity)
	}

	member, err := db.Qry.CreateMember(c.Request().Context(), db.CreateMemberParams{
		ID:          utils.GenerateID(utils.PrefixMember),
		GroupID:     groupID,
		Name:        signals.FormData.Name,
		Description: signals.FormData.Description,
	})
	if err != nil {
		slog.Error("member.create.table: failed to create member", "err", err)
		utils.Notify(c, "error", ctxi18n.T(c.Request().Context(), "members.notifications.create_failed"))
		return c.NoContent(http.StatusInternalServerError)
	}

	slog.Debug("member.create.table", "id", member.ID, "name", member.Name)
	utils.Notify(c, "success", ctxi18n.T(c.Request().Context(), "members.notifications.created"))

	utils.SSEHub.PatchSignals(c, defaultMemberSignals)
	data, err := p.GetIndexData(c.Request().Context(), groupID)
	if err != nil {
		slog.Error("member.create.table: failed to get data", "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}
	data.IsAdmin = middleware.IsAdmin(c)
	data.UserEmail = userEmail
	html, err := utils.RenderComponentStringFor(c, MemberIndex(data))
	if err != nil {
		slog.Error("member.create.table: failed to render", "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	utils.SSEHub.PatchHTML(c, html)

	return c.NoContent(http.StatusOK)
}

func (p *Members) Update(c echo.Context) error {
	groupID := getGroupID(c)
	userEmail := getUserEmail(c)

	id := c.Param("id")
	if !utils.IsValidID(id, utils.PrefixMember) {
		slog.Warn("member.update: invalid id")
		return c.NoContent(http.StatusBadRequest)
	}

	var signals memberTableParams
	err := datastar.ReadSignals(c.Request(), &signals)
	if err != nil {
		slog.Warn("member.update: failed to read signals", "err", err)
		return c.NoContent(http.StatusBadRequest)
	}
	signals.FormData.Name = utils.NormalizeText(signals.FormData.Name)
	signals.FormData.Description = utils.NormalizeText(signals.FormData.Description)

	// Validate
	if errs := utils.ValidateWithLocale(c.Request().Context(), signals.FormData); errs != nil {
		utils.SSEHub.PatchSignals(c, map[string]any{"errors": utils.WithErrors(memberErrorFields, errs)})
		return c.NoContent(http.StatusUnprocessableEntity)
	}

	_, err = db.Qry.UpdateMember(c.Request().Context(), db.UpdateMemberParams{
		Name:        signals.FormData.Name,
		Description: signals.FormData.Description,
		ID:          id,
		GroupID:     groupID,
	})
	if err != nil {
		slog.Error("member.update: failed to update member", "err", err)
		utils.Notify(c, "error", ctxi18n.T(c.Request().Context(), "members.notifications.update_failed"))
		return c.NoContent(http.StatusInternalServerError)
	}

	slog.Debug("member.update", "id", id)
	utils.Notify(c, "success", ctxi18n.T(c.Request().Context(), "members.notifications.updated"))

	if signals.Mode == "single" {
		err = utils.SSEHub.Redirect(c, "/groups/"+groupID+"/members/"+id)
		if err != nil {
			slog.Warn("member.update: failed to redirect", "err", err)
		}
		return c.NoContent(http.StatusOK)
	}

	utils.SSEHub.PatchSignals(c, defaultMemberSignals)
	data, err := p.GetIndexData(c.Request().Context(), groupID)
	if err != nil {
		slog.Error("member.update: failed to get data", "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}
	data.IsAdmin = middleware.IsAdmin(c)
	data.UserEmail = userEmail
	html, err := utils.RenderComponentStringFor(c, MemberIndex(data))
	if err != nil {
		slog.Error("member.update: failed to render", "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	utils.SSEHub.PatchHTML(c, html)

	return c.NoContent(http.StatusOK)
}

func (p *Members) Destroy(c echo.Context) error {
	groupID := getGroupID(c)
	userEmail := getUserEmail(c)

	id := c.Param("id")
	if !utils.IsValidID(id, utils.PrefixMember) {
		slog.Warn("member.destroy: invalid id")
		return c.NoContent(http.StatusBadRequest)
	}

	var signals modeParams
	err := datastar.ReadSignals(c.Request(), &signals)
	if err != nil {
		slog.Warn("member.destroy: failed to read signals", "err", err)
		return c.NoContent(http.StatusBadRequest)
	}

	err = db.Qry.DeleteMember(c.Request().Context(), db.DeleteMemberParams{
		ID:      id,
		GroupID: groupID,
	})
	if err != nil {
		slog.Error("member.destroy: failed to delete member", "err", err)
		utils.Notify(c, "error", ctxi18n.T(c.Request().Context(), "members.notifications.delete_failed"))
		return c.NoContent(http.StatusInternalServerError)
	}

	slog.Debug("member.destroy", "id", id)
	utils.Notify(c, "success", ctxi18n.T(c.Request().Context(), "members.notifications.deleted"))

	if signals.Mode == "single" {
		err = utils.SSEHub.Redirect(c, "/groups/"+groupID+"/members")
		if err != nil {
			slog.Warn("member.destroy: failed to redirect", "err", err)
		}
		return c.NoContent(http.StatusOK)
	}

	utils.SSEHub.PatchSignals(c, defaultMemberSignals)
	data, err := p.GetIndexData(c.Request().Context(), groupID)
	if err != nil {
		slog.Error("member.destroy: failed to get data", "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}
	data.IsAdmin = middleware.IsAdmin(c)
	data.UserEmail = userEmail
	html, err := utils.RenderComponentStringFor(c, MemberIndex(data))
	if err != nil {
		slog.Error("member.destroy: failed to render", "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	utils.SSEHub.PatchHTML(c, html)

	return c.NoContent(http.StatusOK)
}
