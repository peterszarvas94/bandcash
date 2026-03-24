package event

import (
	"database/sql"
	"testing"

	"bandcash/internal/db"
	"bandcash/internal/utils"
)

func TestEventTableSpecs(t *testing.T) {
	e := New()

	mainSpec := e.TableQuerySpec()
	if mainSpec.DefaultSort != "time" || mainSpec.DefaultDir != "desc" {
		t.Fatalf("unexpected events table defaults: %+v", mainSpec)
	}
	for _, key := range []string{"time", "title", "amount", "description", "paid", "paid_at"} {
		if _, ok := mainSpec.AllowedSorts[key]; !ok {
			t.Fatalf("expected allowed sort %q in events spec", key)
		}
	}

	participantsSpec := e.ParticipantTableQuerySpec()
	if participantsSpec.DefaultSort != "name" || participantsSpec.DefaultDir != "asc" {
		t.Fatalf("unexpected participant table defaults: %+v", participantsSpec)
	}
	for _, key := range []string{"name", "amount", "expense", "total", "paid", "paid_at"} {
		if _, ok := participantsSpec.AllowedSorts[key]; !ok {
			t.Fatalf("expected allowed sort %q in participant spec", key)
		}
	}
}

func TestMatchesFilters(t *testing.T) {
	event := db.Event{Title: "Band Rehearsal", Description: "Weekly practice", Time: "2026-03-12T19:00"}

	if !matchesFilters(event, utils.TableQuery{}) {
		t.Fatal("expected empty query to match")
	}
	if matchesFilters(event, utils.TableQuery{Search: "missing"}) {
		t.Fatal("expected missing search to fail")
	}
	if !matchesFilters(event, utils.TableQuery{Search: "weekly"}) {
		t.Fatal("expected description search to match")
	}
	if matchesFilters(event, utils.TableQuery{Year: "2025"}) {
		t.Fatal("expected wrong year to fail")
	}
	if matchesFilters(event, utils.TableQuery{From: "2026-03-13"}) {
		t.Fatal("expected from-date after event to fail")
	}
	if matchesFilters(event, utils.TableQuery{To: "2026-03-11"}) {
		t.Fatal("expected to-date before event to fail")
	}
}

func TestSortEventsByPaidAt(t *testing.T) {
	events := []db.Event{
		{ID: "evt_1", Time: "2026-01-01T10:00", PaidAt: sql.NullString{String: "2026-02-02", Valid: true}},
		{ID: "evt_2", Time: "2026-01-02T10:00", PaidAt: sql.NullString{}},
		{ID: "evt_3", Time: "2026-01-03T10:00", PaidAt: sql.NullString{String: "2026-01-01", Valid: true}},
	}

	sortEvents(events, "paid_at", "asc")
	if events[0].ID != "evt_3" || events[1].ID != "evt_1" || events[2].ID != "evt_2" {
		t.Fatalf("unexpected asc paid_at order: %s, %s, %s", events[0].ID, events[1].ID, events[2].ID)
	}

	sortEvents(events, "paid_at", "desc")
	if events[0].ID != "evt_2" || events[1].ID != "evt_1" || events[2].ID != "evt_3" {
		t.Fatalf("unexpected desc paid_at order: %s, %s, %s", events[0].ID, events[1].ID, events[2].ID)
	}
}
