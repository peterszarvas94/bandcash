package expense

import "bandcash/internal/utils"

func expenseIndexSignals(csrfToken string, query utils.TableQuery) map[string]any {
	return map[string]any{
		"csrf":           csrfToken,
		"tableQuery":     utils.TableQuerySignals(query),
		"dateRange":      map[string]any{"from": query.From, "to": query.To},
		"mode":           "table",
		"formState":      "",
		"eventFormState": "",
		"editingId":      0,
		"formData":       map[string]any{"title": "", "description": "", "amount": 0, "date": ""},
		"errors":         map[string]any{"title": "", "description": "", "amount": "", "date": ""},
	}
}
