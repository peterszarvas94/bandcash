package utils

type TableColumn struct {
	Key     string
	MaxWRem int
}

type TableLayout struct {
	ActionsWidthRem int
	ColumnMaxWRem   map[string]int
	ColumnOrder     []string
}

func NewTableLayout(columns []TableColumn, actionsWidthRem int) TableLayout {
	columnMaxWRemMap := make(map[string]int, len(columns))
	columnOrder := make([]string, 0, len(columns))
	for _, column := range columns {
		columnMaxWRemMap[column.Key] = column.MaxWRem
		columnOrder = append(columnOrder, column.Key)
	}

	return TableLayout{
		ActionsWidthRem: actionsWidthRem,
		ColumnMaxWRem:   columnMaxWRemMap,
		ColumnOrder:     columnOrder,
	}
}

func (t TableLayout) ColMaxWRem(key string) int {
	maxWRem, ok := t.ColumnMaxWRem[key]
	if !ok || maxWRem <= 0 {
		return 24
	}

	return maxWRem
}
