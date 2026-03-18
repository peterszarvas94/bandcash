package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
)

// CalcCache is a simple thread-safe KV cache for calculation results
// Uses hash-based keys for efficient cache lookups
type CalcCache struct {
	mu   sync.RWMutex
	data map[string]any
}

// NewCalcCache creates a new calculation cache
func NewCalcCache() *CalcCache {
	return &CalcCache{
		data: make(map[string]any),
	}
}

// makeKey creates a deterministic hash key from components
func makeKey(prefix string, components ...string) string {
	h := sha256.New()
	h.Write([]byte(prefix))
	for _, c := range components {
		h.Write([]byte("|"))
		h.Write([]byte(c))
	}
	return prefix + ":" + hex.EncodeToString(h.Sum(nil))[:16]
}

// GroupTotalsKey creates a cache key for group financial totals
func GroupTotalsKey(groupID string) string {
	return makeKey("group_totals", groupID)
}

// EventsFilterKey creates a cache key for filtered event calculations
func EventsFilterKey(groupID, search, year, from, to string) string {
	return makeKey("events", groupID, search, year, from, to)
}

// ExpensesFilterKey creates a cache key for filtered expense calculations
func ExpensesFilterKey(groupID, search, year, from, to string) string {
	return makeKey("expenses", groupID, search, year, from, to)
}

// Get retrieves a value from cache
func (c *CalcCache) Get(key string) (any, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	val, ok := c.data[key]
	return val, ok
}

// Set stores a value in cache
func (c *CalcCache) Set(key string, value any) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data[key] = value
}

// Clear removes all entries
func (c *CalcCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data = make(map[string]any)
}

// ClearPrefix removes all entries with keys starting with the given prefix
func (c *CalcCache) ClearPrefix(prefix string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for key := range c.data {
		if len(key) >= len(prefix) && key[:len(prefix)] == prefix {
			delete(c.data, key)
		}
	}
}

// Stats returns cache statistics for debugging
func (c *CalcCache) Stats() (total int, byPrefix map[string]int) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	total = len(c.data)
	byPrefix = make(map[string]int)
	for key := range c.data {
		// Extract prefix (everything before first colon)
		for i, ch := range key {
			if ch == ':' {
				byPrefix[key[:i]]++
				break
			}
		}
	}
	return total, byPrefix
}

// Global cache instance for shared use
var CalcCacheInstance = NewCalcCache()

// GroupTotals holds calculated financial totals for a group
type GroupTotals struct {
	TotalEventAmount   int64
	TotalExpenseAmount int64
	TotalPayoutAmount  int64
	TotalLeftover      int64
	EventPaid          int64
	EventUnpaid        int64
	ExpensePaid        int64
	ExpenseUnpaid      int64
}

func (gt GroupTotals) String() string {
	return fmt.Sprintf("events=%d/%d expenses=%d/%d payouts=%d leftover=%d",
		gt.EventPaid, gt.EventUnpaid,
		gt.ExpensePaid, gt.ExpenseUnpaid,
		gt.TotalPayoutAmount, gt.TotalLeftover)
}
