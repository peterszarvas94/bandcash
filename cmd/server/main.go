package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"sort"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"

	"bandcash/internal/config"
	"bandcash/internal/db"
	"bandcash/internal/middleware"
	"bandcash/internal/utils"
	"bandcash/models/entry"
	"bandcash/models/health"
	"bandcash/models/home"
	"bandcash/models/payee"
	"bandcash/models/sse"
)

func main() {
	utils.InitLogger()

	routesFlag := flag.Bool("routes", false, "Print routes and exit")
	flag.Parse()

	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	e.Use(middleware.Compression())
	e.Use(middleware.RequestID())
	e.Use(echoMiddleware.RequestLoggerWithConfig(echoMiddleware.RequestLoggerConfig{
		LogStatus: true,
		LogURI:    true,
		LogValuesFunc: func(c echo.Context, v echoMiddleware.RequestLoggerValues) error {
			slog.Info("request", "uri", v.URI, "method", c.Request().Method, "status", v.Status)
			return nil
		},
	}))

	// Routes
	e.Static("/static", "web/static")

	health.Register(e)
	home.Register(e)
	entry.Register(e)
	payee.Register(e)
	sse.Register(e)

	if *routesFlag {
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
		return
	}

	cfg := config.Load()

	// Initialize database
	if err := db.Init(cfg.DBPath); err != nil {
		slog.Error("failed to initialize database", "err", err)
		os.Exit(1)
	}
	defer db.Close()

	// Graceful shutdown
	go func() {
		slog.Info("server starting", "port", cfg.Port)
		if err := e.Start(fmt.Sprintf(":%d", cfg.Port)); err != nil {
			slog.Info("server stopped", "err", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		slog.Error("server forced to shutdown", "err", err)
	}
	slog.Info("server exited")
}
