package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/atlasdb/atlasdb/internal/alerts"
	"github.com/atlasdb/atlasdb/internal/analytics"
	"github.com/atlasdb/atlasdb/internal/config"
	"github.com/atlasdb/atlasdb/internal/storage/postgres"
	"github.com/atlasdb/atlasdb/internal/telemetry"
	"github.com/atlasdb/atlasdb/internal/worker"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "fatal: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	logger := telemetry.InitLogger(cfg.Log.Level, cfg.Log.Format, "worker")
	logger.Info().Msg("Starting AtlasDB background worker")

	pgPool, err := postgres.NewPool(ctx, cfg.Postgres, logger)
	if err != nil {
		return fmt.Errorf("postgres: %w", err)
	}
	defer pgPool.Close()

	if err := postgres.RunMigrations(ctx, pgPool, logger); err != nil {
		return fmt.Errorf("migrations: %w", err)
	}

	// Alert system
	alertStore := alerts.NewStore(pgPool, logger)
	notifier := alerts.NewNotifier(logger)
	evaluator := alerts.NewEvaluator(alertStore, pgPool, notifier, logger)

	// Analytics rollup
	rollup := analytics.NewRollup(pgPool, logger)

	// Event store for partition management
	eventStore := postgres.NewEventStore(pgPool, logger)

	// Scheduler
	scheduler := worker.NewScheduler(logger)

	scheduler.Register(worker.ScheduledTask{
		Name:     "alert-evaluation",
		Interval: 1 * time.Minute,
		Fn: func(ctx context.Context) error {
			return evaluator.EvaluateAll(ctx)
		},
	})

	scheduler.Register(worker.ScheduledTask{
		Name:     "analytics-rollup",
		Interval: 5 * time.Minute,
		Fn: func(ctx context.Context) error {
			return rollup.RollupPendingHours(ctx)
		},
	})

	scheduler.Register(worker.ScheduledTask{
		Name:     "partition-management",
		Interval: 1 * time.Hour,
		Fn: func(ctx context.Context) error {
			// Ensure partitions exist for the next 7 days
			for i := 0; i <= 7; i++ {
				day := time.Now().AddDate(0, 0, i)
				if err := eventStore.EnsurePartition(ctx, day); err != nil {
					logger.Warn().Err(err).Time("date", day).Msg("Failed to ensure partition")
				}
			}
			return nil
		},
	})

	scheduler.Run(ctx)
	return nil
}
