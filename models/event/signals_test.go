package event

import (
	"testing"

	"bandcash/internal/db"
	"bandcash/internal/utils"
)

func TestEventShowSignalsInitializesParticipantNoteDialogState(t *testing.T) {
	data := EventData{
		Event: &db.Event{ID: "evt_1", Title: "Event", Date: "2026-01-01", EventTime: "10:00", Amount: 100},
		Query: utils.TableQuery{Summary: utils.SummaryModeAll},
	}

	signals := eventShowSignals(data)
	raw, ok := signals["participantNoteDialog"]
	if !ok {
		t.Fatalf("expected participantNoteDialog signal to be present")
	}

	dialog, ok := raw.(map[string]any)
	if !ok {
		t.Fatalf("expected participantNoteDialog to be map[string]any, got %T", raw)
	}

	if open, ok := dialog["open"].(bool); !ok || open {
		t.Fatalf("expected open=false, got %#v", dialog["open"])
	}
	if readOnly, ok := dialog["readOnly"].(bool); !ok || readOnly {
		t.Fatalf("expected readOnly=false, got %#v", dialog["readOnly"])
	}
	if memberID, ok := dialog["memberId"].(string); !ok || memberID != "" {
		t.Fatalf("expected empty memberId, got %#v", dialog["memberId"])
	}
}
