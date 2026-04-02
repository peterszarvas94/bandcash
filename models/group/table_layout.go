package group

import "bandcash/internal/utils"

func GroupUsersTableLayout() utils.TableLayout {
	return utils.NewTableLayout([]utils.TableColumn{
		{Key: "email"},
		{Key: "role"},
		{Key: "status"},
	}, 0)
}
