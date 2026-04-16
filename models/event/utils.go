package event

import (
	"database/sql"
	"fmt"
	"log/slog"
	"net/url"
	"strings"

	ctxi18n "github.com/invopop/ctxi18n/i18n"
	"github.com/labstack/echo/v4"

	"bandcash/internal/db"
	"bandcash/internal/utils"
)

func normalizeCacheKeyPart(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "all"
	}
	return url.QueryEscape(trimmed)
}

func eventsCachePrefix(groupID string) string {
	return fmt.Sprintf("events_group_%s_", normalizeCacheKeyPart(groupID))
}

func EventsFilterKey(groupID, search, year, from, to, sort, dir string) string {
	return fmt.Sprintf("%ssearch_%s_year_%s_from_%s_to_%s_sort_%s_dir_%s",
		eventsCachePrefix(groupID),
		normalizeCacheKeyPart(search),
		normalizeCacheKeyPart(year),
		normalizeCacheKeyPart(from),
		normalizeCacheKeyPart(to),
		normalizeCacheKeyPart(sort),
		normalizeCacheKeyPart(dir),
	)
}

func eventDateValue(event db.Event) string {
	date := strings.TrimSpace(event.Date)
	if date != "" {
		return date
	}
	return utils.FormatDateInput(event.Time)
}

func eventTimeValue(event db.Event) string {
	eventTime := strings.TrimSpace(event.EventTime)
	if eventTime != "" {
		return eventTime
	}
	trimmed := strings.TrimSpace(event.Time)
	if len(trimmed) >= 16 {
		return trimmed[11:16]
	}
	return ""
}

func eventDateTimeValue(event db.Event) string {
	date := eventDateValue(event)
	eventTime := eventTimeValue(event)
	if date == "" {
		return strings.TrimSpace(event.Time)
	}
	if eventTime == "" {
		return date
	}
	return date + "T" + eventTime
}

type staticTableQueryable struct {
	spec utils.TableQuerySpec
}

func (s staticTableQueryable) TableQuerySpec() utils.TableQuerySpec {
	return s.spec
}

func parseParticipantTableQuery(c echo.Context) utils.TableQuery {
	query := utils.ParseTableQuery(c, staticTableQueryable{spec: ParticipantTableQuerySpec()})
	query.Page = 1
	query.Search = ""
	query.PageSize = utils.DefaultTablePageSize
	return query
}

func normalizePaidAtInput(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return ""
	}

	formatted := utils.FormatDateInput(trimmed)
	if formatted != "" {
		return formatted
	}

	return trimmed
}

func paidAtArg(isPaid bool, paidAt string) sql.NullString {
	if !isPaid {
		return sql.NullString{}
	}

	normalized := normalizePaidAtInput(paidAt)
	if normalized == "" {
		return sql.NullString{String: "", Valid: true}
	}

	return sql.NullString{String: normalized, Valid: true}
}

func applyEventIndexTableByRole(data *EventsData, isAdmin bool) {
	data.IsAdmin = isAdmin
	if !isAdmin {
		data.EventsTable.ActionsWidthRem = 0
	}
}

func applyEventShowTableByRole(data *EventData, isAdmin bool) {
	data.IsAdmin = isAdmin
	if !isAdmin {
		data.ParticipantsTable.ActionsWidthRem = 0
	}
}

func mergeWizardRows(base []ParticipantWizardRow, allMembers []db.Member, incoming []participantBulkRowData, wizardMemberIDs map[string]string, wizardAmounts map[string]int64, wizardExpenses map[string]int64, wizardNotes map[string]string, wizardPaids map[string]bool, wizardPaidAts map[string]string) []ParticipantWizardRow {
	if len(incoming) == 0 {
		for i := range base {
			if wizardMemberIDs != nil {
				if memberID, ok := wizardMemberIDs[base[i].RowID]; ok {
					base[i].MemberID = strings.TrimSpace(memberID)
				}
			}
			if wizardAmounts != nil {
				if amount, ok := wizardAmounts[base[i].RowID]; ok {
					base[i].Amount = amount
				}
			}
			if wizardExpenses != nil {
				if expense, ok := wizardExpenses[base[i].RowID]; ok {
					base[i].Expense = expense
				}
			}
			if wizardNotes != nil {
				if note, ok := wizardNotes[base[i].RowID]; ok {
					base[i].Note = strings.TrimSpace(note)
				}
			}
			if wizardPaids != nil {
				if paid, ok := wizardPaids[base[i].RowID]; ok {
					base[i].Paid = paid
				}
			}
			if wizardPaidAts != nil {
				if paidAt, ok := wizardPaidAts[base[i].RowID]; ok {
					base[i].PaidAt = normalizePaidAtInput(paidAt)
				}
			}
		}

		memberNameByID := make(map[string]string, len(allMembers))
		for _, member := range allMembers {
			memberNameByID[member.ID] = member.Name
		}
		for i := range base {
			if base[i].MemberID == "" {
				base[i].MemberName = ""
				continue
			}
			base[i].MemberName = memberNameByID[base[i].MemberID]
		}

		return base
	}

	memberNameByID := make(map[string]string, len(allMembers))
	for _, member := range allMembers {
		memberNameByID[member.ID] = member.Name
	}

	merged := make([]ParticipantWizardRow, 0, len(incoming))
	for _, incomingRow := range incoming {
		rowID := strings.TrimSpace(incomingRow.RowID)
		if rowID == "" {
			rowID = utils.GenerateID(utils.PrefixParticipant)
		}

		memberID := strings.TrimSpace(incomingRow.MemberID)
		if wizardMemberIDs != nil {
			if value, ok := wizardMemberIDs[rowID]; ok {
				memberID = strings.TrimSpace(value)
			}
		}

		amount := incomingRow.Amount
		if wizardAmounts != nil {
			if value, ok := wizardAmounts[rowID]; ok {
				amount = value
			}
		}

		expense := incomingRow.Expense
		if wizardExpenses != nil {
			if value, ok := wizardExpenses[rowID]; ok {
				expense = value
			}
		}

		note := strings.TrimSpace(incomingRow.Note)
		if wizardNotes != nil {
			if value, ok := wizardNotes[rowID]; ok {
				note = strings.TrimSpace(value)
			}
		}

		paid := incomingRow.Paid
		if wizardPaids != nil {
			if value, ok := wizardPaids[rowID]; ok {
				paid = value
			}
		}

		paidAt := normalizePaidAtInput(incomingRow.PaidAt)
		if wizardPaidAts != nil {
			if value, ok := wizardPaidAts[rowID]; ok {
				paidAt = normalizePaidAtInput(value)
			}
		}

		memberName := ""
		if memberID != "" {
			memberName = memberNameByID[memberID]
		}

		merged = append(merged, ParticipantWizardRow{
			RowID:      rowID,
			MemberID:   memberID,
			MemberName: memberName,
			Included:   incomingRow.Included,
			Amount:     amount,
			Expense:    expense,
			Note:       note,
			Paid:       paid,
			PaidAt:     paidAt,
		})
	}

	return merged
}

