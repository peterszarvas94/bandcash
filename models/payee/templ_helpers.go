package payee

import (
	"fmt"

	"bandcash/internal/utils"
)

func payeeIndexSignals() string {
	return "{mode: '', formState: '', editingId: 0, formData: {name: '', description: ''}}"
}

func payeeShowSignals(data PayeeData) string {
	return fmt.Sprintf(
		"{mode: 'single', formState: '', formData: {name: %s, description: %s}}",
		utils.JSString(data.Payee.Name),
		utils.JSString(data.Payee.Description),
	)
}

func payeeShowResetAction(data PayeeData) string {
	return fmt.Sprintf(
		"$formState = ''; $formData = {name: %s, description: %s}; $errors = {name: '', description: ''}",
		utils.JSString(data.Payee.Name),
		utils.JSString(data.Payee.Description),
	)
}
