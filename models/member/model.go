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
	return utils.StandardTableQuerySpec(utils.StandardTableQuerySpecParams{
		DefaultSort:  "createdAt",
		DefaultDir:   "desc",
		AllowedSorts: []string{"name", "createdAt", "description"},
	})
}

func New() *Members {
	return &Members{}
}

func (p *Members) MemberEventsTableQuerySpec() utils.TableQuerySpec {
	return utils.StandardTableQuerySpec(utils.StandardTableQuerySpecParams{
		DefaultSort:  "time",
		DefaultDir:   "desc",
		AllowedSorts: []string{"title", "time", "participant_amount", "participant_expense", "paid", "paid_at"},
	})
}

func convertToMemberEvent(row interface{}) MemberEvent {
	switch r := row.(type) {
	case db.ListParticipantsByMemberByTitleAscFilteredRow:
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
	case db.ListParticipantsByMemberByTitleDescFilteredRow:
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
	case db.ListParticipantsByMemberByTimeAscFilteredRow:
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
	case db.ListParticipantsByMemberByTimeDescFilteredRow:
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
	case db.ListParticipantsByMemberByAmountAscFilteredRow:
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
	case db.ListParticipantsByMemberByAmountDescFilteredRow:
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
	case db.ListParticipantsByMemberByCutAscFilteredRow:
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
	case db.ListParticipantsByMemberByCutDescFilteredRow:
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
	case db.ListParticipantsByMemberByExpenseAscFilteredRow:
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
	case db.ListParticipantsByMemberByExpenseDescFilteredRow:
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
	case db.ListParticipantsByMemberByPaidAscFilteredRow:
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
	case db.ListParticipantsByMemberByPaidDescFilteredRow:
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
	case db.ListParticipantsByMemberByPaidAtAscFilteredRow:
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
	case db.ListParticipantsByMemberByPaidAtDescFilteredRow:
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
	return MemberEvent{}
}

func (p *Members) GetShowData(ctx context.Context, groupID, memberID string, query utils.TableQuery) (MemberData, error) {
	group, err := db.GetGroupByID(ctx, groupID)
	if err != nil {
		return MemberData{}, err
	}

	member, err := db.GetMember(ctx, db.GetMemberParams{
		ID:      memberID,
		GroupID: groupID,
	})
	if err != nil {
		return MemberData{}, err
	}

	totalItems, err := db.CountParticipantsByMemberFiltered(ctx, db.CountParticipantsByMemberFilteredParams{
		MemberID: memberID,
		GroupID:  groupID,
		Search:   query.Search,
		Year:     query.Year,
		From:     query.From,
		To:       query.To,
	})
	if err != nil {
		return MemberData{}, err
	}

	totals, err := db.SumParticipantTotalsByMemberFiltered(ctx, db.SumParticipantTotalsByMemberFilteredParams{
		MemberID: memberID,
		GroupID:  groupID,
		Search:   query.Search,
		Year:     query.Year,
		From:     query.From,
		To:       query.To,
	})
	if err != nil {
		return MemberData{}, err
	}

	query = utils.ClampPage(query, int64(totalItems))

	params := db.ListParticipantsByMemberByTimeDescFilteredParams{
		MemberID: memberID,
		GroupID:  groupID,
		Search:   query.Search,
		Year:     query.Year,
		From:     query.From,
		To:       query.To,
		Limit:    int64(query.PageSize),
		Offset:   query.Offset(),
	}

	var events []MemberEvent
	switch query.Sort {
	case "title":
		if query.Dir == "desc" {
			rows, err := db.ListParticipantsByMemberByTitleDescFiltered(ctx, db.ListParticipantsByMemberByTitleDescFilteredParams(params))
			if err != nil {
				return MemberData{}, err
			}
			for _, r := range rows {
				events = append(events, convertToMemberEvent(r))
			}
		} else {
			rows, err := db.ListParticipantsByMemberByTitleAscFiltered(ctx, db.ListParticipantsByMemberByTitleAscFilteredParams(params))
			if err != nil {
				return MemberData{}, err
			}
			for _, r := range rows {
				events = append(events, convertToMemberEvent(r))
			}
		}
	case "participant_amount":
		if query.Dir == "desc" {
			rows, err := db.ListParticipantsByMemberByCutDescFiltered(ctx, db.ListParticipantsByMemberByCutDescFilteredParams(params))
			if err != nil {
				return MemberData{}, err
			}
			for _, r := range rows {
				events = append(events, convertToMemberEvent(r))
			}
		} else {
			rows, err := db.ListParticipantsByMemberByCutAscFiltered(ctx, db.ListParticipantsByMemberByCutAscFilteredParams(params))
			if err != nil {
				return MemberData{}, err
			}
			for _, r := range rows {
				events = append(events, convertToMemberEvent(r))
			}
		}
	case "participant_expense":
		if query.Dir == "desc" {
			rows, err := db.ListParticipantsByMemberByExpenseDescFiltered(ctx, db.ListParticipantsByMemberByExpenseDescFilteredParams(params))
			if err != nil {
				return MemberData{}, err
			}
			for _, r := range rows {
				events = append(events, convertToMemberEvent(r))
			}
		} else {
			rows, err := db.ListParticipantsByMemberByExpenseAscFiltered(ctx, db.ListParticipantsByMemberByExpenseAscFilteredParams(params))
			if err != nil {
				return MemberData{}, err
			}
			for _, r := range rows {
				events = append(events, convertToMemberEvent(r))
			}
		}
	case "paid":
		if query.Dir == "desc" {
			rows, err := db.ListParticipantsByMemberByPaidDescFiltered(ctx, db.ListParticipantsByMemberByPaidDescFilteredParams(params))
			if err != nil {
				return MemberData{}, err
			}
			for _, r := range rows {
				events = append(events, convertToMemberEvent(r))
			}
		} else {
			rows, err := db.ListParticipantsByMemberByPaidAscFiltered(ctx, db.ListParticipantsByMemberByPaidAscFilteredParams(params))
			if err != nil {
				return MemberData{}, err
			}
			for _, r := range rows {
				events = append(events, convertToMemberEvent(r))
			}
		}
	case "paid_at":
		if query.Dir == "desc" {
			rows, err := db.ListParticipantsByMemberByPaidAtDescFiltered(ctx, db.ListParticipantsByMemberByPaidAtDescFilteredParams(params))
			if err != nil {
				return MemberData{}, err
			}
			for _, r := range rows {
				events = append(events, convertToMemberEvent(r))
			}
		} else {
			rows, err := db.ListParticipantsByMemberByPaidAtAscFiltered(ctx, db.ListParticipantsByMemberByPaidAtAscFilteredParams(params))
			if err != nil {
				return MemberData{}, err
			}
			for _, r := range rows {
				events = append(events, convertToMemberEvent(r))
			}
		}
	default:
		if query.Dir == "asc" {
			rows, err := db.ListParticipantsByMemberByTimeAscFiltered(ctx, db.ListParticipantsByMemberByTimeAscFilteredParams(params))
			if err != nil {
				return MemberData{}, err
			}
			for _, r := range rows {
				events = append(events, convertToMemberEvent(r))
			}
		} else {
			rows, err := db.ListParticipantsByMemberByTimeDescFiltered(ctx, params)
			if err != nil {
				return MemberData{}, err
			}
			for _, r := range rows {
				events = append(events, convertToMemberEvent(r))
			}
		}
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

func (p *Members) GetIndexData(ctx context.Context, groupID string, query utils.TableQuery) (MembersData, error) {
	group, err := db.GetGroupByID(ctx, groupID)
	if err != nil {
		return MembersData{}, err
	}

	totalItems, err := db.CountMembersTable(ctx, groupID, query.Search)
	if err != nil {
		return MembersData{}, err
	}

	query = utils.ClampPage(query, totalItems)

	members, err := db.ListMembersTable(ctx, db.MemberTableListParams{
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
