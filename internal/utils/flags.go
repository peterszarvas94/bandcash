package utils

import (
	"context"
	"database/sql"
	"errors"

	"bandcash/internal/db"
)

const EnableSignupFlagKey = "enable_signup"

func IsSignupEnabled(ctx context.Context) (bool, error) {
	value, err := db.Qry.GetAppFlagBool(ctx, EnableSignupFlagKey)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return value != 0, nil
}

func SetSignupEnabled(ctx context.Context, enabled bool) error {
	var boolValue int64
	if enabled {
		boolValue = 1
	}
	return db.Qry.UpsertAppFlagBool(ctx, db.UpsertAppFlagBoolParams{
		Key:       EnableSignupFlagKey,
		BoolValue: boolValue,
	})
}
