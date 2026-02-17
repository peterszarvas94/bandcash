package payeepages

import (
	"fmt"

	"bandcash/internal/utils"
	payeeview "bandcash/models/payee/templates/view"
)

func payeeIndexSignals() string {
	return "{mode: '', formState: '', editingId: 0, formData: {name: '', description: ''}}"
}

func payeeShowSignals(data payeeview.PayeeData) string {
	return fmt.Sprintf(
		"{mode: 'single', formState: '', formData: {name: %s, description: %s}}",
		utils.JSString(data.Payee.Name),
		utils.JSString(data.Payee.Description),
	)
}
