package member

import (
	"context"

	ctxi18n "github.com/invopop/ctxi18n/i18n"

	"bandcash/internal/db"
	"bandcash/internal/utils"
)

type Members struct {
}

func (p *Members) TableQuerySpec() utils.TableQuerySpec {
	return utils.StandardTableQuerySpec("createdAt", "desc", "name", "createdAt")
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
			{Label: group.Name, Href: "/groups/" + groupID},
			{Label: ctxi18n.T(ctx, "members.title"), Href: "/groups/" + groupID + "/members"},
			{Label: member.Name},
		},
	}, nil
}

func (p *Members) GetIndexData(ctx context.Context, groupID string, query utils.TableQuery) (MembersData, error) {
	group, err := db.Qry.GetGroupByID(ctx, groupID)
	if err != nil {
		return MembersData{}, err
	}

	totalItems, err := db.Qry.CountMembersFiltered(ctx, db.CountMembersFilteredParams{
		GroupID: groupID,
		Search:  query.Search,
	})
	if err != nil {
		return MembersData{}, err
	}

	query = utils.ClampPage(query, totalItems)

	params := db.ListMembersByNameAscFilteredParams{
		GroupID: groupID,
		Search:  query.Search,
		Limit:   int64(query.PageSize),
		Offset:  query.Offset(),
	}

	var members []db.Member
	switch query.Sort {
	case "name":
		if query.Dir == "desc" {
			members, err = db.Qry.ListMembersByNameDescFiltered(ctx, db.ListMembersByNameDescFilteredParams(params))
		} else {
			members, err = db.Qry.ListMembersByNameAscFiltered(ctx, params)
		}
	default:
		if query.Dir == "asc" {
			members, err = db.Qry.ListMembersByCreatedAtAscFiltered(ctx, db.ListMembersByCreatedAtAscFilteredParams(params))
		} else {
			members, err = db.Qry.ListMembersByCreatedAtDescFiltered(ctx, db.ListMembersByCreatedAtDescFilteredParams(params))
		}
	}
	if err != nil {
		return MembersData{}, err
	}
	return MembersData{
		Title:   ctxi18n.T(ctx, "members.title"),
		Members: members,
		Query:   query,
		Pager:   utils.BuildTablePagination(totalItems, query),
		GroupID: groupID,
		Breadcrumbs: []utils.Crumb{
			{Label: ctxi18n.T(ctx, "groups.title"), Href: "/dashboard"},
			{Label: group.Name, Href: "/groups/" + groupID},
			{Label: ctxi18n.T(ctx, "members.title")},
		},
	}, nil
}
