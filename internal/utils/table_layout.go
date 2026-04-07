package utils

type TableColumn struct {
	Key     string
	MaxWRem int
	WRem    int
}

type TableLayout struct {
	ActionsWidthRem int
	ColumnMaxWRem   map[string]int
	ColumnWRem      map[string]int
	ColumnOrder     []string
}

func NewTableLayout(columns []TableColumn, actionsWidthRem int) TableLayout {
	columnMaxWRemMap := make(map[string]int, len(columns))
	columnWRemMap := make(map[string]int, len(columns))
	columnOrder := make([]string, 0, len(columns))
	for _, column := range columns {
		columnMaxWRemMap[column.Key] = column.MaxWRem
		columnWRemMap[column.Key] = column.WRem
		columnOrder = append(columnOrder, column.Key)
	}

	return TableLayout{
		ActionsWidthRem: actionsWidthRem,
		ColumnMaxWRem:   columnMaxWRemMap,
		ColumnWRem:      columnWRemMap,
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

func (t TableLayout) ColWRem(key string) int {
	wRem, ok := t.ColumnWRem[key]
	if !ok || wRem <= 0 {
		return 0
	}
	return wRem
}
