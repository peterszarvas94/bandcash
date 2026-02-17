package payeecomponents

import (
	"fmt"

	"bandcash/internal/utils"
	payeeview "bandcash/models/payee/templates/view"
)

func payeeShowResetAction(data payeeview.PayeeData) string {
	return fmt.Sprintf(
		"$formState = ''; $formData = {name: %s, description: %s}; $errors = {name: '', description: ''}",
		utils.JSString(data.Payee.Name),
		utils.JSString(data.Payee.Description),
	)
}
