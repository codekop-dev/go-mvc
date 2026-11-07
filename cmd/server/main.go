package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/codekop/gomvc/internal/app"
	"github.com/codekop/gomvc/internal/app/config"
	"github.com/codekop/gomvc/internal/app/database"
)

// main bootstraps configuration, database, Gin routes, and graceful shutdown.
func main() {
	// .env is convenient in development. Existing environment variables always win.
	_ = config.LoadDotEnv(".env")
	cfg := config.Load()

	logger := newLogger(cfg)
	db, err := database.Open(cfg.Database)
	if err != nil {
		logger.Error("database setup failed", "error", err)
		os.Exit(1)
	}
	if db != nil {
		defer db.Close()
	}

	application := app.New(cfg, db, logger)
	server := &http.Server{
		Addr:              cfg.HTTP.Address,
		Handler:           application.Handler(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       cfg.HTTP.ReadTimeout,
		WriteTimeout:      cfg.HTTP.WriteTimeout,
		IdleTimeout:       cfg.HTTP.IdleTimeout,
		MaxHeaderBytes:    1 << 20,
	}

	go func() {
		logger.Info("server started", "address", cfg.HTTP.Address, "environment", cfg.AppEnv)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server stopped unexpectedly", "error", err)
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("graceful shutdown failed", "error", err)
		os.Exit(1)
	}
	logger.Info("server stopped")
}

// newLogger creates the structured logger used by the HTTP middleware and server.
func newLogger(cfg config.Config) *slog.Logger {
	level := slog.LevelInfo
	if cfg.IsDevelopment() {
		level = slog.LevelDebug
	}
	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level}))
}
