package member

import "bandcash/internal/utils"

func memberQuerySignals(query utils.TableQuery) map[string]any {
	return utils.TableQuerySignals(query)
}
