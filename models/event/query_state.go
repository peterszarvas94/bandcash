package event

import "bandcash/internal/utils"

func eventQuerySignals(query utils.TableQuery) map[string]any {
	return utils.TableQuerySignals(query)
}
