package expense

import "bandcash/internal/utils"

func expenseIndexSignals(csrfToken string, query utils.TableQuery) map[string]any {
	return map[string]any{
		"csrf":            csrfToken,
		"tableQuery":      utils.TableQuerySignals(query),
		"dateRange":       map[string]any{"from": query.From, "to": query.To},
		"showCustomRange": query.DateMode == "custom" || query.From != "" || query.To != "",
		"mode":            "table",
		"formState":       "",
		"eventFormState":  "",
		"editingId":       0,
		"formData":        map[string]any{"title": "", "description": "", "amount": 0, "date": ""},
		"errors":          map[string]any{"title": "", "description": "", "amount": "", "date": ""},
	}
}

func expenseShowSignals(data ExpenseData, csrfToken string) map[string]any {
	return map[string]any{
		"csrf":      csrfToken,
		"mode":      "single",
		"formState": "",
		"formData": map[string]any{
			"title":       data.Expense.Title,
			"description": data.Expense.Description,
			"amount":      data.Expense.Amount,
			"date":        data.Expense.Date,
		},
		"errors": map[string]any{"title": "", "description": "", "amount": "", "date": ""},
	}
}
