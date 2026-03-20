package utils

import (
	"context"
)

type detailCardStatesContextKey string

const detailCardStatesKey detailCardStatesContextKey = "detail_card_states"

func WithDetailCardStates(ctx context.Context, states map[string]bool) context.Context {
	return context.WithValue(ctx, detailCardStatesKey, states)
}

func DetailCardOpen(ctx context.Context, key string, defaultOpen bool) bool {
	if key == "" {
		return defaultOpen
	}
	states, ok := ctx.Value(detailCardStatesKey).(map[string]bool)
	if !ok {
		return defaultOpen
	}
	open, ok := states[key]
	if !ok {
		return defaultOpen
	}

	return open
}
