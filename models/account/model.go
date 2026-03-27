package account

import (
	"context"

	ctxi18n "github.com/invopop/ctxi18n/i18n"

	appi18n "bandcash/internal/i18n"
	"bandcash/internal/utils"
)

type Account struct{}

func New() *Account {
	return &Account{}
}

func (s *Account) Data(ctx context.Context) AccountData {
	return AccountData{
		Title:       ctxi18n.T(ctx, "account.page_title"),
		CurrentLang: appi18n.LocaleCode(ctx),
		Breadcrumbs: []utils.Crumb{
			{Label: ctxi18n.T(ctx, "account.title")},
		},
	}
}
