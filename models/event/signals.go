package event

func eventIndexSignals(csrfToken string, query map[string]any) map[string]any {
	return map[string]any{
		"csrf":           csrfToken,
		"tableQuery":     query,
		"mode":           "",
		"formState":      "",
		"editingId":      0,
		"formData":       map[string]any{"title": "", "time": "", "description": "", "amount": 0},
		"eventFormState": "",
		"errors":         map[string]any{"title": "", "time": "", "description": "", "amount": "", "memberId": "", "expense": ""},
	}
}

func eventShowSignals(data EventData, csrfToken string) map[string]any {
	return map[string]any{
		"csrf":           csrfToken,
		"mode":           "single",
		"eventFormState": "",
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
