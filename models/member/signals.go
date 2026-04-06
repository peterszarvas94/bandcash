package member

import "bandcash/internal/utils"

func memberIndexSignals(query map[string]any) map[string]any {
	return map[string]any{
		"tableQuery":     query,
		"mode":           "table",
		"formState":      "",
		"eventFormState": "",
		"editingId":      0,
		"formData":       map[string]any{"name": "", "description": ""},
		"errors":         map[string]any{"name": "", "description": ""},
	}
}

func memberShowSignals(data MemberData) map[string]any {
	return map[string]any{
		"mode":           "single",
		"tableQuery":     utils.TableQuerySignals(data.Query),
		"formState":      "",
		"eventFormState": "",
		"summaryMode":    data.Query.Summary,
		"formData": map[string]any{
			"name":        data.Member.Name,
			"description": data.Member.Description,
		},
		"errors": map[string]any{
			"name":        "",
			"description": "",
		},
		"participantPaidAtDialog": map[string]any{
			"open":        data.PaidAtDialog.Open,
			"fetching":    data.PaidAtDialog.Fetching,
			"title":       data.PaidAtDialog.Title,
			"message":     data.PaidAtDialog.Message,
			"eventId":     data.PaidAtDialog.EventID,
			"value":       data.PaidAtDialog.Value,
			"submitLabel": data.PaidAtDialog.SubmitLabel,
			"cancelLabel": data.PaidAtDialog.CancelLabel,
			"url":         data.PaidAtDialog.URL,
			"triggerID":   data.PaidAtDialog.TriggerID,
		},
	}
}
