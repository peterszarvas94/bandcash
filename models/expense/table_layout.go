package expense

import "bandcash/internal/utils"

func ExpensesIndexTableLayout() utils.TableLayout {
	return utils.NewTableLayout([]utils.TableColumn{
		{Key: "title"},
		{Key: "date"},
		{Key: "amount"},
		{Key: "paid"},
		{Key: "paid_at"},
	}, 0)
}
