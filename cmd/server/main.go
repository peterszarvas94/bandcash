package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"bandcash/app/entry"
	"bandcash/app/health"
	"bandcash/app/home"
	"bandcash/app/todo"
	"bandcash/internal/config"
	"bandcash/internal/db"
	"bandcash/internal/logger"
	appmw "bandcash/internal/middleware"
	"bandcash/internal/store"
	"bandcash/internal/utils"
)

func main() {
	cfg := config.Load()

	// Setup log file
	if err := os.MkdirAll(filepath.Dir(cfg.LogFile), 0755); err != nil {
		slog.Error("failed to create logs directory", "err", err)
		os.Exit(1)
	}
	logFile, err := os.OpenFile(cfg.LogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		slog.Error("failed to open log file", "err", err)
		os.Exit(1)
	}
	defer func() {
		if err := logFile.Close(); err != nil {
			slog.Error("failed to close log file", "err", err)
		}
	}()

	// Setup logger
	log := logger.New(logFile, cfg.LogLevel)
	slog.SetDefault(log)

	// Initialize store
	utils.Store = store.New()

	// Initialize database
	if err := db.Init(cfg.DBPath); err != nil {
		slog.Error("failed to initialize database", "err", err)
		os.Exit(1)
	}
	defer db.Close()

	// Run migrations
	if err := db.Migrate(); err != nil {
		slog.Error("failed to migrate database", "err", err)
		os.Exit(1)
	}

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
	todo.Register(e)
	entry.Register(e)

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

	// Close all SSE client connections first
	utils.Store.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		slog.Error("server forced to shutdown", "err", err)
	}
	slog.Info("server exited")
}
