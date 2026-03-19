package utils

import (
	"strings"
	"testing"
)

func TestCacheKeysAreGroupScopedForPrefixInvalidation(t *testing.T) {
	groupID := "grp_test"

	groupTotalsKey := GroupTotalsCacheKey(groupID)
	eventsKey := EventsCachePrefix(groupID) + "search_search_year_2026_from_2026-01-01_to_2026-12-31_sort_time_dir_desc"
	expensesKey := ExpensesCachePrefix(groupID) + "search_search_year_2026_from_2026-01-01_to_2026-12-31_sort_date_dir_desc"

	if want := "group_totals_group_" + groupID; !strings.HasPrefix(groupTotalsKey, want) {
		t.Fatalf("GroupTotalsCacheKey prefix mismatch: got %q, want prefix %q", groupTotalsKey, want)
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

	cache.Set(GroupTotalsCacheKey(groupA), GroupTotals{TotalEventAmount: 1})
	cache.Set(EventsCachePrefix(groupA)+"x", 1)
	cache.Set(ExpensesCachePrefix(groupA)+"x", 1)

	cache.Set(GroupTotalsCacheKey(groupB), GroupTotals{TotalEventAmount: 2})
	cache.Set(EventsCachePrefix(groupB)+"x", 2)
	cache.Set(ExpensesCachePrefix(groupB)+"x", 2)

	InvalidateGroupCaches(groupA)

	if _, ok := cache.Get(GroupTotalsCacheKey(groupA)); ok {
		t.Fatal("expected group A group_totals cache to be cleared")
	}
	if _, ok := cache.Get(EventsCachePrefix(groupA) + "x"); ok {
		t.Fatal("expected group A events cache to be cleared")
	}
	if _, ok := cache.Get(ExpensesCachePrefix(groupA) + "x"); ok {
		t.Fatal("expected group A expenses cache to be cleared")
	}

	if _, ok := cache.Get(GroupTotalsCacheKey(groupB)); !ok {
		t.Fatal("expected group B group_totals cache to remain")
	}
	if _, ok := cache.Get(EventsCachePrefix(groupB) + "x"); !ok {
		t.Fatal("expected group B events cache to remain")
	}
	if _, ok := cache.Get(ExpensesCachePrefix(groupB) + "x"); !ok {
		t.Fatal("expected group B expenses cache to remain")
	}
}
