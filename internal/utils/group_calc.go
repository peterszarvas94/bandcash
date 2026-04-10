package utils

import (
	"context"
	"log/slog"

	eventstore "bandcash/models/event/data"
	expensestore "bandcash/models/expense/data"
)

// CalculateGroupTotals computes all financial totals for a group in-memory
// respecting paid/unpaid status. Results are cached.
func CalculateGroupTotals(ctx context.Context, groupID string) (GroupTotals, error) {
	// Check cache first
	cacheKey := GroupTotalsCacheKey(groupID)
	if cached, ok := CalcCacheInstance.Get(cacheKey); ok {
		if totals, valid := cached.(GroupTotals); valid {
			return totals, nil
		}
	}

	totals := GroupTotals{}

	// Calculate from events
	events, err := eventstore.ListEvents(ctx, groupID)
	if err != nil {
		slog.Error("failed to list events for totals", "group_id", groupID, "err", err)
		return totals, err
	}
	for _, event := range events {
		totals.Income.All += event.Amount
		if event.Paid == 1 {
			totals.Income.Paid += event.Amount
		} else {
			totals.Income.Unpaid += event.Amount
		}
	}

	// Calculate from expenses
	expenses, err := expensestore.ListExpenses(ctx, groupID)
	if err != nil {
		slog.Error("failed to list expenses for totals", "group_id", groupID, "err", err)
		return totals, err
	}
	for _, expense := range expenses {
		totals.Expenses.All += expense.Amount
		if expense.Paid == 1 {
			totals.Expenses.Paid += expense.Amount
		} else {
			totals.Expenses.Unpaid += expense.Amount
		}
	}

	// Calculate from participants
	payoutTotals, err := eventstore.SumParticipantPaidAmountsByGroup(ctx, groupID)
	if err != nil {
		slog.Error("failed to sum participants for totals", "group_id", groupID, "err", err)
		return totals, err
	}
	totals.Payouts.Paid = payoutTotals.PaidAmount
	totals.Payouts.Unpaid = payoutTotals.UnpaidAmount
	totals.Payouts.All = payoutTotals.PaidAmount + payoutTotals.UnpaidAmount

	// Calculate balance: events - expenses - payouts
	totals.Balance.Paid = totals.Income.Paid - totals.Expenses.Paid - totals.Payouts.Paid
	totals.Balance.Unpaid = totals.Income.Unpaid - totals.Expenses.Unpaid - totals.Payouts.Unpaid
	totals.Balance.All = totals.Income.All - totals.Expenses.All - totals.Payouts.All

	// Cache the result
	CalcCacheInstance.Set(cacheKey, totals)
	slog.Debug("calculated group totals", "group_id", groupID, "totals", totals.String())

	return totals, nil
}

// InvalidateGroupTotals clears the cached totals for a group
func InvalidateGroupTotals(groupID string) {
	CalcCacheInstance.ClearPrefix(GroupTotalsCachePrefix(groupID))
}

// InvalidateGroupCaches clears all cached calculations for a group
// Call this when any event, expense, or participant changes
func InvalidateGroupCaches(groupID string) {
	CalcCacheInstance.ClearPrefix(GroupTotalsCachePrefix(groupID))
	CalcCacheInstance.ClearPrefix(EventsCachePrefix(groupID))
	CalcCacheInstance.ClearPrefix(ExpensesCachePrefix(groupID))
}
