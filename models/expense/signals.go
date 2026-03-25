package expense

import "bandcash/internal/utils"

func expenseIndexSignals(tabID, csrfToken string, query utils.TableQuery) map[string]any {
	return map[string]any{
		"tab_id":          tabID,
		"csrf":            csrfToken,
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

func expenseShowSignals(tabID string, data ExpenseData, csrfToken string) map[string]any {
	return map[string]any{
		"tab_id":    tabID,
		"csrf":      csrfToken,
		"mode":      "single",
		"formState": "",
		"formData": map[string]any{
			"title":       data.Expense.Title,
			"description": data.Expense.Description,
			"amount":      data.Expense.Amount,
			"date":        data.Expense.Date,
			"paid":        data.Expense.Paid == 1,
			"paidAt": func() string {
				if !data.Expense.PaidAt.Valid {
					return ""
				}
				return utils.FormatDateInput(data.Expense.PaidAt.String)
			}(),
		},
		"errors": map[string]any{"title": "", "description": "", "amount": "", "date": ""},
	}
}
