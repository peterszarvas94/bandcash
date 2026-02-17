package entry

func entryIndexSignals() map[string]any {
	return map[string]any{
		"mode":           "",
		"formState":      "",
		"editingId":      0,
		"formData":       map[string]any{"title": "", "time": "", "description": "", "amount": 0},
		"entryFormState": "",
	}
}

func entryShowSignals(data EntryData) map[string]any {
	return map[string]any{
		"mode":           "single",
		"entryFormState": "",
		"entryFormData": map[string]any{
			"title":       data.Entry.Title,
			"time":        data.Entry.Time,
			"description": data.Entry.Description,
			"amount":      data.Entry.Amount,
		},
		"formState":   "",
		"editingId":   0,
		"calcPercent": 0,
		"formData": map[string]any{
			"payeeId":   "",
			"payeeName": "",
			"amount":    0,
			"expense":   0,
		},
		"errors": map[string]any{
			"title":       "",
			"time":        "",
			"description": "",
			"amount":      "",
		},
	}
}
