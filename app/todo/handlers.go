package todo

import (
	"github.com/labstack/echo/v4"
	"github.com/starfederation/datastar-go/datastar"

	appmw "webapp/internal/middleware"
	"webapp/internal/utils"
)

type createSignals struct {
	Todo string `json:"todo"`
}

// Index renders the todo list page.
func (t *Todo) Index(c echo.Context) error {
	utils.EnsureClientID(c)

	log := appmw.Logger(c)
	data := t.All()

	log.Debug("todo.index", "client_id", "", "todo_count", len(data))
	return t.tmpl.ExecuteTemplate(c.Response().Writer, "index", t.Data())
}

// Create handles adding a new todo item.
func (t *Todo) Create(c echo.Context) error {
	log := appmw.Logger(c)

	var signals createSignals
	if err := datastar.ReadSignals(c.Request(), &signals); err != nil {
		log.Warn("todo.create: failed to read signals", "err", err)
		return c.NoContent(400)
	}

	if signals.Todo == "" {
		log.Debug("todo.create: empty text")
		return c.NoContent(200)
	}

	todoCount := t.Add(signals.Todo)
	log.Debug("todo.create", "text", signals.Todo, "todo_count", todoCount)

	utils.Store.SignalAll()
	log.Debug("todo.create: signaled all clients")

	return c.NoContent(200)
}
