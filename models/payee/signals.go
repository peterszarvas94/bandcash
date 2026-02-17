package payee

func payeeIndexSignals() map[string]any {
	return map[string]any{
		"mode":      "",
		"formState": "",
		"editingId": 0,
		"formData":  map[string]any{"name": "", "description": ""},
	}
}

func payeeShowSignals(data PayeeData) map[string]any {
	return map[string]any{
		"mode":      "single",
		"formState": "",
		"formData": map[string]any{
			"name":        data.Payee.Name,
			"description": data.Payee.Description,
		},
		"errors": map[string]any{
			"name":        "",
			"description": "",
		},
	}
}
