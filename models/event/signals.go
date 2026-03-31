package event

import "bandcash/internal/utils"

func eventIndexSignals(query utils.TableQuery) map[string]any {
	return map[string]any{
		"tableQuery":      utils.TableQuerySignals(query),
		"dateRange":       map[string]any{"from": query.From, "to": query.To},
		"showCustomRange": query.DateMode == "custom" || query.From != "" || query.To != "",
		"mode":            "table",
		"formState":       "",
		"editingId":       0,
		"formData":        map[string]any{"title": "", "time": "", "place": "", "description": "", "amount": 0, "paid": false, "paidAt": ""},
		"eventFormState":  "",
		"summaryMode":     query.Summary,
		"_fetching":       false,
		"errors":          map[string]any{"title": "", "time": "", "place": "", "description": "", "amount": "", "memberId": "", "expense": ""},
	}
}

func eventShowSignals(data EventData) map[string]any {
	wizardRows := make([]map[string]any, 0, len(data.WizardRows))
	wizardMemberIDs := make(map[string]string, len(data.WizardRows))
	wizardAmounts := make(map[string]int64, len(data.WizardRows))
	wizardExpenses := make(map[string]int64, len(data.WizardRows))
	wizardNotes := make(map[string]string, len(data.WizardRows))
	wizardPaids := make(map[string]bool, len(data.WizardRows))
	wizardPaidAts := make(map[string]string, len(data.WizardRows))
	wizardTotal := int64(0)
	for _, row := range data.WizardRows {
		rowID := row.RowID
		if rowID == "" {
			rowID = row.MemberID
		}
		wizardRows = append(wizardRows, map[string]any{
			"rowId":      rowID,
			"memberId":   row.MemberID,
			"memberName": row.MemberName,
			"included":   row.Included,
			"amount":     row.Amount,
			"expense":    row.Expense,
			"note":       row.Note,
			"paid":       row.Paid,
			"paidAt":     row.PaidAt,
		})
		wizardMemberIDs[rowID] = row.MemberID
		wizardAmounts[rowID] = row.Amount
		wizardExpenses[rowID] = row.Expense
		wizardNotes[rowID] = row.Note
		wizardPaids[rowID] = row.Paid
		wizardPaidAts[rowID] = row.PaidAt
		wizardTotal += row.Amount + row.Expense
	}

	noteExpanded := make(map[string]bool, len(data.Participants))
	for _, participant := range data.Participants {
		noteExpanded[participant.ID] = false
	}

	return map[string]any{
		"mode":                  "single",
		"draftRowsAction":       "",
		"draftRowsRowId":        "",
		"tableQuery":            utils.TableQuerySignals(data.Query),
		"eventFormState":        "",
		"participantEditorMode": data.EditorMode,
		"wizard": map[string]any{
			"error":       data.WizardError,
			"eventAmount": data.WizardEventAmount,
			"rows":        wizardRows,
			"memberIds":   wizardMemberIDs,
			"amounts":     wizardAmounts,
			"expenses":    wizardExpenses,
			"notes":       wizardNotes,
			"paids":       wizardPaids,
			"paidAts":     wizardPaidAts,
			"total":       wizardTotal,
			"leftover":    data.WizardEventAmount - wizardTotal,
		},
		"eventFormData": map[string]any{
			"title":       data.Event.Title,
			"time":        data.Event.Time,
			"place":       data.Event.Place,
			"description": data.Event.Description,
			"amount":      data.Event.Amount,
			"paid":        data.Event.Paid == 1,
			"paidAt": func() string {
				if !data.Event.PaidAt.Valid {
					return ""
				}
				return utils.FormatDateInput(data.Event.PaidAt.String)
			}(),
		},
		"formState":    "",
		"editingId":    0,
		"calcPercent":  0,
		"summaryMode":  data.Query.Summary,
		"_fetching":    false,
		"noteExpanded": noteExpanded,
		"formData": map[string]any{
			"memberId":   "",
			"memberName": "",
			"amount":     0,
			"expense":    0,
		},
		"errors": map[string]any{
			"title":       "",
			"time":        "",
			"place":       "",
			"description": "",
			"amount":      "",
			"memberId":    "",
			"expense":     "",
		},
	}
}
