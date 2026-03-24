package event

import (
	"strings"
	"testing"

	"bandcash/internal/db"
)

func TestEventPaidAtArg(t *testing.T) {
	t.Run("unpaid always null", func(t *testing.T) {
		got := paidAtArg(false, "2026-04-12")
		if got.Valid {
			t.Fatalf("expected invalid null string for unpaid, got %+v", got)
		}
	})

	t.Run("paid with empty keeps explicit empty", func(t *testing.T) {
		got := paidAtArg(true, "")
		if !got.Valid || got.String != "" {
			t.Fatalf("expected valid empty string, got %+v", got)
		}
	})

	t.Run("paid with date uses normalized date", func(t *testing.T) {
		got := paidAtArg(true, " 2026-04-12 ")
		if !got.Valid || got.String != "2026-04-12" {
			t.Fatalf("expected valid normalized date, got %+v", got)
		}
	})
}

func TestMergeWizardRowsAppliesWizardMapsOnBaseRows(t *testing.T) {
	base := []ParticipantWizardRow{{
		RowID:    "row_1",
		MemberID: "mem_old",
		Note:     "old",
		PaidAt:   "2026-01-01",
	}}

	allMembers := []db.Member{
		{ID: "mem_old", Name: "Old Name"},
		{ID: "mem_new", Name: "New Name"},
	}

	got := mergeWizardRows(
		base,
		allMembers,
		nil,
		map[string]string{"row_1": " mem_new "},
		map[string]int64{"row_1": 111},
		map[string]int64{"row_1": 22},
		map[string]string{"row_1": "  updated note  "},
		map[string]bool{"row_1": true},
		map[string]string{"row_1": " 2026-05-02 "},
	)

	if len(got) != 1 {
		t.Fatalf("expected one row, got %d", len(got))
	}
	if got[0].MemberID != "mem_new" {
		t.Fatalf("expected memberID mem_new, got %q", got[0].MemberID)
	}
	if got[0].MemberName != "New Name" {
		t.Fatalf("expected memberName New Name, got %q", got[0].MemberName)
	}
	if got[0].Note != "updated note" {
		t.Fatalf("expected trimmed note, got %q", got[0].Note)
	}
	if got[0].Amount != 111 || got[0].Expense != 22 {
		t.Fatalf("unexpected amount/expense: %d/%d", got[0].Amount, got[0].Expense)
	}
	if got[0].PaidAt != "2026-05-02" {
		t.Fatalf("expected normalized paidAt, got %q", got[0].PaidAt)
	}
}

func TestMergeWizardRowsFromIncomingAppliesNoteOverride(t *testing.T) {
	incoming := []participantBulkRowData{{
		RowID:    "row_2",
		MemberID: "mem_1",
		Note:     "incoming",
	}}

	allMembers := []db.Member{{ID: "mem_2", Name: "Member Two"}}

	got := mergeWizardRows(
		nil,
		allMembers,
		incoming,
		map[string]string{"row_2": "mem_2"},
		nil,
		nil,
		map[string]string{"row_2": "  from map  "},
		nil,
		nil,
	)

	if len(got) != 1 {
		t.Fatalf("expected one row, got %d", len(got))
	}
	if got[0].MemberID != "mem_2" || got[0].MemberName != "Member Two" {
		t.Fatalf("expected member remapped to mem_2/Member Two, got %q/%q", got[0].MemberID, got[0].MemberName)
	}
	if got[0].Note != "from map" {
		t.Fatalf("expected note override from map, got %q", got[0].Note)
	}
	if !strings.HasPrefix(got[0].RowID, "row_") {
		t.Fatalf("expected row ID to be preserved, got %q", got[0].RowID)
	}
}
