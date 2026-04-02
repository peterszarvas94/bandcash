package admin

import "bandcash/internal/utils"

func AdminUsersTableLayout() utils.TableLayout {
	return utils.NewTableLayout([]utils.TableColumn{
		{Key: "email"},
		{Key: "created_at"},
		{Key: "status"},
	}, 2)
}

func AdminGroupsTableLayout() utils.TableLayout {
	return utils.NewTableLayout([]utils.TableColumn{
		{Key: "name"},
		{Key: "admin"},
		{Key: "viewers"},
		{Key: "created_at"},
	}, 0)
}

func AdminSessionsTableLayout() utils.TableLayout {
	return utils.NewTableLayout([]utils.TableColumn{
		{Key: "user_email"},
		{Key: "session_id"},
		{Key: "created_at"},
		{Key: "expires_at"},
		{Key: "actions"},
	}, 0)
}
