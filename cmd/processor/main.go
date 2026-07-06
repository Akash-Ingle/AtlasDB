package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/atlasdb/atlasdb/internal/config"
	"github.com/atlasdb/atlasdb/internal/processor"
	"github.com/atlasdb/atlasdb/internal/queue"
	"github.com/atlasdb/atlasdb/internal/storage/postgres"
	redisstore "github.com/atlasdb/atlasdb/internal/storage/redis"
	"github.com/atlasdb/atlasdb/internal/telemetry"
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

	logger := telemetry.InitLogger(cfg.Log.Level, cfg.Log.Format, "processor")
	logger.Info().Msg("Starting AtlasDB stream processor")

	metrics := telemetry.NewMetrics("atlas")

	pgPool, err := postgres.NewPool(ctx, cfg.Postgres, logger)
	if err != nil {
		return fmt.Errorf("postgres: %w", err)
	}
	defer pgPool.Close()

	// Run migrations (idempotent)
	if err := postgres.RunMigrations(ctx, pgPool, logger); err != nil {
		return fmt.Errorf("migrations: %w", err)
	}

	redisClient, err := redisstore.NewClient(ctx, cfg.Redis, logger)
	if err != nil {
		return fmt.Errorf("redis: %w", err)
	}
	defer redisClient.Close()

	eventStore := postgres.NewEventStore(pgPool, logger)
	pipeline := processor.NewPipeline(eventStore, metrics, logger)

	hostname, _ := os.Hostname()

	var wg sync.WaitGroup

	for i := 0; i < cfg.Processor.Workers; i++ {
		workerName := fmt.Sprintf("%s-worker-%d", hostname, i)

		consumer := queue.NewConsumer(
			redisClient,
			cfg.Processor.ConsumerGroup,
			workerName,
			cfg.Ingest.NumPartitions,
			cfg.Processor.BatchSize,
			cfg.Processor.ClaimTimeout,
			cfg.Processor.MaxRetries,
			logger.With().Str("worker", workerName).Logger(),
		)

		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := consumer.Run(ctx, pipeline.Process); err != nil && ctx.Err() == nil {
				logger.Error().Err(err).Str("worker", workerName).Msg("Consumer exited with error")
			}
		}()
	}

	logger.Info().Int("workers", cfg.Processor.Workers).Msg("All consumer workers started")

	<-ctx.Done()
	logger.Info().Msg("Shutdown signal received, waiting for workers...")
	cancel()
	wg.Wait()
	logger.Info().Msg("All workers stopped")

	return nil
}
