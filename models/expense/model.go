package expense

import (
	"context"

	ctxi18n "github.com/invopop/ctxi18n/i18n"

	"bandcash/internal/db"
	"bandcash/internal/utils"
)

type Expenses struct{}

func New() *Expenses {
	return &Expenses{}
}

func (e *Expenses) GetIndexData(ctx context.Context, groupID string) (ExpensesData, error) {
	group, err := db.Qry.GetGroupByID(ctx, groupID)
	if err != nil {
		return ExpensesData{}, err
	}

	expenses, err := db.Qry.ListExpenses(ctx, groupID)
	if err != nil {
		return ExpensesData{}, err
	}

	return ExpensesData{
		Title:    ctxi18n.T(ctx, "expenses.title"),
		Expenses: expenses,
		GroupID:  groupID,
		Breadcrumbs: []utils.Crumb{
			{Label: ctxi18n.T(ctx, "groups.title"), Href: "/dashboard"},
			{Label: group.Name, Href: "/groups/" + groupID},
			{Label: ctxi18n.T(ctx, "expenses.title")},
		},
	}, nil
}
