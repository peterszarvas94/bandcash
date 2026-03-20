package utils

import (
	"context"
	"sync"
)

type detailStateStore struct {
	mu       sync.RWMutex
	byClient map[string]map[string]bool
}

var detailStates = &detailStateStore{byClient: map[string]map[string]bool{}}

func SetDetailCardOpen(clientID, key string, open bool) {
	if clientID == "" || key == "" {
		return
	}

	detailStates.mu.Lock()
	defer detailStates.mu.Unlock()

	if _, ok := detailStates.byClient[clientID]; !ok {
		detailStates.byClient[clientID] = map[string]bool{}
	}
	detailStates.byClient[clientID][key] = open
}

func DetailCardOpen(ctx context.Context, key string, defaultOpen bool) bool {
	if key == "" {
		return defaultOpen
	}

	clientID := ClientIDFromContext(ctx)
	if clientID == "" {
		return defaultOpen
	}

	detailStates.mu.RLock()
	defer detailStates.mu.RUnlock()

	clientStates, ok := detailStates.byClient[clientID]
	if !ok {
		return defaultOpen
	}

	open, ok := clientStates[key]
	if !ok {
		return defaultOpen
	}

	return open
}
