package utils

type TableLayout struct {
	DataWidthCh     int
	ActionsWidthRem int
	Columns         map[string]int
}

func NewTableLayout(columns map[string]int, actionsWidthRem int) TableLayout {
	dataWidthCh := 0
	for _, widthCh := range columns {
		dataWidthCh += widthCh
	}

	return TableLayout{
		DataWidthCh:     dataWidthCh,
		ActionsWidthRem: actionsWidthRem,
		Columns:         columns,
	}
}

func (t TableLayout) Col(key string) int {
	return t.Columns[key]
}
