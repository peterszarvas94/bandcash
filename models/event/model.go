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

func (e *Events) GetShowData(ctx context.Context, id int) (EventData, error) {
	event, err := db.Qry.GetEvent(ctx, int64(id))
	if err != nil {
		return EventData{}, err
	}

	participants, err := db.Qry.ListParticipantsByEvent(ctx, int64(id))
	if err != nil {
		return EventData{}, err
	}

	members, err := db.Qry.ListMembers(ctx)
	if err != nil {
		return EventData{}, err
	}

	memberIDs := make(map[int64]bool, len(participants))
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

	slog.Info("event.show.data", "event_id", id, "participants", len(participants), "members_total", len(members), "members_filtered", len(filteredMembers), "leftover", leftover)

	return EventData{
		Title:            event.Title,
		Event:            &event,
		Participants:     participants,
		Members:          filteredMembers,
		MemberIDs:        memberIDs,
		Leftover:         leftover,
		TotalDistributed: totalDistributed,
		Breadcrumbs: []utils.Crumb{
			{Label: ctxi18n.T(ctx, "events.title"), Href: "/event"},
			{Label: event.Title},
		},
	}, nil
}

func (e *Events) GetIndexData(ctx context.Context) (EventsData, error) {
	events, err := db.Qry.ListEvents(ctx)
	if err != nil {
		return EventsData{}, err
	}

	return EventsData{
		Title:  ctxi18n.T(ctx, "events.title"),
		Events: events,
		Breadcrumbs: []utils.Crumb{
			{Label: ctxi18n.T(ctx, "events.title")},
		},
	}, nil
}
