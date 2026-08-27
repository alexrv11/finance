package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/finance/seed/internal/api"
	"github.com/finance/seed/internal/config"
	"github.com/finance/seed/internal/ingest"
	"github.com/finance/seed/internal/sources"
	"github.com/finance/seed/internal/sources/stooq"
	"github.com/finance/seed/internal/sources/yahoo"
	"github.com/finance/seed/internal/store"
)

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})))

	cfg := config.Load()

	// --- Storage ------------------------------------------------------------
	db, err := store.Open(cfg.DBPath)
	if err != nil {
		slog.Error("failed to open database", "err", err)
		os.Exit(1)
	}
	defer db.Close()
	slog.Info("database ready", "path", cfg.DBPath)

	// --- Data sources (Yahoo first = higher priority, Stooq as fallback) ----
	srcs := []sources.Source{
		yahoo.New(cfg.FetchTimeout),
		stooq.New(cfg.FetchTimeout),
	}

	// --- Ingest jobs (one per symbol, daily bars, 1 year lookback) ----------
	jobs := make([]ingest.Job, 0, len(cfg.Symbols))
	for _, sym := range cfg.Symbols {
		jobs = append(jobs, ingest.Job{
			Symbol:    sym,
			Intervals: []string{"1d"},
			Lookback:  365 * 24 * time.Hour,
		})
	}

	// --- Scheduler ----------------------------------------------------------
	scheduler := ingest.NewScheduler(srcs, db, jobs, cfg.ScheduleInterval)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := scheduler.Start(ctx); err != nil {
		slog.Error("failed to start scheduler", "err", err)
		os.Exit(1)
	}
	defer scheduler.Stop()

	// --- API server ---------------------------------------------------------
	server := api.NewServer(db, scheduler)

	go func() {
		slog.Info("API server starting", "port", cfg.APIPort)
		if err := server.Run(cfg.APIPort); err != nil {
			slog.Error("API server stopped", "err", err)
		}
	}()

	// --- Graceful shutdown --------------------------------------------------
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutting down...")
	cancel()
}
