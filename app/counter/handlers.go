package counter

import (
	"github.com/labstack/echo/v4"
	"github.com/starfederation/datastar-go/datastar"

	appmw "webapp/internal/middleware"
	"webapp/internal/utils"
)

type updateSignals struct {
	Dir int `json:"dir"`
}

// Index renders the counter page.
func (ctr *Counter) Index(c echo.Context) error {
	utils.EnsureClientID(c)

	log := appmw.Logger(c)
	log.Debug("counter.index", "value", ctr.Value())
	return ctr.tmpl.ExecuteTemplate(c.Response().Writer, "index", ctr.Data())
}

// Update handles increment/decrement based on dir signal.
func (ctr *Counter) Update(c echo.Context) error {
	log := appmw.Logger(c)

	var signals updateSignals
	if err := datastar.ReadSignals(c.Request(), &signals); err != nil {
		log.Warn("counter.update: failed to read signals", "err", err)
		return c.NoContent(400)
	}

	var newValue int
	if signals.Dir > 0 {
		newValue = ctr.Inc()
	} else if signals.Dir < 0 {
		newValue = ctr.Dec()
	} else {
		log.Debug("counter.update: dir is 0, no change")
		return c.NoContent(200)
	}

	log.Debug("counter.update", "dir", signals.Dir, "value", newValue)

	utils.Store.SignalAll()
	log.Debug("counter.update: signaled all clients")

	return c.NoContent(200)
}
