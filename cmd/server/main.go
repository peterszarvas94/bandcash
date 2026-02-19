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

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"

	"bandcash/internal/config"
	"bandcash/internal/db"
	"bandcash/internal/i18n"
	"bandcash/internal/middleware"
	"bandcash/internal/utils"
	"bandcash/models/auth"
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

	if err := godotenv.Load(); err != nil {
		slog.Info("env: no .env file loaded", "err", err)
	}
	if err := utils.ValidateEmailEnv(); err != nil {
		slog.Error("env: invalid email configuration", "err", err)
		os.Exit(1)
	}

	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	e.Use(middleware.Compression())
	e.Use(middleware.RequestID())
	e.Use(middleware.Locale())
	e.Use(middleware.FetchSiteProtection())
	e.Use(middleware.OriginProtection())
	e.Use(middleware.CSRFToken())
	e.Use(middleware.CSRFProtection())
	e.Use(echoMiddleware.RequestLoggerWithConfig(echoMiddleware.RequestLoggerConfig{
		LogStatus: true,
		LogURI:    true,
		LogValuesFunc: func(c echo.Context, v echoMiddleware.RequestLoggerValues) error {
			slog.Info("request", "uri", v.URI, "method", c.Request().Method, "status", v.Status)
			return nil
		},
	}))

	// Routes
	e.Static("/static", "static")

	health.Register(e)
	auth.Register(e)
	group.Register(e)
	home.Register(e)
	event.Register(e)
	member.Register(e)
	settings.Register(e)
	sse.Register(e)

	if *routesFlag {
		utils.PrintRoutes(e)
		return
	}

	cfg := config.Load()

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

	// Graceful shutdown
	go func() {
		slog.Info("server starting", "port", cfg.Port)
		err := e.Start(fmt.Sprintf(":%d", cfg.Port))
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
