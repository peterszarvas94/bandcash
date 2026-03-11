package utils

import (
	"fmt"
	"strings"
)

type TableColumn struct {
	Key     string
	WidthCh int
}

type TableLayout struct {
	DataWidthCh     int
	ActionsWidthRem int
	Columns         map[string]int
	ColumnOrder     []string
}

func NewTableLayout(columns []TableColumn, actionsWidthRem int) TableLayout {
	dataWidthCh := 0
	columnMap := make(map[string]int, len(columns))
	columnOrder := make([]string, 0, len(columns))
	for _, column := range columns {
		dataWidthCh += column.WidthCh
		columnMap[column.Key] = column.WidthCh
		columnOrder = append(columnOrder, column.Key)
	}

	return TableLayout{
		DataWidthCh:     dataWidthCh,
		ActionsWidthRem: actionsWidthRem,
		Columns:         columnMap,
		ColumnOrder:     columnOrder,
	}
}

func (t TableLayout) Col(key string) int {
	return t.Columns[key]
}

func (t TableLayout) GridTemplate() string {
	parts := make([]string, 0, len(t.ColumnOrder)+1)
	for _, key := range t.ColumnOrder {
		widthCh := t.Columns[key]
		parts = append(parts, fmt.Sprintf("minmax(%dch, %dfr)", widthCh, widthCh))
	}
	if t.ActionsWidthRem > 0 {
		parts = append(parts, fmt.Sprintf("%drem", t.ActionsWidthRem))
	}

	return strings.Join(parts, " ")
}
