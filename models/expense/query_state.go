package expense

import "bandcash/internal/utils"

func expenseQuerySignals(query utils.TableQuery) map[string]any {
	return utils.TableQuerySignals(query)
}
