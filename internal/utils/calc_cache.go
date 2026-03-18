package utils

import "sync"

// CalcCache is a simple thread-safe cache for calculation results
// Key: hash string, Value: any calculated result
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

// Global cache instance for shared use
var CalcCacheInstance = NewCalcCache()
