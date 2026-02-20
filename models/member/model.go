package member

import (
	"context"

	ctxi18n "github.com/invopop/ctxi18n/i18n"

	"bandcash/internal/db"
	"bandcash/internal/utils"
)

type Members struct {
}

func New() *Members {
	return &Members{}
}

func (p *Members) GetShowData(ctx context.Context, groupID, memberID string) (MemberData, error) {
	group, err := db.Qry.GetGroupByID(ctx, groupID)
	if err != nil {
		return MemberData{}, err
	}

	member, err := db.Qry.GetMember(ctx, db.GetMemberParams{
		ID:      memberID,
		GroupID: groupID,
	})
	if err != nil {
		return MemberData{}, err
	}

	events, err := db.Qry.ListParticipantsByMember(ctx, db.ListParticipantsByMemberParams{
		MemberID: memberID,
		GroupID:  groupID,
	})
	if err != nil {
		return MemberData{}, err
	}

	return MemberData{
		Title:   member.Name,
		Member:  &member,
		Events:  events,
		GroupID: groupID,
		Breadcrumbs: []utils.Crumb{
			{Label: ctxi18n.T(ctx, "groups.title"), Href: "/dashboard"},
			{Label: group.Name, Href: "/groups/" + groupID + "/events"},
			{Label: ctxi18n.T(ctx, "members.title"), Href: "/groups/" + groupID + "/members"},
			{Label: member.Name},
		},
	}, nil
}

func (p *Members) GetIndexData(ctx context.Context, groupID string) (MembersData, error) {
	group, err := db.Qry.GetGroupByID(ctx, groupID)
	if err != nil {
		return MembersData{}, err
	}

	members, err := db.Qry.ListMembers(ctx, groupID)
	if err != nil {
		return MembersData{}, err
	}
	return MembersData{
		Title:   ctxi18n.T(ctx, "members.title"),
		Members: members,
		GroupID: groupID,
		Breadcrumbs: []utils.Crumb{
			{Label: ctxi18n.T(ctx, "groups.title"), Href: "/dashboard"},
			{Label: group.Name, Href: "/groups/" + groupID + "/events"},
			{Label: ctxi18n.T(ctx, "members.title")},
		},
	}, nil
}
