package expense

func expenseIndexSignals(csrfToken string) map[string]any {
	return map[string]any{
		"csrf":           csrfToken,
		"mode":           "",
		"formState":      "",
		"eventFormState": "",
		"editingId":      0,
		"formData":       map[string]any{"title": "", "description": "", "amount": 0, "date": ""},
		"errors":         map[string]any{"title": "", "description": "", "amount": "", "date": ""},
	}
}
