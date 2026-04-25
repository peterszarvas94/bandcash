package billing

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"bandcash/internal/db"
	"bandcash/internal/utils"
)

const (
	TierFree = "free"
	TierPro  = "pro"
)

const PastDueGracePeriod = 7 * 24 * time.Hour

var ErrInvalidUserID = errors.New("invalid user id")

type AccessState struct {
	SubscriptionCount int
	OwnedGroupCount   int
	RemainingSlots    int
}

func IsLimitExceeded(state AccessState) bool {
	return state.OwnedGroupCount > state.SubscriptionCount
}

func HasAvailableGroupSlot(state AccessState) bool {
	return state.OwnedGroupCount < state.SubscriptionCount
}

func RemainingGroupSlots(state AccessState) int {
	if state.RemainingSlots < 0 {
		return 0
	}
	return state.RemainingSlots
}

func IsSubscriptionActive(status string, graceUntil sql.NullTime, now time.Time) bool {
	status = strings.ToLower(strings.TrimSpace(status))
	switch status {
	case "active", "trialing":
		return true
	case "past_due":
		return graceUntil.Valid && now.Before(graceUntil.Time)
	default:
		return false
	}
}

func IsSupportedSubscriptionPrice(priceID string) bool {
	return strings.TrimSpace(priceID) != ""
}

func TierFromPriceID(priceID string) string {
	if IsSupportedSubscriptionPrice(priceID) {
		return TierPro
	}
	return TierFree
}

func CountOwnedGroups(ctx context.Context, userID string, excludeGroupID string) (int, error) {
	q := db.BunDB.NewSelect().TableExpr("groups").Where("admin_user_id = ?", userID)
	if strings.TrimSpace(excludeGroupID) != "" {
		q = q.Where("id != ?", strings.TrimSpace(excludeGroupID))
	}
	return q.Count(ctx)
}

func GroupOwnerID(ctx context.Context, groupID string) (string, error) {
	var ownerID string
	err := db.BunDB.NewSelect().
		TableExpr("groups").
		Column("admin_user_id").
		Where("id = ?", groupID).
		Scan(ctx, &ownerID)
	return ownerID, err
}

func CountActiveSubscriptionSlots(ctx context.Context, userID string) (int, error) {
	rows := make([]db.BillingSubscription, 0)
	err := db.BunDB.NewSelect().
		Model(&rows).
		Where("user_id = ?", userID).
		Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, nil
		}
		return 0, err
	}

	now := time.Now().UTC()
	total := 0
	for _, row := range rows {
		if !IsSupportedSubscriptionPrice(row.ProviderVariantID) {
			continue
		}
		if IsSubscriptionActive(row.Status, row.GraceUntil, now) {
			seats := row.SeatQuantity
			if seats < 1 {
				seats = 1
			}
			total += seats
		}
	}
	return total, nil
}

func GetUserSubscription(ctx context.Context, userID string) (db.BillingSubscription, bool, error) {
	if !utils.IsValidID(userID, "usr") {
		return db.BillingSubscription{}, false, ErrInvalidUserID
	}

	var row db.BillingSubscription
	err := db.BunDB.NewSelect().
		Model(&row).
		Where("user_id = ?", userID).
		Limit(1).
		Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return db.BillingSubscription{}, false, nil
		}
		return db.BillingSubscription{}, false, err
	}
	if row.SeatQuantity < 1 {
		row.SeatQuantity = 1
	}
	return row, true, nil
}

func CurrentAccessState(ctx context.Context, userID string) (AccessState, error) {
	if !utils.IsValidID(userID, "usr") {
		return AccessState{}, ErrInvalidUserID
	}

	slots, err := CountActiveSubscriptionSlots(ctx, userID)
	if err != nil {
		return AccessState{}, err
	}
	ownedGroups, err := CountOwnedGroups(ctx, userID, "")
	if err != nil {
		return AccessState{}, err
	}

	remaining := slots - ownedGroups
	return AccessState{SubscriptionCount: slots, OwnedGroupCount: ownedGroups, RemainingSlots: remaining}, nil
}

func CanOwnAnotherGroup(ctx context.Context, userID string) (bool, AccessState, error) {
	state, err := CurrentAccessState(ctx, userID)
	if err != nil {
		return false, AccessState{}, err
	}
	return HasAvailableGroupSlot(state), state, nil
}

func CanCreateEventInGroup(ctx context.Context, groupID string) (bool, AccessState, error) {
	ownerID, err := GroupOwnerID(ctx, groupID)
	if err != nil {
		return false, AccessState{}, err
	}
	state, err := CurrentAccessState(ctx, ownerID)
	if err != nil {
		return false, AccessState{}, err
	}
	return true, state, nil
}

