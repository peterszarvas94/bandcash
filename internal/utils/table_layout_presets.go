package utils

func EventsIndexTableLayout() TableLayout {
	return NewTableLayout(map[string]int{
		"title":       28,
		"time":        19,
		"amount":      12,
		"description": 30,
	}, 8)
}

func EventParticipantsTableLayout() TableLayout {
	return NewTableLayout(map[string]int{
		"name":    24,
		"total":   12,
		"amount":  12,
		"expense": 12,
	}, 8)
}

func ExpensesIndexTableLayout() TableLayout {
	return NewTableLayout(map[string]int{
		"title":       20,
		"description": 30,
		"amount":      12,
		"date":        12,
	}, 8)
}

func GroupsIndexTableLayout() TableLayout {
	return NewTableLayout(map[string]int{
		"name":    24,
		"role":    10,
		"created": 12,
		"admin":   28,
	}, 8)
}

func MembersIndexTableLayout() TableLayout {
	return NewTableLayout(map[string]int{
		"name":        20,
		"description": 32,
	}, 8)
}

func MemberEventsTableLayout() TableLayout {
	return NewTableLayout(map[string]int{
		"title":               22,
		"time":                19,
		"description":         32,
		"amount":              12,
		"participant_amount":  12,
		"participant_expense": 12,
		"total":               12,
	}, 0)
}

func ViewersAdminsTableLayout() TableLayout {
	return NewTableLayout(map[string]int{
		"email": 34,
	}, 0)
}

func ViewersPendingTableLayout() TableLayout {
	return NewTableLayout(map[string]int{
		"email":      34,
		"created_at": 18,
	}, 4)
}

func ViewersTableLayout() TableLayout {
	return NewTableLayout(map[string]int{
		"email": 34,
	}, 4)
}

func AdminUsersTableLayout() TableLayout {
	return NewTableLayout(map[string]int{
		"email":      34,
		"created_at": 18,
		"status":     12,
	}, 18)
}

func AdminGroupsTableLayout() TableLayout {
	return NewTableLayout(map[string]int{
		"name":       24,
		"admin":      24,
		"viewers":    10,
		"created_at": 18,
	}, 0)
}
