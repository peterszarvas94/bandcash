package member

import (
	"context"

	ctxi18n "github.com/invopop/ctxi18n/i18n"

	"bandcash/internal/utils"
	groupstore "bandcash/models/group/data"
	memberstore "bandcash/models/member/data"
)

func TableQuerySpec() utils.TableQuerySpec {
	return utils.StandardTableQuerySpec(utils.StandardTableQuerySpecParams{
		DefaultSort:  "createdAt",
		DefaultDir:   "desc",
		AllowedSorts: []string{"name", "createdAt", "description"},
	})
}

func MemberEventsTableQuerySpec() utils.TableQuerySpec {
	return utils.StandardTableQuerySpec(utils.StandardTableQuerySpecParams{
		DefaultSort:  "time",
		DefaultDir:   "desc",
		AllowedSorts: []string{"title", "time", "participant_amount", "participant_expense", "paid", "paid_at"},
	})
}

func convertToMemberEvent(r memberstore.MemberEventRow) MemberEvent {
	return MemberEvent{
		ID:                 r.ID,
		GroupID:            r.GroupID,
		Title:              r.Title,
		Time:               r.Time,
		Description:        r.Description,
		Amount:             r.Amount,
		ParticipantAmount:  r.ParticipantAmount,
		ParticipantExpense: r.ParticipantExpense,
		ParticipantPaid:    r.ParticipantPaid,
		ParticipantPaidAt:  r.ParticipantPaidAt,
	}
}

func GetShowData(ctx context.Context, groupID, memberID string, query utils.TableQuery) (MemberData, error) {
	group, err := groupstore.GetGroupByID(ctx, groupID)
	if err != nil {
		return MemberData{}, err
	}

	member, err := memberstore.GetMember(ctx, memberstore.GetMemberParams{
		ID:      memberID,
		GroupID: groupID,
	})
	if err != nil {
		return MemberData{}, err
	}

	filter := memberstore.MemberEventFilter{
		MemberID: memberID,
		GroupID:  groupID,
		Search:   query.Search,
		Year:     query.Year,
		From:     query.From,
		To:       query.To,
	}
	totalItems, err := memberstore.CountMemberEventsTable(ctx, filter)
	if err != nil {
		return MemberData{}, err
	}

	totals, err := memberstore.SumMemberEventTotalsTable(ctx, filter)
	if err != nil {
		return MemberData{}, err
	}

	query = utils.ClampPage(query, int64(totalItems))

	rows, err := memberstore.ListMemberEventsTable(ctx, memberstore.MemberEventListParams{
		MemberEventFilter: filter,
		Sort:              query.Sort,
		Dir:               query.Dir,
		Limit:             query.PageSize,
		Offset:            int(query.Offset()),
	})
	if err != nil {
		return MemberData{}, err
	}
	events := make([]MemberEvent, 0, len(rows))
	for _, r := range rows {
		events = append(events, convertToMemberEvent(r))
	}

	return MemberData{
		Title:        "Bandcash - " + member.Name,
		Member:       &member,
		Events:       events,
		GroupID:      groupID,
		Query:        query,
		Pager:        utils.BuildTablePagination(int64(totalItems), query),
		RecentYears:  utils.RecentYears(3),
		TotalCut:     totals.TotalCut,
		TotalExpense: totals.TotalExpense,
		TotalPayout:  totals.TotalPayout,
		TotalPaid:    totals.TotalPaid,
		TotalUnpaid:  totals.TotalUnpaid,
		Breadcrumbs: []utils.Crumb{
			{Label: ctxi18n.T(ctx, "groups.title"), Href: "/groups"},
			{Label: group.Name, Href: "/groups/" + groupID + "/events"},
			{Label: ctxi18n.T(ctx, "members.title"), Href: "/groups/" + groupID + "/members"},
			{Label: member.Name},
		},
		EventsTable: MemberEventsTableLayout(),
	}, nil
}

func GetIndexData(ctx context.Context, groupID string, query utils.TableQuery) (MembersData, error) {
	group, err := groupstore.GetGroupByID(ctx, groupID)
	if err != nil {
		return MembersData{}, err
	}

	totalItems, err := memberstore.CountMembersTable(ctx, groupID, query.Search)
	if err != nil {
		return MembersData{}, err
	}

	query = utils.ClampPage(query, totalItems)

	members, err := memberstore.ListMembersTable(ctx, memberstore.MemberTableListParams{
		GroupID: groupID,
		Search:  query.Search,
		Sort:    query.Sort,
		Dir:     query.Dir,
		Limit:   query.PageSize,
		Offset:  int(query.Offset()),
	})
	if err != nil {
		return MembersData{}, err
	}
	return MembersData{
		Title:     ctxi18n.T(ctx, "members.page_title"),
		GroupName: group.Name,
		Members:   members,
		Query:     query,
		Pager:     utils.BuildTablePagination(totalItems, query),
		GroupID:   groupID,
		Breadcrumbs: []utils.Crumb{
			{Label: ctxi18n.T(ctx, "groups.title"), Href: "/groups"},
			{Label: group.Name, Href: "/groups/" + groupID + "/events"},
			{Label: ctxi18n.T(ctx, "members.title")},
		},
		MembersTable: MembersIndexTableLayout(),
	}, nil
}
