package expense

import "bandcash/internal/utils"

func expenseIndexSignals(query utils.TableQuery) map[string]any {
	return map[string]any{
		"tableQuery":      utils.TableQuerySignals(query),
		"dateRange":       map[string]any{"from": query.From, "to": query.To},
		"showCustomRange": query.DateMode == "custom" || query.From != "" || query.To != "",
		"mode":            "table",
		"formState":       "",
		"eventFormState":  "",
		"summaryMode":     query.Summary,
		"editingId":       0,
		"paidAtDialog": map[string]any{
			"open":        false,
			"fetching":    false,
			"mode":        "table",
			"title":       "",
			"message":     "",
			"value":       "",
			"placeholder": "",
			"submitLabel": "",
			"cancelLabel": "",
			"url":         "",
			"triggerID":   "",
		},
		"formData": map[string]any{"title": "", "description": "", "amount": 0, "date": "", "paid": false, "paidAt": ""},
		"errors":   map[string]any{"title": "", "description": "", "amount": "", "date": ""},
	}
}

func expenseShowSignals(data ExpenseData) map[string]any {
	return map[string]any{
		"mode": "single",
		"paidAtDialog": map[string]any{
			"open":        data.PaidAtDialog.Open,
			"fetching":    data.PaidAtDialog.Fetching,
			"mode":        data.PaidAtDialog.Mode,
			"title":       data.PaidAtDialog.Title,
			"message":     data.PaidAtDialog.Message,
			"value":       data.PaidAtDialog.Value,
			"placeholder": data.PaidAtDialog.Placeholder,
			"submitLabel": data.PaidAtDialog.SubmitLabel,
			"cancelLabel": data.PaidAtDialog.CancelLabel,
			"url":         data.PaidAtDialog.URL,
			"triggerID":   data.PaidAtDialog.TriggerID,
		},
	}
}
