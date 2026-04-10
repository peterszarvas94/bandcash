package event

import "bandcash/internal/utils"

type eventInlineParams struct {
	TabID                 string           `json:"tab_id"`
	FormData              eventData        `json:"formData"`
	EventFormData         eventData        `json:"eventFormData"`
	TableQuery            utils.TableQuery `json:"tableQuery"`
	Mode                  string           `json:"mode"`
	ParticipantNoteDialog struct {
		MemberID string `json:"memberId"`
		Value    string `json:"value"`
	} `json:"participantNoteDialog"`
	ParticipantPaidAtDialog struct {
		MemberID string `json:"memberId"`
		Value    string `json:"value"`
	} `json:"participantPaidAtDialog"`
}

type modeParams struct {
	TabID      string           `json:"tab_id"`
	Mode       string           `json:"mode"`
	TableQuery utils.TableQuery `json:"tableQuery"`
}

type paidAtParams struct {
	TabID        string           `json:"tab_id"`
	Mode         string           `json:"mode"`
	TableQuery   utils.TableQuery `json:"tableQuery"`
	PaidAtDialog struct {
		Value string `json:"value"`
	} `json:"paidAtDialog"`
}

type participantNoteDialogParams struct {
	TabID                 string           `json:"tab_id"`
	TableQuery            utils.TableQuery `json:"tableQuery"`
	ParticipantNoteDialog struct {
		MemberID string `json:"memberId"`
	} `json:"participantNoteDialog"`
}

type participantPaidAtDialogParams struct {
	TabID                   string           `json:"tab_id"`
	TableQuery              utils.TableQuery `json:"tableQuery"`
	ParticipantPaidAtDialog struct {
		MemberID string `json:"memberId"`
		Value    string `json:"value"`
	} `json:"participantPaidAtDialog"`
}

type eventData struct {
	Title       string `json:"title" validate:"required,min=1,max=255"`
	Date        string `json:"date" validate:"required"`
	Time        string `json:"time" validate:"required"`
	Place       string `json:"place" validate:"max=255"`
	Description string `json:"description" validate:"max=1000"`
	Amount      int64  `json:"amount" validate:"required,gt=0"`
	Paid        bool   `json:"paid"`
	PaidAt      string `json:"paidAt"`
}

type participantBulkRowData struct {
	RowID      string `json:"rowId"`
	MemberID   string `json:"memberId"`
	MemberName string `json:"memberName"`
	Included   bool   `json:"included"`
	Amount     int64  `json:"amount" validate:"gte=0"`
	Expense    int64  `json:"expense" validate:"gte=0"`
	Note       string `json:"note" validate:"max=1000"`
	Paid       bool   `json:"paid"`
	PaidAt     string `json:"paidAt"`
}

type participantWizardSignals struct {
	EventAmount int64                    `json:"eventAmount"`
	Rows        []participantBulkRowData `json:"rows"`
	MemberIDs   map[string]string        `json:"memberIds"`
	Amounts     map[string]int64         `json:"amounts"`
	Expenses    map[string]int64         `json:"expenses"`
	Notes       map[string]string        `json:"notes"`
	Paids       map[string]bool          `json:"paids"`
	PaidAts     map[string]string        `json:"paidAts"`
	Total       int64                    `json:"total"`
	Balance     int64                    `json:"balance"`
	Error       string                   `json:"error"`
}

type participantBulkParams struct {
	TabID         string                   `json:"tab_id"`
	EventFormData eventData                `json:"eventFormData"`
	Wizard        participantWizardSignals `json:"wizard"`
	TableQuery    utils.TableQuery         `json:"tableQuery"`
}

type participantDraftParams struct {
	TabID         string                   `json:"tab_id"`
	EventFormData eventData                `json:"eventFormData"`
	Wizard        participantWizardSignals `json:"wizard"`
	TableQuery    utils.TableQuery         `json:"tableQuery"`
}

type participantDraftRowParams struct {
	TabID           string                   `json:"tab_id"`
	DraftRowsAction string                   `json:"draftRowsAction"`
	DraftRowsRowID  string                   `json:"draftRowsRowId"`
	EventFormData   eventData                `json:"eventFormData"`
	Wizard          participantWizardSignals `json:"wizard"`
	TableQuery      utils.TableQuery         `json:"tableQuery"`
}

func eventIndexSignals(query utils.TableQuery) map[string]any {
	return map[string]any{
		"tableQuery":      utils.TableQuerySignals(query),
		"dateRange":       map[string]any{"from": query.From, "to": query.To},
		"showCustomRange": query.DateMode == "custom" || query.From != "" || query.To != "",
		"mode":            "table",
		"formState":       "",
		"editingId":       0,
		"formData":        map[string]any{"title": "", "date": "", "time": "", "place": "", "description": "", "amount": 0, "paid": false, "paidAt": ""},
		"eventFormState":  "",
		"summaryMode":     query.Summary,
		"_fetching":       false,
		"paidAtDialog": map[string]any{
			"open":        false,
			"fetching":    false,
			"mode":        "table",
			"title":       "",
			"message":     "",
			"value":       "",
			"placeholder": "",
			"submitLabel": "",
			"cancelLabel": "",
			"url":         "",
			"triggerID":   "",
		},
		"errors": map[string]any{"title": "", "date": "", "time": "", "place": "", "description": "", "amount": "", "memberId": "", "expense": ""},
	}
}

