package event

func eventIndexSignals(csrfToken, groupName string) map[string]any {
	return map[string]any{
		"csrf":           csrfToken,
		"mode":           "",
		"formState":      "",
		"groupFormState": "",
		"editingId":      0,
		"formData":       map[string]any{"title": "", "time": "", "description": "", "amount": 0},
		"groupFormData":  map[string]any{"name": groupName},
		"groupErrors":    map[string]any{"name": ""},
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
