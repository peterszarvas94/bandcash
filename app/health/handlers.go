package health

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// Check returns the health status.
func (h *Health) Check(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
}
