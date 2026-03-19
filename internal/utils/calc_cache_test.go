package utils

import (
	"strings"
	"testing"
)

func TestCacheKeysAreGroupScopedForPrefixInvalidation(t *testing.T) {
	groupID := "grp_test"

	groupTotalsKey := GroupTotalsKey(groupID)
	eventsKey := EventsFilterKey(groupID, "search", "2026", "2026-01-01", "2026-12-31", "time", "desc")
	expensesKey := ExpensesFilterKey(groupID, "search", "2026", "2026-01-01", "2026-12-31", "date", "desc")

	if want := "group_totals_group_" + groupID; !strings.HasPrefix(groupTotalsKey, want) {
		t.Fatalf("GroupTotalsKey prefix mismatch: got %q, want prefix %q", groupTotalsKey, want)
	}
	if want := "events_group_" + groupID + "_"; !strings.HasPrefix(eventsKey, want) {
		t.Fatalf("EventsFilterKey prefix mismatch: got %q, want prefix %q", eventsKey, want)
	}
	if want := "expenses_group_" + groupID + "_"; !strings.HasPrefix(expensesKey, want) {
		t.Fatalf("ExpensesFilterKey prefix mismatch: got %q, want prefix %q", expensesKey, want)
	}

	if !strings.Contains(eventsKey, "search_search") || !strings.Contains(eventsKey, "year_2026") {
		t.Fatalf("EventsFilterKey should be readable, got %q", eventsKey)
	}
	if !strings.Contains(expensesKey, "from_2026-01-01") || !strings.Contains(expensesKey, "to_2026-12-31") {
		t.Fatalf("ExpensesFilterKey should be readable, got %q", expensesKey)
	}
}

func TestInvalidateGroupCachesClearsOnlyTargetGroup(t *testing.T) {
	cache := NewCalcCache()
	CalcCacheInstance = cache
	t.Cleanup(func() {
		CalcCacheInstance = NewCalcCache()
	})

	groupA := "grp_a"
	groupB := "grp_b"

	cache.Set(GroupTotalsKey(groupA), GroupTotals{TotalEventAmount: 1})
	cache.Set(EventsFilterKey(groupA, "", "", "", "", "", ""), 1)
	cache.Set(ExpensesFilterKey(groupA, "", "", "", "", "", ""), 1)

	cache.Set(GroupTotalsKey(groupB), GroupTotals{TotalEventAmount: 2})
	cache.Set(EventsFilterKey(groupB, "", "", "", "", "", ""), 2)
	cache.Set(ExpensesFilterKey(groupB, "", "", "", "", "", ""), 2)

	InvalidateGroupCaches(groupA)

	if _, ok := cache.Get(GroupTotalsKey(groupA)); ok {
		t.Fatal("expected group A group_totals cache to be cleared")
	}
	if _, ok := cache.Get(EventsFilterKey(groupA, "", "", "", "", "", "")); ok {
		t.Fatal("expected group A events cache to be cleared")
	}
	if _, ok := cache.Get(ExpensesFilterKey(groupA, "", "", "", "", "", "")); ok {
		t.Fatal("expected group A expenses cache to be cleared")
	}

	if _, ok := cache.Get(GroupTotalsKey(groupB)); !ok {
		t.Fatal("expected group B group_totals cache to remain")
	}
	if _, ok := cache.Get(EventsFilterKey(groupB, "", "", "", "", "", "")); !ok {
		t.Fatal("expected group B events cache to remain")
	}
	if _, ok := cache.Get(ExpensesFilterKey(groupB, "", "", "", "", "", "")); !ok {
		t.Fatal("expected group B expenses cache to remain")
	}
}
