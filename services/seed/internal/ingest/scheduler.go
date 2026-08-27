package ingest

import (
	"context"
	"log/slog"
	"time"

	"github.com/finance/seed/internal/sources"
	"github.com/finance/seed/internal/store"
	"github.com/robfig/cron/v3"
)

// Job defines what to fetch for a given symbol.
type Job struct {
	Symbol    string
	Intervals []string // e.g. ["1d", "1h"]
	Lookback  time.Duration
}

// Scheduler orchestrates periodic data ingestion from multiple sources.
type Scheduler struct {
	sources  []sources.Source // priority order: first source wins on conflicts
	store    *store.DB
	jobs     []Job
	cron     *cron.Cron
	schedule string
}

func NewScheduler(srcs []sources.Source, db *store.DB, jobs []Job, schedule string) *Scheduler {
	return &Scheduler{
		sources:  srcs,
		store:    db,
		jobs:     jobs,
		cron:     cron.New(),
		schedule: schedule,
	}
}

// Start registers the cron job and runs an immediate fetch on startup.
func (s *Scheduler) Start(ctx context.Context) error {
	go s.runAll(ctx)

	_, err := s.cron.AddFunc(s.schedule, func() {
		s.runAll(ctx)
	})
	if err != nil {
		return err
	}

	s.cron.Start()
	return nil
}

// Stop gracefully stops the scheduler.
func (s *Scheduler) Stop() { s.cron.Stop() }

// RunNow triggers a manual fetch outside of the cron schedule.
func (s *Scheduler) RunNow(ctx context.Context) { go s.runAll(ctx) }

func (s *Scheduler) runAll(ctx context.Context) {
	slog.Info("ingest: run started", "jobs", len(s.jobs))
	for _, job := range s.jobs {
		for _, interval := range job.Intervals {
			s.fetchJob(ctx, job, interval)
		}
	}
	slog.Info("ingest: run complete")
}

func (s *Scheduler) fetchJob(ctx context.Context, job Job, interval string) {
	to := time.Now().UTC()
	from := to.Add(-job.Lookback)

	var allBars []sources.Bar

	for _, src := range s.sources {
		if !supportsInterval(src, interval) {
			continue
		}

		bars, err := src.Fetch(ctx, job.Symbol, interval, from, to)
		if err != nil {
			slog.Warn("ingest: fetch failed",
				"source", src.Name(),
				"symbol", job.Symbol,
				"interval", interval,
				"err", err,
			)
			continue
		}

		slog.Info("ingest: fetched",
			"source", src.Name(),
			"symbol", job.Symbol,
			"interval", interval,
			"bars", len(bars),
		)
		allBars = append(allBars, bars...)
	}

	if len(allBars) == 0 {
		return
	}

	normalized := Normalize(allBars)

	saved, err := s.store.SaveBars(ctx, normalized)
	if err != nil {
		slog.Error("ingest: save failed",
			"symbol", job.Symbol,
			"interval", interval,
			"err", err,
		)
		return
	}

	slog.Info("ingest: saved",
		"symbol", job.Symbol,
		"interval", interval,
		"bars", saved,
	)
}

func supportsInterval(src sources.Source, interval string) bool {
	for _, si := range src.SupportedIntervals() {
		if si == interval {
			return true
		}
	}
	return false
}
