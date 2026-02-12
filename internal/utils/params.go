package utils

import (
	"strconv"

	"github.com/labstack/echo/v4"
)

func ParamInt(c echo.Context, name string) (int, error) {
	return strconv.Atoi(c.Param(name))
}
