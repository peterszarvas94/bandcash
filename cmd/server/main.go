package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"

	"bandcash/internal/db"
	"bandcash/internal/i18n"
	"bandcash/internal/middleware"
	"bandcash/internal/utils"
)

func main() {
	routesFlag := flag.Bool("routes", false, "Print routes and exit")
	flag.Parse()

	utils.SetupLogger()

	e := echo.New()
	e.HideBanner = true
	e.HidePort = true
	e.HTTPErrorHandler = middleware.ErrorHandler()

	e.Use(middleware.Compression)
	e.Use(middleware.RequestID)
	// e.Use(middleware.GlobalDelay)
	e.Use(middleware.Locale)
	e.Use(middleware.GlobalRateLimit)
	e.Use(middleware.GlobalBodyLimit)
	e.Use(middleware.FetchSiteProtection)
	e.Use(middleware.OriginProtection)
	e.Use(middleware.CSRFToken)
	e.Use(middleware.CSRFProtection)
	e.Use(middleware.RequestLogger)

	// Routes
	registerRoutes(e)

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
	defer func() {
		if err := db.Close(); err != nil {
			slog.Error("failed to close database", "err", err)
		}
	}()

	err = db.Migrate()
	if err != nil {
		slog.Error("failed to run database migrations", "err", err)
		os.Exit(1)
	}

	startErr := make(chan error, 1)

	// Graceful shutdown
	go func() {
		addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
		slog.Info("server starting", "host", cfg.Host, "port", cfg.Port)
		err := e.Start(addr)
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			startErr <- err
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-startErr:
		slog.Error("server failed to start", "err", err)
		os.Exit(1)
	case <-quit:
	}

	slog.Info("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = e.Shutdown(ctx)
	if err != nil {
		slog.Error("server forced to shutdown", "err", err)
	}
	slog.Info("server exited")
}
