package utils

import (
	"fmt"
	"net/url"
	"strings"
	"sync"
)

func normalizeKeyPart(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "all"
	}
	return url.QueryEscape(trimmed)
}

func EventsCachePrefix(groupID string) string {
	return fmt.Sprintf("events_group_%s_", normalizeKeyPart(groupID))
}

func ExpensesCachePrefix(groupID string) string {
	return fmt.Sprintf("expenses_group_%s_", normalizeKeyPart(groupID))
}

func GroupTotalsCachePrefix(groupID string) string {
	return fmt.Sprintf("group_totals_group_%s", normalizeKeyPart(groupID))
}

func GroupTotalsCacheKey(groupID string) string {
	return GroupTotalsCachePrefix(groupID)
}

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
	PayoutPaid         int64
	PayoutUnpaid       int64
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
