package member

func memberIndexSignals() map[string]any {
	return map[string]any{
		"mode":      "",
		"formState": "",
		"editingId": 0,
		"formData":  map[string]any{"name": "", "description": ""},
	}
}

func memberShowSignals(data MemberData) map[string]any {
	return map[string]any{
		"mode":      "single",
		"formState": "",
		"formData": map[string]any{
			"name":        data.Member.Name,
			"description": data.Member.Description,
		},
		"errors": map[string]any{
			"name":        "",
			"description": "",
		},
	}
}
