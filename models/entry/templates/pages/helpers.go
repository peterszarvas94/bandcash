package entrypages

import (
	"fmt"

	"bandcash/internal/utils"
	entryview "bandcash/models/entry/templates/view"
)

func entryIndexSignals() string {
	return "{mode: '', formState: '', editingId: 0, formData: {title: '', time: '', description: '', amount: 0}}"
}

func entryShowSignals(data entryview.EntryData) string {
	return fmt.Sprintf(
		"{mode: 'single', entryFormState: '', entryFormData: {title: %s, time: %s, description: %s, amount: %d}, formState: '', editingId: 0, calcPercent: 0, formData: {payeeId: '', payeeName: '', amount: 0, expense: 0}}",
		utils.JSString(data.Entry.Title),
		utils.JSString(data.Entry.Time),
		utils.JSString(data.Entry.Description),
		data.Entry.Amount,
	)
}
