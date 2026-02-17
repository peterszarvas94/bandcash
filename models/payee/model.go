package payee

import (
	"context"
	"strconv"

	"bandcash/internal/db"
	"bandcash/internal/utils"
)

type PayeeData struct {
	Title       string
	Payee       *db.Payee
	Entries     []db.ListParticipantsByPayeeRow
	Breadcrumbs []utils.Crumb
}

type PayeesData struct {
	Title       string
	Payees      []db.Payee
	Breadcrumbs []utils.Crumb
}

type Payees struct {
}

func New() *Payees {
	return &Payees{}
}

func (p *Payees) GetShowData(ctx context.Context, id int) (PayeeData, error) {
	payee, err := db.Qry.GetPayee(ctx, int64(id))
	if err != nil {
		return PayeeData{}, err
	}

	entries, err := db.Qry.ListParticipantsByPayee(ctx, int64(id))
	if err != nil {
		return PayeeData{}, err
	}

	return PayeeData{
		Title:   payee.Name,
		Payee:   &payee,
		Entries: entries,
		Breadcrumbs: []utils.Crumb{
			{Label: "Payees", Href: "/payee"},
			{Label: payee.Name, Href: "/payee/" + strconv.Itoa(id)},
		},
	}, nil
}

func (p *Payees) GetIndexData(ctx context.Context) (PayeesData, error) {
	payees, err := db.Qry.ListPayees(ctx)
	if err != nil {
		return PayeesData{}, err
	}
	return PayeesData{
		Title:  "Payees",
		Payees: payees,
		Breadcrumbs: []utils.Crumb{
			{Label: "Payees"},
		},
	}, nil
}
