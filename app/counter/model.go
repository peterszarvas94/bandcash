package counter

import (
	"sync"

	"html/template"
)

type CounterData struct {
	Title string
	Value int
}

type Counter struct {
	mu    sync.RWMutex
	count int
	tmpl  *template.Template
}

// Inc increments the counter and returns the new value.
func (ctr *Counter) Inc() int {
	ctr.mu.Lock()
	defer ctr.mu.Unlock()
	ctr.count++
	return ctr.count
}

// Dec decrements the counter and returns the new value.
func (ctr *Counter) Dec() int {
	ctr.mu.Lock()
	defer ctr.mu.Unlock()
	ctr.count--
	return ctr.count
}

// Value returns the current counter value.
func (ctr *Counter) Value() int {
	ctr.mu.RLock()
	defer ctr.mu.RUnlock()
	return ctr.count
}

// Data returns data for rendering.
func (ctr *Counter) Data() any {
	return CounterData{
		Title: "Counter",
		Value: ctr.Value(),
	}
}
