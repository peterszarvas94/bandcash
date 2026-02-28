package event

import (
	"context"
	"log/slog"

	ctxi18n "github.com/invopop/ctxi18n/i18n"

	"bandcash/internal/db"
	"bandcash/internal/utils"
)

type Events struct {
}

func (e *Events) TableQuerySpec() utils.TableQuerySpec {
	return utils.TableQuerySpec{
		DefaultSort: "time",
		DefaultDir:  "asc",
		AllowedSorts: map[string]struct{}{
			"time":   {},
			"title":  {},
			"amount": {},
		},
		AllowedPageSizes: map[int]struct{}{
			10:  {},
			50:  {},
			100: {},
			200: {},
		},
		DefaultSize:  50,
		MaxSearchLen: 100,
	}
}

func New() *Events {
	return &Events{}
}

func (e *Events) GetShowData(ctx context.Context, groupID, eventID string) (EventData, error) {
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

	participants, err := db.Qry.ListParticipantsByEvent(ctx, db.ListParticipantsByEventParams{
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

	memberIDs := make(map[string]bool, len(participants))
	for _, participant := range participants {
		memberIDs[participant.ID] = true
	}

	filteredMembers := make([]db.Member, 0, len(members))
	for _, member := range members {
		if memberIDs[member.ID] {
			continue
		}
		filteredMembers = append(filteredMembers, member)
	}

	// Calculate total distributed and leftover
	var totalDistributed int64
	for _, p := range participants {
		totalDistributed += p.ParticipantAmount + p.ParticipantExpense
	}
	leftover := event.Amount - totalDistributed

	slog.Info("event.show.data", "event_id", eventID, "participants", len(participants), "members_total", len(members), "members_filtered", len(filteredMembers), "leftover", leftover)

	return EventData{
		Title:            event.Title,
		Event:            &event,
		Participants:     participants,
		Members:          filteredMembers,
		MemberIDs:        memberIDs,
		Leftover:         leftover,
		TotalDistributed: totalDistributed,
		GroupID:          groupID,
		Breadcrumbs: []utils.Crumb{
			{Label: ctxi18n.T(ctx, "groups.title"), Href: "/dashboard"},
			{Label: group.Name, Href: "/groups/" + groupID},
			{Label: ctxi18n.T(ctx, "events.title"), Href: "/groups/" + groupID + "/events"},
			{Label: event.Title},
		},
	}, nil
}

func (e *Events) GetIndexData(ctx context.Context, groupID string, query utils.TableQuery) (EventsData, error) {
	group, err := db.Qry.GetGroupByID(ctx, groupID)
	if err != nil {
		return EventsData{}, err
	}

	totalItems, err := db.Qry.CountEventsFiltered(ctx, db.CountEventsFilteredParams{
		GroupID: groupID,
		Search:  query.Search,
	})
	if err != nil {
		return EventsData{}, err
	}

	query = utils.ClampPage(query, totalItems)

	params := db.ListEventsByTimeAscFilteredParams{
		GroupID: groupID,
		Search:  query.Search,
		Limit:   int64(query.PageSize),
		Offset:  query.Offset(),
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
		Title:   ctxi18n.T(ctx, "events.title"),
		Events:  events,
		Query:   query,
		Pager:   utils.BuildTablePagination(totalItems, query),
		GroupID: groupID,
		Breadcrumbs: []utils.Crumb{
			{Label: ctxi18n.T(ctx, "groups.title"), Href: "/dashboard"},
			{Label: group.Name, Href: "/groups/" + groupID},
			{Label: ctxi18n.T(ctx, "events.title")},
		},
	}, nil
}
