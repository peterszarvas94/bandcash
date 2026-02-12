package payee

import (
	"context"
	"html/template"

	"bandcash/internal/db"
	"bandcash/internal/view"
)

type PayeeData struct {
	Title       string
	Payee       *db.Payee
	Entries     []db.ListParticipantsByPayeeRow
	Breadcrumbs []view.Crumb
}

type PayeesData struct {
	Title       string
	Payees      []db.Payee
	Breadcrumbs []view.Crumb
}

type Payees struct {
	tmpl *template.Template
}

func New() *Payees {
	return &Payees{}
}

func (p *Payees) SetTemplate(tmpl *template.Template) {
	p.tmpl = tmpl
}

func (p *Payees) CreatePayee(ctx context.Context, name, description string) (*db.Payee, error) {
	payee, err := db.Qry.CreatePayee(ctx, db.CreatePayeeParams{
		Name:        name,
		Description: description,
	})
	if err != nil {
		return nil, err
	}
	return &payee, nil
}

func (p *Payees) GetPayee(ctx context.Context, id int) (*db.Payee, error) {
	payee, err := db.Qry.GetPayee(ctx, int64(id))
	if err != nil {
		return nil, err
	}
	return &payee, nil
}

func (p *Payees) GetEntries(ctx context.Context, payeeID int) ([]db.ListParticipantsByPayeeRow, error) {
	return db.Qry.ListParticipantsByPayee(ctx, int64(payeeID))
}

func (p *Payees) UpdatePayee(ctx context.Context, id int, name, description string) (*db.Payee, error) {
	updated, err := db.Qry.UpdatePayee(ctx, db.UpdatePayeeParams{
		Name:        name,
		Description: description,
		ID:          int64(id),
	})
	if err != nil {
		return nil, err
	}
	return &updated, nil
}

func (p *Payees) AllPayees(ctx context.Context) ([]db.Payee, error) {
	return db.Qry.ListPayees(ctx)
}

func (p *Payees) DeletePayee(ctx context.Context, id int) error {
	return db.Qry.DeletePayee(ctx, int64(id))
}

func (p *Payees) GetIndexData(ctx context.Context) (any, error) {
	payees, err := p.AllPayees(ctx)
	if err != nil {
		return nil, err
	}
	return PayeesData{
		Title:  "Payees",
		Payees: payees,
		Breadcrumbs: []view.Crumb{
			{Label: "Payees"},
		},
	}, nil
}
