package payee

import (
	"fmt"

	"bandcash/internal/utils"
)

func payeeShowResetAction(data PayeeData) string {
	return fmt.Sprintf(
		"$formState = ''; $formData = {name: %s, description: %s}; $errors = {name: '', description: ''}",
		utils.JSString(data.Payee.Name),
		utils.JSString(data.Payee.Description),
	)
}
