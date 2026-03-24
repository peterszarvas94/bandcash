package member

import (
	"database/sql"
	"testing"

	"bandcash/internal/db"
)

func TestMemberTableSpecs(t *testing.T) {
	m := New()

	mainSpec := m.TableQuerySpec()
	if mainSpec.DefaultSort != "createdAt" || mainSpec.DefaultDir != "desc" {
		t.Fatalf("unexpected members spec defaults: %+v", mainSpec)
	}
	for _, key := range []string{"name", "createdAt", "description"} {
		if _, ok := mainSpec.AllowedSorts[key]; !ok {
			t.Fatalf("expected allowed sort %q in members spec", key)
		}
	}

	eventsSpec := m.MemberEventsTableQuerySpec()
	if eventsSpec.DefaultSort != "time" || eventsSpec.DefaultDir != "desc" {
		t.Fatalf("unexpected member events spec defaults: %+v", eventsSpec)
	}
	for _, key := range []string{"title", "time", "participant_amount", "participant_expense", "paid", "paid_at"} {
		if _, ok := eventsSpec.AllowedSorts[key]; !ok {
			t.Fatalf("expected allowed sort %q in member events spec", key)
		}
	}
}

func TestConvertToMemberEvent(t *testing.T) {
	row := db.ListParticipantsByMemberByTitleAscFilteredRow{
		ID:                 "evt_1",
		GroupID:            "grp_1",
		Title:              "Event",
		Time:               "2026-01-01T10:00",
		Description:        "desc",
		Amount:             100,
		ParticipantAmount:  60,
		ParticipantExpense: 20,
		ParticipantPaid:    1,
		ParticipantPaidAt:  sql.NullString{String: "2026-01-02", Valid: true},
	}

	converted := convertToMemberEvent(row)
	if converted.ID != row.ID || converted.Title != row.Title {
		t.Fatalf("unexpected converted event: %+v", converted)
	}
	if converted.ParticipantPaidAt.String != "2026-01-02" || !converted.ParticipantPaidAt.Valid {
		t.Fatalf("expected paid_at to carry over, got %+v", converted.ParticipantPaidAt)
	}

	empty := convertToMemberEvent(struct{}{})
	if empty.ID != "" {
		t.Fatalf("expected unsupported row type to return zero value, got %+v", empty)
	}
}
