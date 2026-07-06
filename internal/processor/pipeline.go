package processor

import (
	"context"
	"time"

	"github.com/rs/zerolog"

	"github.com/atlasdb/atlasdb/internal/storage/postgres"
	"github.com/atlasdb/atlasdb/internal/telemetry"
	"github.com/atlasdb/atlasdb/pkg/models"
)

type Pipeline struct {
	eventStore *postgres.EventStore
	metrics    *telemetry.Metrics
	logger     zerolog.Logger
}

func NewPipeline(
	eventStore *postgres.EventStore,
	metrics *telemetry.Metrics,
	logger zerolog.Logger,
) *Pipeline {
	return &Pipeline{
		eventStore: eventStore,
		metrics:    metrics,
		logger:     logger,
	}
}

// Process handles a batch of events from the queue:
// 1. Enrich events (add derived fields)
// 2. Write to PostgreSQL
// 3. Update aggregation counters
func (p *Pipeline) Process(ctx context.Context, events []models.Event) error {
	start := time.Now()

	// Enrich
	for i := range events {
		p.enrich(&events[i])
	}

	// Store events
	if err := p.eventStore.InsertBatch(ctx, events); err != nil {
		p.metrics.ProcessingErrors.WithLabelValues("storage").Inc()
		p.logger.Error().Err(err).Int("count", len(events)).Msg("Failed to store events")
		return err
	}

	// Update aggregation counters (non-fatal if this fails)
	if err := p.eventStore.IncrementCounts(ctx, events); err != nil {
		p.metrics.ProcessingErrors.WithLabelValues("aggregation").Inc()
		p.logger.Warn().Err(err).Msg("Failed to update aggregation counters")
	}

	duration := time.Since(start).Seconds()
	p.metrics.ProcessingDuration.Observe(duration)
	p.metrics.EventsProcessedTotal.WithLabelValues("success").Add(float64(len(events)))

	p.logger.Debug().
		Int("count", len(events)).
		Float64("duration_s", duration).
		Msg("Batch processed")

	return nil
}

func (p *Pipeline) enrich(event *models.Event) {
	if event.Tags == nil {
		event.Tags = []string{}
	}
	if event.Metadata == nil {
		event.Metadata = []byte("{}")
	}
	if event.Data == nil {
		event.Data = []byte("{}")
	}
}
