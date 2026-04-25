package flags

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"bandcash/internal/db"
	"bandcash/internal/utils"
)

const EnableSignupKey = "enable_signup"
const EnablePaymentsKey = "enable_payments"
const AllowBypassLimitForSuperadminKey = "allow_bypass_limit_for_superadmin"

func GetBool(ctx context.Context, key string) (bool, error) {
	var row db.AppFlag
	err := db.BunDB.NewSelect().Model(&row).Where("key = ?", key).Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return row.BoolValue != 0, nil
}

func UpsertBool(ctx context.Context, key string, enabled bool) error {
	value := int64(0)
	if enabled {
		value = 1
	}
	_, err := db.BunDB.NewInsert().
		Model(&db.AppFlag{Key: key, BoolValue: value}).
		On("CONFLICT(key) DO UPDATE").
		Set("bool_value = EXCLUDED.bool_value").
		Set("updated_at = CURRENT_TIMESTAMP").
		Exec(ctx)
	return err
}

// IsSignupEnabled reports the enable_signup flag, or (when email is non-empty) whether that
// address may self-provision as the configured superadmin while signup is off.
func IsSignupEnabled(ctx context.Context, email string) (bool, error) {
	on, err := GetBool(ctx, EnableSignupKey)
	if err != nil {
		return false, err
	}
	if on {
		return true, nil
	}
	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" {
		return false, nil
	}
	return utils.EmailMatchesSuperadmin(email), nil
}

func SetSignupEnabled(ctx context.Context, enabled bool) error {
	return UpsertBool(ctx, EnableSignupKey, enabled)
}

func IsPaymentEnabled(ctx context.Context) (bool, error) {
	return GetBool(ctx, EnablePaymentsKey)
}

func SetPaymentEnabled(ctx context.Context, enabled bool) error {
	return UpsertBool(ctx, EnablePaymentsKey, enabled)
}

func IsBypassLimitForSuperadminEnabled(ctx context.Context) (bool, error) {
	return GetBool(ctx, AllowBypassLimitForSuperadminKey)
}

func SetBypassLimitForSuperadminEnabled(ctx context.Context, enabled bool) error {
	return UpsertBool(ctx, AllowBypassLimitForSuperadminKey, enabled)
}
