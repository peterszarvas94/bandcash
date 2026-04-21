package billing

import (
	"context"
	"fmt"
	"strings"

	"bandcash/internal/db"
)

type StartupSyncReport struct {
	Candidates int
	Synced     int
	Failed     int
}

func StartupSyncSubscriptions(ctx context.Context) (StartupSyncReport, error) {
	type userRow struct {
		UserID string `bun:"user_id"`
	}

	rows := make([]userRow, 0)
	if err := db.BunDB.NewSelect().
		TableExpr("billing_subscriptions").
		ColumnExpr("DISTINCT user_id").
		Scan(ctx, &rows); err != nil {
		return StartupSyncReport{}, err
	}

	report := StartupSyncReport{Candidates: len(rows)}
	if report.Candidates == 0 {
		return report, nil
	}

	var firstErr error
	for _, row := range rows {
		userID := strings.TrimSpace(row.UserID)
		if userID == "" {
			report.Failed++
			if firstErr == nil {
				firstErr = fmt.Errorf("empty user id in billing_subscriptions")
			}
			continue
		}

		if _, _, err := SyncSubscriptionFromProvider(ctx, userID); err != nil {
			report.Failed++
			if firstErr == nil {
				firstErr = err
			}
			continue
		}
		report.Synced++
	}

	if report.Failed > 0 {
		return report, fmt.Errorf("startup subscription sync had %d failures: %w", report.Failed, firstErr)
	}

	return report, nil
}
