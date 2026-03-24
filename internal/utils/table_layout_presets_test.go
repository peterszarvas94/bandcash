package utils

import "testing"

func TestEventParticipantsTableLayoutColumnOrder(t *testing.T) {
	layout := EventParticipantsTableLayout()
	want := []string{"name", "amount", "expense", "total", "paid", "paid_at", "note"}

	if len(layout.ColumnOrder) != len(want) {
		t.Fatalf("expected %d columns, got %d (%v)", len(want), len(layout.ColumnOrder), layout.ColumnOrder)
	}

	for i := range want {
		if layout.ColumnOrder[i] != want[i] {
			t.Fatalf("unexpected column order at %d: got %q want %q", i, layout.ColumnOrder[i], want[i])
		}
	}
}
