package todo

import (
	"sync"

	"html/template"
)

type Item struct {
	ID   int
	Text string
	Done bool
}

type TodoData struct {
	Title string
	Items []Item
}

type Todo struct {
	mu    sync.RWMutex
	items []Item
	tmpl  *template.Template
}

// Add creates a new todo and returns the total count.
func (t *Todo) Add(text string) int {
	t.mu.Lock()
	defer t.mu.Unlock()

	item := Item{
		ID:   len(t.items) + 1,
		Text: text,
		Done: false,
	}
	t.items = append(t.items, item)
	return len(t.items)
}

// All returns a copy of all todos.
func (t *Todo) All() []Item {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return append([]Item(nil), t.items...)
}

// Data returns data for rendering.
func (t *Todo) Data() any {
	return TodoData{
		Title: "Todo List",
		Items: t.All(),
	}
}
