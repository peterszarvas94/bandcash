package expense

import (
	"context"

	ctxi18n "github.com/invopop/ctxi18n/i18n"

	"bandcash/internal/db"
	"bandcash/internal/utils"
)

type Expenses struct{}

func (e *Expenses) TableQuerySpec() utils.TableQuerySpec {
	return utils.TableQuerySpec{
		DefaultSort: "date",
		DefaultDir:  "desc",
		AllowedSorts: map[string]struct{}{
			"date":   {},
			"title":  {},
			"amount": {},
		},
		AllowedPageSizes: map[int]struct{}{
			10:  {},
			50:  {},
			100: {},
			200: {},
		},
		DefaultSize:  50,
		MaxSearchLen: 100,
	}
}

func New() *Expenses {
	return &Expenses{}
}

func (e *Expenses) GetIndexData(ctx context.Context, groupID string, query utils.TableQuery) (ExpensesData, error) {
	group, err := db.Qry.GetGroupByID(ctx, groupID)
	if err != nil {
		return ExpensesData{}, err
	}

	totalItems, err := db.Qry.CountExpensesFiltered(ctx, db.CountExpensesFilteredParams{
		GroupID: groupID,
		Search:  query.Search,
	})
	if err != nil {
		return ExpensesData{}, err
	}

	query = utils.ClampPage(query, totalItems)

	params := db.ListExpensesByDateDescFilteredParams{
		GroupID: groupID,
		Search:  query.Search,
		Limit:   int64(query.PageSize),
		Offset:  query.Offset(),
	}

	var expenses []db.Expense
	switch query.Sort {
	case "title":
		if query.Dir == "desc" {
			expenses, err = db.Qry.ListExpensesByTitleDescFiltered(ctx, db.ListExpensesByTitleDescFilteredParams(params))
		} else {
			expenses, err = db.Qry.ListExpensesByTitleAscFiltered(ctx, db.ListExpensesByTitleAscFilteredParams(params))
		}
	case "amount":
		if query.Dir == "desc" {
			expenses, err = db.Qry.ListExpensesByAmountDescFiltered(ctx, db.ListExpensesByAmountDescFilteredParams(params))
		} else {
			expenses, err = db.Qry.ListExpensesByAmountAscFiltered(ctx, db.ListExpensesByAmountAscFilteredParams(params))
		}
	default:
		if query.Dir == "asc" {
			expenses, err = db.Qry.ListExpensesByDateAscFiltered(ctx, db.ListExpensesByDateAscFilteredParams(params))
		} else {
			expenses, err = db.Qry.ListExpensesByDateDescFiltered(ctx, params)
		}
	}
	if err != nil {
		return ExpensesData{}, err
	}

	return ExpensesData{
		Title:    ctxi18n.T(ctx, "expenses.title"),
		Expenses: expenses,
		Query:    query,
		Pager:    utils.BuildTablePagination(totalItems, query),
		GroupID:  groupID,
		Breadcrumbs: []utils.Crumb{
			{Label: ctxi18n.T(ctx, "groups.title"), Href: "/dashboard"},
			{Label: group.Name, Href: "/groups/" + groupID},
			{Label: ctxi18n.T(ctx, "expenses.title")},
		},
	}, nil
}
