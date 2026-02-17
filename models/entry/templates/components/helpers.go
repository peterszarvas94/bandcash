package entrycomponents

import (
	"fmt"

	"bandcash/internal/utils"
	entryview "bandcash/models/entry/templates/view"
)

func entryShowResetAction(data entryview.EntryData) string {
	return fmt.Sprintf(
		"$entryFormState = ''; $entryFormData = {title: %s, time: %s, description: %s, amount: %d}; $errors = {title: '', time: '', description: '', amount: ''}",
		utils.JSString(data.Entry.Title),
		utils.JSString(data.Entry.Time),
		utils.JSString(data.Entry.Description),
		data.Entry.Amount,
	)
}