func eventShowSignals(data EventData) map[string]any {
	wizardRows := make([]map[string]any, 0, len(data.WizardRows))
	wizardMemberIDs := make(map[string]string, len(data.WizardRows))
	wizardAmounts := make(map[string]int64, len(data.WizardRows))
	wizardExpenses := make(map[string]int64, len(data.WizardRows))
	wizardNotes := make(map[string]string, len(data.WizardRows))
	wizardPaids := make(map[string]bool, len(data.WizardRows))
	wizardPaidAts := make(map[string]string, len(data.WizardRows))
	wizardTotal := int64(0)
	for _, row := range data.WizardRows {
		rowID := row.RowID
		if rowID == "" {
			rowID = row.MemberID
		}
		wizardRows = append(wizardRows, map[string]any{
			"rowId":      rowID,
			"memberId":   row.MemberID,
			"memberName": row.MemberName,
			"included":   row.Included,
			"amount":     row.Amount,
			"expense":    row.Expense,
			"note":       row.Note,
			"paid":       row.Paid,
			"paidAt":     row.PaidAt,
		})
		wizardMemberIDs[rowID] = row.MemberID
		wizardAmounts[rowID] = row.Amount
		wizardExpenses[rowID] = row.Expense
		wizardNotes[rowID] = row.Note
		wizardPaids[rowID] = row.Paid
		wizardPaidAts[rowID] = row.PaidAt
		wizardTotal += row.Amount + row.Expense
	}

	return map[string]any{
		"mode":                  "single",
		"draftRowsAction":       "",
		"draftRowsRowId":        "",
		"tableQuery":            utils.TableQuerySignals(data.Query),
		"eventFormState":        "",
		"participantEditorMode": data.EditorMode,
		"wizard": map[string]any{
			"error":       data.WizardError,
			"eventAmount": data.WizardEventAmount,
			"rows":        wizardRows,
			"memberIds":   wizardMemberIDs,
			"amounts":     wizardAmounts,
			"expenses":    wizardExpenses,
			"notes":       wizardNotes,
			"paids":       wizardPaids,
			"paidAts":     wizardPaidAts,
			"total":       wizardTotal,
			"balance":     data.WizardEventAmount - wizardTotal,
		},
		"eventFormData": map[string]any{
			"title":       data.Event.Title,
			"date":        eventDateValue(*data.Event),
			"time":        eventTimeValue(*data.Event),
			"place":       data.Event.Place,
			"description": data.Event.Description,
			"amount":      data.Event.Amount,
			"paid":        data.Event.Paid == 1,
			"paidAt": func() string {
				if !data.Event.PaidAt.Valid {
					return ""
				}
				return utils.FormatDateInput(data.Event.PaidAt.String)
			}(),
		},
		"formState":   "",
		"editingId":   0,
		"calcPercent": 0,
		"summaryMode": data.Query.Summary,
		"_fetching":   false,
		"paidAtDialog": map[string]any{
			"open":        data.PaidAtDialog.Open,
			"fetching":    data.PaidAtDialog.Fetching,
			"mode":        data.PaidAtDialog.Mode,
			"title":       data.PaidAtDialog.Title,
			"message":     data.PaidAtDialog.Message,
			"value":       data.PaidAtDialog.Value,
			"placeholder": data.PaidAtDialog.Placeholder,
			"submitLabel": data.PaidAtDialog.SubmitLabel,
			"cancelLabel": data.PaidAtDialog.CancelLabel,
			"url":         data.PaidAtDialog.URL,
			"triggerID":   data.PaidAtDialog.TriggerID,
		},
		"participantPaidAtDialog": map[string]any{
			"open":        data.ParticipantPaidAtDialog.Open,
			"fetching":    data.ParticipantPaidAtDialog.Fetching,
			"title":       data.ParticipantPaidAtDialog.Title,
			"message":     data.ParticipantPaidAtDialog.Message,
			"memberId":    data.ParticipantPaidAtDialog.MemberID,
			"value":       data.ParticipantPaidAtDialog.Value,
			"submitLabel": data.ParticipantPaidAtDialog.SubmitLabel,
			"cancelLabel": data.ParticipantPaidAtDialog.CancelLabel,
			"url":         data.ParticipantPaidAtDialog.URL,
			"triggerID":   data.ParticipantPaidAtDialog.TriggerID,
		},
		"participantNoteDialog": map[string]any{
			"open":        data.ParticipantNoteDialog.Open,
			"fetching":    data.ParticipantNoteDialog.Fetching,
			"readOnly":    data.ParticipantNoteDialog.ReadOnly,
			"title":       data.ParticipantNoteDialog.Title,
			"message":     data.ParticipantNoteDialog.Message,
			"memberId":    data.ParticipantNoteDialog.MemberID,
			"value":       data.ParticipantNoteDialog.Value,
			"submitLabel": data.ParticipantNoteDialog.SubmitLabel,
			"cancelLabel": data.ParticipantNoteDialog.CancelLabel,
			"url":         data.ParticipantNoteDialog.URL,
			"triggerID":   data.ParticipantNoteDialog.TriggerID,
		},
		"formData": map[string]any{
			"memberId":   "",
			"memberName": "",
			"amount":     0,
			"expense":    0,
		},
		"errors": map[string]any{
			"title":       "",
			"date":        "",
			"time":        "",
			"place":       "",
			"description": "",
			"amount":      "",
			"memberId":    "",
			"expense":     "",
		},
	}
}
