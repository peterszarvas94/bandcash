package event

import (
	"context"
	"log/slog"
	"sort"
	"strings"

	ctxi18n "github.com/invopop/ctxi18n/i18n"

	"bandcash/internal/db"
	"bandcash/internal/utils"
)

type Events struct {
}

func (e *Events) TableQuerySpec() utils.TableQuerySpec {
	return utils.StandardTableQuerySpec(utils.StandardTableQuerySpecParams{
		DefaultSort:  "time",
		DefaultDir:   "desc",
		AllowedSorts: []string{"time", "title", "amount", "description"},
	})
}

func New() *Events {
	return &Events{}
}

func (e *Events) ParticipantTableQuerySpec() utils.TableQuerySpec {
	return utils.StandardTableQuerySpec(utils.StandardTableQuerySpecParams{
		DefaultSort:  "name",
		DefaultDir:   "asc",
		AllowedSorts: []string{"name", "amount", "expense", "total"},
	})
}

func (e *Events) GetShowData(ctx context.Context, groupID, eventID string, query utils.TableQuery) (EventData, error) {
	group, err := db.Qry.GetGroupByID(ctx, groupID)
	if err != nil {
		return EventData{}, err
	}

	event, err := db.Qry.GetEvent(ctx, db.GetEventParams{
		ID:      eventID,
		GroupID: groupID,
	})
	if err != nil {
		return EventData{}, err
	}

	allParticipants, err := db.Qry.ListParticipantsByEvent(ctx, db.ListParticipantsByEventParams{
		EventID: eventID,
		GroupID: groupID,
	})
	if err != nil {
		return EventData{}, err
	}

	members, err := db.Qry.ListMembers(ctx, groupID)
	if err != nil {
		return EventData{}, err
	}

	participantMemberIDs := make(map[string]bool, len(allParticipants))
	for _, participant := range allParticipants {
		participantMemberIDs[participant.ID] = true
	}

	filteredMembers := make([]db.Member, 0, len(members))
	for _, member := range members {
		if participantMemberIDs[member.ID] {
			continue
		}
		filteredMembers = append(filteredMembers, member)
	}

	participants := allParticipants
	if query.Search != "" {
		search := strings.ToLower(strings.TrimSpace(query.Search))
		filtered := make([]db.ListParticipantsByEventRow, 0, len(allParticipants))
		for _, participant := range allParticipants {
			if strings.Contains(strings.ToLower(participant.Name), search) {
				filtered = append(filtered, participant)
			}
		}
		participants = filtered
	}

	sort.SliceStable(participants, func(i, j int) bool {
		left := participants[i]
		right := participants[j]

		less := false
		equal := false

		switch query.Sort {
		case "total":
			leftTotal := left.ParticipantAmount + left.ParticipantExpense
			rightTotal := right.ParticipantAmount + right.ParticipantExpense
			if leftTotal == rightTotal {
				leftName := strings.ToLower(left.Name)
				rightName := strings.ToLower(right.Name)
				less = leftName < rightName
				equal = leftName == rightName
			} else {
				less = leftTotal < rightTotal
			}
		case "amount":
			if left.ParticipantAmount == right.ParticipantAmount {
				leftName := strings.ToLower(left.Name)
				rightName := strings.ToLower(right.Name)
				less = leftName < rightName
				equal = leftName == rightName
			} else {
				less = left.ParticipantAmount < right.ParticipantAmount
			}
		case "expense":
			if left.ParticipantExpense == right.ParticipantExpense {
				leftName := strings.ToLower(left.Name)
				rightName := strings.ToLower(right.Name)
				less = leftName < rightName
				equal = leftName == rightName
			} else {
				less = left.ParticipantExpense < right.ParticipantExpense
			}
		default:
			leftName := strings.ToLower(left.Name)
			rightName := strings.ToLower(right.Name)
			less = leftName < rightName
			equal = leftName == rightName
		}

		if equal {
			return false
		}

		if query.Dir != "desc" {
			return less
		}
		return !less
	})

	totalItems := int64(len(participants))
	query = utils.ClampPage(query, totalItems)

	start := int(query.Offset())
	if start > len(participants) {
		start = len(participants)
	}
	end := start + query.PageSize
	if end > len(participants) {
		end = len(participants)
	}
	displayParticipants := participants[start:end]

	// Calculate total distributed and leftover
	var totalDistributed int64
	for _, p := range allParticipants {
		totalDistributed += p.ParticipantAmount + p.ParticipantExpense
	}
	leftover := event.Amount - totalDistributed

	slog.Info("event.show.data", "event_id", eventID, "participants", len(participants), "members_total", len(members), "members_filtered", len(filteredMembers), "leftover", leftover)

	return EventData{
		Title:            "Bandcash - " + event.Title,
		Event:            &event,
		Participants:     displayParticipants,
		Query:            query,
		Pager:            utils.BuildTablePagination(totalItems, query),
		Members:          filteredMembers,
		Leftover:         leftover,
		TotalDistributed: totalDistributed,
		GroupID:          groupID,
		Breadcrumbs: []utils.Crumb{
			{Label: ctxi18n.T(ctx, "groups.title"), Href: "/dashboard"},
			{Label: group.Name, Href: "/groups/" + groupID},
			{Label: ctxi18n.T(ctx, "events.title"), Href: "/groups/" + groupID + "/events"},
			{Label: event.Title},
		},
		ParticipantsTable: utils.EventParticipantsTableLayout(),
	}, nil
}

