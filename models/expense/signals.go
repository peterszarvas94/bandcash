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
		"formData":        map[string]any{"title": "", "description": "", "amount": 0, "date": "", "paid": false, "paidAt": ""},
		"errors":          map[string]any{"title": "", "description": "", "amount": "", "date": ""},
	}
}

func expenseShowSignals(_ ExpenseData) map[string]any {
	return map[string]any{
		"mode": "single",
	}
}
