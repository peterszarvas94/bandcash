package payee

import (
	"context"
	"strconv"

	"bandcash/internal/db"
	"bandcash/internal/utils"
	payeeview "bandcash/models/payee/templates/view"
)

type Payees struct {
}

func New() *Payees {
	return &Payees{}
}

func (p *Payees) GetShowData(ctx context.Context, id int) (payeeview.PayeeData, error) {
	payee, err := db.Qry.GetPayee(ctx, int64(id))
	if err != nil {
		return payeeview.PayeeData{}, err
	}

	entries, err := db.Qry.ListParticipantsByPayee(ctx, int64(id))
	if err != nil {
		return payeeview.PayeeData{}, err
	}

	return payeeview.PayeeData{
		Title:   payee.Name,
		Payee:   &payee,
		Entries: entries,
		Breadcrumbs: []utils.Crumb{
			{Label: "Payees", Href: "/payee"},
			{Label: payee.Name, Href: "/payee/" + strconv.Itoa(id)},
		},
	}, nil
}

func (p *Payees) GetIndexData(ctx context.Context) (payeeview.PayeesData, error) {
	payees, err := db.Qry.ListPayees(ctx)
	if err != nil {
		return payeeview.PayeesData{}, err
	}
	return payeeview.PayeesData{
		Title:  "Payees",
		Payees: payees,
		Breadcrumbs: []utils.Crumb{
			{Label: "Payees"},
		},
	}, nil
}