func patchWizardError(c echo.Context, wizard participantWizardSignals, message string, rowID string) {
	_ = rowID

	if err := utils.SSEHub.PatchSignals(c, map[string]any{
		"wizard": map[string]any{
			"eventAmount": wizard.EventAmount,
			"rows":        wizard.Rows,
			"rowErrors":   map[string]string{},
			"memberIds":   wizard.MemberIDs,
			"amounts":     wizard.Amounts,
			"expenses":    wizard.Expenses,
			"notes":       wizard.Notes,
			"paids":       wizard.Paids,
			"paidAts":     wizard.PaidAts,
			"total":       wizard.Total,
			"balance":     wizard.Balance,
			"error":       message,
		},
		"errors": map[string]any{
			"memberId": "",
		},
	}); err != nil {
		slog.Warn("event.wizard: failed to patch error signals", "err", err)
	}
}

func patchEventShow(c echo.Context, groupID, eventID string, query utils.TableQuery, editorMode string, eventForm eventData, wizardEventAmount int64, wizardRows []participantBulkRowData, wizardMemberIDs map[string]string, wizardAmounts map[string]int64, wizardExpenses map[string]int64, wizardNotes map[string]string, wizardPaids map[string]bool, wizardPaidAts map[string]string, wizardError string) error {
	data, err := GetShowData(c.Request().Context(), groupID, eventID, query)
	if err != nil {
		return err
	}

	applyEventShowTableByRole(&data, utils.IsAdmin(c))
	data.IsAuthenticated = true
	data.IsSuperAdmin = utils.IsSuperadmin(c)

	if editorMode != "" {
		data.EditorMode = editorMode
	}
	if data.EditorMode == "edit" {
		if len(data.Breadcrumbs) > 0 {
			data.Breadcrumbs[len(data.Breadcrumbs)-1].Href = "/groups/" + groupID + "/events/" + eventID
		}
		data.Breadcrumbs = append(data.Breadcrumbs, utils.Crumb{Label: ctxi18n.T(c.Request().Context(), "events.edit")})
	}

	if eventForm.Title != "" || eventForm.Date != "" || eventForm.Time != "" || eventForm.Place != "" || eventForm.Amount > 0 {
		data.Event.Title = eventForm.Title
		data.Event.Date = eventForm.Date
		data.Event.EventTime = eventForm.Time
		data.Event.Time = eventForm.Date + "T" + eventForm.Time
		data.Event.Place = eventForm.Place
		data.Event.Description = eventForm.Description
		data.Event.Amount = eventForm.Amount
		if eventForm.Paid {
			data.Event.Paid = 1
		} else {
			data.Event.Paid = 0
		}
		if eventForm.Paid {
			if eventForm.PaidAt != "" {
				data.Event.PaidAt = sql.NullString{String: eventForm.PaidAt, Valid: true}
			}
		} else {
			data.Event.PaidAt = sql.NullString{}
		}
	}

	if wizardEventAmount > 0 {
		data.WizardEventAmount = wizardEventAmount
	}

	if wizardRows != nil {
		if len(wizardRows) == 0 {
			data.WizardRows = []ParticipantWizardRow{}
		} else {
			data.WizardRows = mergeWizardRows(data.WizardRows, data.AllMembers, wizardRows, wizardMemberIDs, wizardAmounts, wizardExpenses, wizardNotes, wizardPaids, wizardPaidAts)
		}
	} else if wizardMemberIDs != nil || wizardAmounts != nil || wizardExpenses != nil || wizardNotes != nil || wizardPaids != nil || wizardPaidAts != nil {
		data.WizardRows = mergeWizardRows(data.WizardRows, data.AllMembers, nil, wizardMemberIDs, wizardAmounts, wizardExpenses, wizardNotes, wizardPaids, wizardPaidAts)
	}

	data.WizardError = wizardError
	data.Signals = eventShowSignals(data)

	var html string
	if data.EditorMode == "edit" {
		html, err = utils.RenderHTMLForRequest(c, EventEditPage(data))
	} else {
		html, err = utils.RenderHTMLForRequest(c, EventShowPage(data))
	}
	if err != nil {
		return err
	}

	if err := utils.SSEHub.PatchHTML(c, html); err != nil {
		return err
	}
	if err := utils.SSEHub.PatchSignals(c, eventShowSignals(data)); err != nil {
		return err
	}
	return nil
}
