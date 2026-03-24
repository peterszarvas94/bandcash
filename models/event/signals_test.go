package event

import (
	"testing"

	"bandcash/internal/db"
	"bandcash/internal/utils"
)

func TestEventShowSignalsInitializesNoteExpandedPerParticipant(t *testing.T) {
	data := EventData{
		Event: &db.Event{ID: "evt_1", Title: "Event", Time: "2026-01-01T10:00", Amount: 100},
		Query: utils.TableQuery{Summary: utils.SummaryModeAll},
		Participants: []db.ListParticipantsByEventRow{
			{ID: "mem_1"},
			{ID: "mem_2"},
		},
	}

	signals := eventShowSignals(data, "csrf-token")
	raw, ok := signals["noteExpanded"]
	if !ok {
		t.Fatalf("expected noteExpanded signal to be present")
	}

	noteExpanded, ok := raw.(map[string]bool)
	if !ok {
		t.Fatalf("expected noteExpanded to be map[string]bool, got %T", raw)
	}

	if len(noteExpanded) != 2 {
		t.Fatalf("expected 2 noteExpanded entries, got %d", len(noteExpanded))
	}
	if noteExpanded["mem_1"] || noteExpanded["mem_2"] {
		t.Fatalf("expected all noteExpanded values to default to false, got %+v", noteExpanded)
	}
}
