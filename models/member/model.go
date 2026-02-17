package member

import (
	"context"
	"strconv"

	ctxi18n "github.com/invopop/ctxi18n/i18n"

	"bandcash/internal/db"
	"bandcash/internal/utils"
)

type Members struct {
}

func New() *Members {
	return &Members{}
}

func (p *Members) GetShowData(ctx context.Context, id int) (MemberData, error) {
	member, err := db.Qry.GetMember(ctx, int64(id))
	if err != nil {
		return MemberData{}, err
	}

	events, err := db.Qry.ListParticipantsByMember(ctx, int64(id))
	if err != nil {
		return MemberData{}, err
	}

	return MemberData{
		Title:  member.Name,
		Member: &member,
		Events: events,
		Breadcrumbs: []utils.Crumb{
			{Label: ctxi18n.T(ctx, "members.title"), Href: "/member"},
			{Label: member.Name, Href: "/member/" + strconv.Itoa(id)},
		},
	}, nil
}

func (p *Members) GetIndexData(ctx context.Context) (MembersData, error) {
	members, err := db.Qry.ListMembers(ctx)
	if err != nil {
		return MembersData{}, err
	}
	return MembersData{
		Title:   ctxi18n.T(ctx, "members.title"),
		Members: members,
		Breadcrumbs: []utils.Crumb{
			{Label: ctxi18n.T(ctx, "members.title")},
		},
	}, nil
}
