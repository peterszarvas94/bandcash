package utils

import (
	"fmt"
	"sort"

	"github.com/labstack/echo/v4"
)

func PrintRoutes(e *echo.Echo) {
	routes := e.Routes()
	sort.Slice(routes, func(i, j int) bool {
		if routes[i].Path == routes[j].Path {
			return routes[i].Method < routes[j].Method
		}
		return routes[i].Path < routes[j].Path
	})
	for _, route := range routes {
		fmt.Printf("%s\t%s\n", route.Method, route.Path)
	}
}
