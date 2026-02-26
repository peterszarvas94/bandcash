package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"

	"bandcash/internal/db"
	"bandcash/internal/i18n"
	"bandcash/internal/middleware"
	"bandcash/internal/utils"
	"bandcash/models/auth"
	"bandcash/models/dev"
	"bandcash/models/event"
	"bandcash/models/group"
	"bandcash/models/health"
	"bandcash/models/home"
	"bandcash/models/member"
	"bandcash/models/settings"
	"bandcash/models/sse"
)

func main() {
	routesFlag := flag.Bool("routes", false, "Print routes and exit")
	flag.Parse()

	utils.SetupLogger()
	utils.LoadAppDotEnv()

	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	e.Use(middleware.Compression())
	e.Use(middleware.RequestID())
	e.Use(middleware.Locale())
	e.Use(middleware.GlobalRateLimit())
	e.Use(middleware.GlobalBodyLimit())
	e.Use(middleware.FetchSiteProtection())
	e.Use(middleware.OriginProtection())
	e.Use(middleware.CSRFToken())
	e.Use(middleware.CSRFProtection())
	e.Use(echoMiddleware.RequestLoggerWithConfig(echoMiddleware.RequestLoggerConfig{
		LogStatus: true,
		LogURI:    false,
		LogValuesFunc: func(c echo.Context, v echoMiddleware.RequestLoggerValues) error {
			req := c.Request()
			slog.Info("http.request.completed",
				"path", req.URL.Path,
				"query", req.URL.RawQuery,
				"method", req.Method,
				"status", v.Status,
			)
			return nil
		},
	}))

	// Routes
	e.Static("/static", "static")

	health.RegisterRoutes(e)
	auth.RegisterRoutes(e)
	group.RegisterRoutes(e)
	home.RegisterRoutes(e)
	event.RegisterRoutes(e)
	member.RegisterRoutes(e)
	settings.RegisterRoutes(e)
	sse.RegisterRoutes(e)
	dev.RegisterRoutes(e)

	if *routesFlag {
		utils.PrintRoutes(e)
		return
	}

	cfg := utils.Env()

	err := i18n.Load()
	if err != nil {
		slog.Error("failed to load locales", "err", err)
		os.Exit(1)
	}

	// Initialize database
	err = db.Init(cfg.DBPath)
	if err != nil {
		slog.Error("failed to initialize database", "err", err)
		os.Exit(1)
	}
	defer db.Close()

	err = db.Migrate()
	if err != nil {
		slog.Error("failed to run database migrations", "err", err)
		os.Exit(1)
	}

	// Graceful shutdown
	go func() {
		addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
		slog.Info("server starting", "host", cfg.Host, "port", cfg.Port)
		err := e.Start(addr)
		if err != nil {
			slog.Info("server stopped", "err", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = e.Shutdown(ctx)
	if err != nil {
		slog.Error("server forced to shutdown", "err", err)
	}
	slog.Info("server exited")
}
