package utils

type TableLayout struct {
	MinWidthCh     int
	ActionsWidthCh int
	Columns        map[string]int
}

func NewTableLayout(columns map[string]int, actionsWidthCh int) TableLayout {
	minWidthCh := actionsWidthCh
	for _, widthCh := range columns {
		minWidthCh += widthCh
	}

	return TableLayout{
		MinWidthCh:     minWidthCh,
		ActionsWidthCh: actionsWidthCh,
		Columns:        columns,
	}
}

func (t TableLayout) Col(key string) int {
	return t.Columns[key]
}
