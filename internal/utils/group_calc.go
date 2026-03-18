package utils

import (
	"context"
	"log/slog"

	"bandcash/internal/db"
)

// CalculateGroupTotals computes all financial totals for a group in-memory
// respecting paid/unpaid status. Results are cached.
func CalculateGroupTotals(ctx context.Context, groupID string) (GroupTotals, error) {
	// Check cache first
	cacheKey := GroupTotalsKey(groupID)
	if cached, ok := CalcCacheInstance.Get(cacheKey); ok {
		if totals, valid := cached.(GroupTotals); valid {
			return totals, nil
		}
	}

	totals := GroupTotals{}

	// Calculate from events
	events, err := db.Qry.ListEvents(ctx, groupID)
	if err != nil {
		slog.Error("failed to list events for totals", "group_id", groupID, "err", err)
		return totals, err
	}
	for _, event := range events {
		totals.TotalEventAmount += event.Amount
		if event.Paid == 1 {
			totals.EventPaid += event.Amount
		} else {
			totals.EventUnpaid += event.Amount
		}
	}

	// Calculate from expenses
	expenses, err := db.Qry.ListExpenses(ctx, groupID)
	if err != nil {
		slog.Error("failed to list expenses for totals", "group_id", groupID, "err", err)
		return totals, err
	}
	for _, expense := range expenses {
		totals.TotalExpenseAmount += expense.Amount
		if expense.Paid == 1 {
			totals.ExpensePaid += expense.Amount
		} else {
			totals.ExpenseUnpaid += expense.Amount
		}
	}

	// Calculate from participants
	payoutSum, err := db.Qry.SumParticipantAmountsByGroup(ctx, groupID)
	if err != nil {
		slog.Error("failed to sum participants for totals", "group_id", groupID, "err", err)
		return totals, err
	}
	totals.TotalPayoutAmount = payoutSum

	// Calculate leftover: events - expenses - payouts
	totals.TotalLeftover = totals.TotalEventAmount - totals.TotalExpenseAmount - totals.TotalPayoutAmount

	// Cache the result
	CalcCacheInstance.Set(cacheKey, totals)
	slog.Debug("calculated group totals", "group_id", groupID, "totals", totals.String())

	return totals, nil
}

// InvalidateGroupTotals clears the cached totals for a group
func InvalidateGroupTotals(groupID string) {
	CalcCacheInstance.ClearPrefix("group_totals:" + groupID)
}

// InvalidateGroupCaches clears all cached calculations for a group
// Call this when any event, expense, or participant changes
func InvalidateGroupCaches(groupID string) {
	CalcCacheInstance.ClearPrefix("group_totals:" + groupID)
	CalcCacheInstance.ClearPrefix("events:" + groupID)
	CalcCacheInstance.ClearPrefix("expenses:" + groupID)
}
