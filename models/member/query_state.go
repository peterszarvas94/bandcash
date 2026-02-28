package member

import "bandcash/internal/utils"

func memberQuerySignals(query utils.TableQuery) map[string]any {
	sort := ""
	dir := ""
	if query.SortSet {
		sort = query.Sort
		dir = query.Dir
	}

	return map[string]any{
		"search":   query.Search,
		"sort":     sort,
		"sortSet":  query.SortSet,
		"dir":      dir,
		"page":     query.Page,
		"pageSize": query.PageSize,
	}
}