func (e *Events) GetIndexData(ctx context.Context, groupID string, query utils.TableQuery) (EventsData, error) {
	group, err := db.Qry.GetGroupByID(ctx, groupID)
	if err != nil {
		return EventsData{}, err
	}

	totalItems, err := db.Qry.CountEventsFiltered(ctx, db.CountEventsFilteredParams{
		GroupID:    groupID,
		Search:     query.Search,
		YearFilter: query.Year,
		FromDate:   query.From,
		ToDate:     query.To,
	})
	if err != nil {
		return EventsData{}, err
	}

	filteredTotal, err := db.Qry.SumEventsFiltered(ctx, db.SumEventsFilteredParams{
		GroupID:    groupID,
		Search:     query.Search,
		YearFilter: query.Year,
		FromDate:   query.From,
		ToDate:     query.To,
	})
	if err != nil {
		return EventsData{}, err
	}

	query = utils.ClampPage(query, totalItems)

	params := db.ListEventsByTimeAscFilteredParams{
		GroupID:    groupID,
		Search:     query.Search,
		YearFilter: query.Year,
		FromDate:   query.From,
		ToDate:     query.To,
		Limit:      int64(query.PageSize),
		Offset:     query.Offset(),
	}

	var events []db.Event
	switch query.Sort {
	case "title":
		if query.Dir == "desc" {
			events, err = db.Qry.ListEventsByTitleDescFiltered(ctx, db.ListEventsByTitleDescFilteredParams(params))
		} else {
			events, err = db.Qry.ListEventsByTitleAscFiltered(ctx, db.ListEventsByTitleAscFilteredParams(params))
		}
	case "amount":
		if query.Dir == "desc" {
			events, err = db.Qry.ListEventsByAmountDescFiltered(ctx, db.ListEventsByAmountDescFilteredParams(params))
		} else {
			events, err = db.Qry.ListEventsByAmountAscFiltered(ctx, db.ListEventsByAmountAscFilteredParams(params))
		}
	case "description":
		if query.Dir == "desc" {
			events, err = db.Qry.ListEventsByDescriptionDescFiltered(ctx, db.ListEventsByDescriptionDescFilteredParams(params))
		} else {
			events, err = db.Qry.ListEventsByDescriptionAscFiltered(ctx, db.ListEventsByDescriptionAscFilteredParams(params))
		}
	default:
		if query.Dir == "desc" {
			events, err = db.Qry.ListEventsByTimeDescFiltered(ctx, db.ListEventsByTimeDescFilteredParams(params))
		} else {
			events, err = db.Qry.ListEventsByTimeAscFiltered(ctx, params)
		}
	}
	if err != nil {
		return EventsData{}, err
	}

	return EventsData{
		Title:            ctxi18n.T(ctx, "events.page_title"),
		Events:           events,
		RecentYears:      utils.RecentYears(3),
		Query:            query,
		Pager:            utils.BuildTablePagination(totalItems, query),
		GroupID:          groupID,
		TotalEventAmount: group.TotalEventAmount,
		FilteredTotal:    filteredTotal,
		Breadcrumbs: []utils.Crumb{
			{Label: ctxi18n.T(ctx, "groups.title"), Href: "/dashboard"},
			{Label: group.Name, Href: "/groups/" + groupID},
			{Label: ctxi18n.T(ctx, "events.title")},
		},
		EventsTable: utils.EventsIndexTableLayout(),
	}, nil
}
