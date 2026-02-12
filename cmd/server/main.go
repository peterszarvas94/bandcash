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
	"github.com/labstack/echo/v4/middleware"

	"bandcash/app/entry"
	"bandcash/app/health"
	"bandcash/app/home"
	"bandcash/app/payee"
	appSSE "bandcash/app/sse"
	"bandcash/internal/config"
	"bandcash/internal/db"
	_ "bandcash/internal/logger"
	appmw "bandcash/internal/middleware"
)

func main() {
	routesFlag := flag.Bool("routes", false, "Print routes and exit")
	flag.Parse()

	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	e.Use(appmw.Compression())
	e.Use(appmw.RequestID())
	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogStatus: true,
		LogURI:    true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			appmw.Logger(c).Info("request", "uri", v.URI, "method", c.Request().Method, "status", v.Status)
			return nil
		},
	}))

	// Routes
	e.Static("/static", "web/static")

	health.Register(e)
	home.Register(e)
	entry.Register(e)
	payee.Register(e)
	appSSE.Register(e)

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
	defer func() {
		if err := db.Close(); err != nil {
			slog.Error("failed to close database", "err", err)
		}
	}()

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
