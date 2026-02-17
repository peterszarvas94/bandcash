package event

func eventIndexSignals() map[string]any {
	return map[string]any{
		"mode":           "",
		"formState":      "",
		"editingId":      0,
		"formData":       map[string]any{"title": "", "time": "", "description": "", "amount": 0},
		"eventFormState": "",
	}
}

func eventShowSignals(data EventData) map[string]any {
	return map[string]any{
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
		},
	}
}
