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
			{Label: group.Name, Href: "/groups/" + groupID + "/events"},
			{Label: event.Title},
		},
	}, nil
}

func (e *Events) GetIndexData(ctx context.Context, groupID string) (EventsData, error) {
	group, err := db.Qry.GetGroupByID(ctx, groupID)
	if err != nil {
		return EventsData{}, err
	}

	events, err := db.Qry.ListEvents(ctx, groupID)
	if err != nil {
		return EventsData{}, err
	}

	return EventsData{
		Title:     ctxi18n.T(ctx, "events.title"),
		Events:    events,
		GroupID:   groupID,
		GroupName: group.Name,
		Breadcrumbs: []utils.Crumb{
			{Label: ctxi18n.T(ctx, "groups.title"), Href: "/dashboard"},
			{Label: group.Name, Href: "/groups/" + groupID + "/events"},
			{Label: ctxi18n.T(ctx, "events.title")},
		},
	}, nil
}
