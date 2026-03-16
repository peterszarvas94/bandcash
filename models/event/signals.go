package event

import "bandcash/internal/utils"

func eventIndexSignals(csrfToken string, query utils.TableQuery) map[string]any {
	return map[string]any{
		"csrf":            csrfToken,
		"tableQuery":      utils.TableQuerySignals(query),
		"dateRange":       map[string]any{"from": query.From, "to": query.To},
		"showCustomRange": query.DateMode == "custom" || query.From != "" || query.To != "",
		"mode":            "table",
		"formState":       "",
		"editingId":       0,
		"formData":        map[string]any{"title": "", "time": "", "description": "", "amount": 0},
		"eventFormState":  "",
		"errors":          map[string]any{"title": "", "time": "", "description": "", "amount": "", "memberId": "", "expense": ""},
	}
}

func eventShowSignals(data EventData, csrfToken string) map[string]any {
	wizardRows := make([]map[string]any, 0, len(data.WizardRows))
	wizardAmounts := make(map[string]int64, len(data.WizardRows))
	wizardExpenses := make(map[string]int64, len(data.WizardRows))
	wizardTotal := int64(0)
	for _, row := range data.WizardRows {
		wizardRows = append(wizardRows, map[string]any{
			"memberId":   row.MemberID,
			"memberName": row.MemberName,
			"included":   row.Included,
			"amount":     row.Amount,
			"expense":    row.Expense,
		})
		wizardAmounts[row.MemberID] = row.Amount
		wizardExpenses[row.MemberID] = row.Expense
		wizardTotal += row.Amount + row.Expense
	}

	return map[string]any{
		"csrf":                   csrfToken,
		"mode":                   "single",
		"tableQuery":             utils.TableQuerySignals(data.Query),
		"eventFormState":         "",
		"participantEditorMode":  data.EditorMode,
		"wizardError":            data.WizardError,
		"wizardEventAmount":      data.WizardEventAmount,
		"wizardSelectedMemberId": "",
		"wizardRows":             wizardRows,
		"wizardAmounts":          wizardAmounts,
		"wizardExpenses":         wizardExpenses,
		"wizardTotal":            wizardTotal,
		"wizardLeftover":         data.WizardEventAmount - wizardTotal,
		"eventFormData": map[string]any{
			"title":       data.Event.Title,
			"time":        data.Event.Time,
			"description": data.Event.Description,
			"amount":      data.Event.Amount,
		},
		"formState":   "",
		"editingId":   0,
		"calcPercent": 0,
		"formData": map[string]any{
			"memberId":   "",
			"memberName": "",
			"amount":     0,
			"expense":    0,
		},
		"errors": map[string]any{
			"title":       "",
			"time":        "",
			"description": "",
			"amount":      "",
			"memberId":    "",
			"expense":     "",
		},
	}
}
