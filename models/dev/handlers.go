package dev

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/starfederation/datastar-go/datastar"

	"bandcash/internal/db"
	"bandcash/internal/email"
	"bandcash/internal/utils"
	"bandcash/models/event"
	"bandcash/models/expense"
	"bandcash/models/member"
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

func (h *DevNotifications) TestTableQuery(c echo.Context) error {
	model := c.Param("model")
	groupID, err := h.resolveTableQueryGroupID(c, model)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"error": err.Error(),
		})
	}

	var queryable utils.Queryable
	var data any
	var pager utils.TablePagination
	var errData error
	raw := rawTableQuery(c)
	strict := c.QueryParam("strict") == "1"

	switch model {
	case "events":
		eventsModel := event.New()
		queryable = eventsModel
		parsedResult := utils.ParseTableQueryWithResult(c, queryable)
		parsed := parsedResult.Query
		if strict && len(parsedResult.Rejected) > 0 {
			return c.JSON(http.StatusUnprocessableEntity, map[string]any{
				"model":    model,
				"groupId":  groupID,
				"raw":      raw,
				"parsed":   parsed,
				"rejected": parsedResult.Rejected,
				"error":    "invalid query params",
			})
		}
		indexData, getErr := eventsModel.GetIndexData(c.Request().Context(), groupID, parsed)
		if getErr != nil {
			errData = getErr
			break
		}
		data = indexData.Events
		pager = indexData.Pager
		return c.JSON(http.StatusOK, map[string]any{
			"model":    model,
			"groupId":  groupID,
			"raw":      raw,
			"parsed":   parsed,
			"rejected": parsedResult.Rejected,
			"offset":   parsed.Offset(),
			"pager":    pager,
			"rowCount": len(indexData.Events),
			"data":     data,
		})
	case "members":
		membersModel := member.New()
		queryable = membersModel
		parsedResult := utils.ParseTableQueryWithResult(c, queryable)
		parsed := parsedResult.Query
		if strict && len(parsedResult.Rejected) > 0 {
			return c.JSON(http.StatusUnprocessableEntity, map[string]any{
				"model":    model,
				"groupId":  groupID,
				"raw":      raw,
				"parsed":   parsed,
				"rejected": parsedResult.Rejected,
				"error":    "invalid query params",
			})
		}
		indexData, getErr := membersModel.GetIndexData(c.Request().Context(), groupID, parsed)
		if getErr != nil {
			errData = getErr
			break
		}
		data = indexData.Members
		pager = indexData.Pager
		return c.JSON(http.StatusOK, map[string]any{
			"model":    model,
			"groupId":  groupID,
			"raw":      raw,
			"parsed":   parsed,
			"rejected": parsedResult.Rejected,
			"offset":   parsed.Offset(),
			"pager":    pager,
			"rowCount": len(indexData.Members),
			"data":     data,
		})
	case "expenses":
		expensesModel := expense.New()
		queryable = expensesModel
		parsedResult := utils.ParseTableQueryWithResult(c, queryable)
		parsed := parsedResult.Query
		if strict && len(parsedResult.Rejected) > 0 {
			return c.JSON(http.StatusUnprocessableEntity, map[string]any{
				"model":    model,
				"groupId":  groupID,
				"raw":      raw,
				"parsed":   parsed,
				"rejected": parsedResult.Rejected,
				"error":    "invalid query params",
			})
		}
		indexData, getErr := expensesModel.GetIndexData(c.Request().Context(), groupID, parsed)
		if getErr != nil {
			errData = getErr
			break
		}
		data = indexData.Expenses
		pager = indexData.Pager
		return c.JSON(http.StatusOK, map[string]any{
			"model":    model,
			"groupId":  groupID,
			"raw":      raw,
			"parsed":   parsed,
			"rejected": parsedResult.Rejected,
			"offset":   parsed.Offset(),
			"pager":    pager,
			"rowCount": len(indexData.Expenses),
			"data":     data,
		})
	default:
		return c.JSON(http.StatusBadRequest, map[string]any{
			"error": "unknown model",
		})
	}

	if errData != nil {
		return c.JSON(http.StatusInternalServerError, map[string]any{
			"error":   errData.Error(),
			"model":   model,
			"groupId": groupID,
		})
	}

	return c.JSON(http.StatusInternalServerError, map[string]any{
		"error": "unknown table query failure",
	})
}

func rawTableQuery(c echo.Context) map[string]any {
	return map[string]any{
		"q":        c.QueryParam("q"),
		"sort":     c.QueryParam("sort"),
		"dir":      c.QueryParam("dir"),
		"page":     c.QueryParam("page"),
		"pageSize": c.QueryParam("pageSize"),
		"groupId":  c.QueryParam("groupId"),
	}
}

func (h *DevNotifications) resolveTableQueryGroupID(c echo.Context, model string) (string, error) {
	if groupID := strings.TrimSpace(c.QueryParam("groupId")); groupID != "" {
		if !utils.IsValidID(groupID, "grp") {
			return "", fmt.Errorf("invalid groupId")
		}
		_, err := db.Qry.GetGroupByID(c.Request().Context(), groupID)
		if err != nil {
			if err == sql.ErrNoRows {
				return "", fmt.Errorf("group not found")
			}
			return "", err
		}
		return groupID, nil
	}

	groups, err := db.Qry.ListRecentGroups(c.Request().Context(), 50)
	if err != nil {
		return "", err
	}
	if len(groups) == 0 {
		return "", fmt.Errorf("no groups found; add ?groupId=grp_xxx")
	}

	ctx := c.Request().Context()
	for _, group := range groups {
		var count int64
		switch model {
		case "events":
			count, err = db.Qry.CountEventsFiltered(ctx, db.CountEventsFilteredParams{
				GroupID: group.ID,
				Search:  "",
			})
		case "members":
			count, err = db.Qry.CountMembersFiltered(ctx, db.CountMembersFilteredParams{
				GroupID: group.ID,
				Search:  "",
			})
		case "expenses":
			count, err = db.Qry.CountExpensesFiltered(ctx, db.CountExpensesFilteredParams{
				GroupID: group.ID,
				Search:  "",
			})
		default:
			return "", fmt.Errorf("unknown model")
		}
		if err != nil {
			return "", err
		}
		if count > 0 {
			return group.ID, nil
		}
	}

	return groups[0].ID, nil
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
