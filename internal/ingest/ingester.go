package ingest

import (
	"context"
	"crypto/rand"
	"fmt"
	"time"

	"github.com/oklog/ulid/v2"
	"github.com/rs/zerolog"

	"github.com/atlasdb/atlasdb/internal/config"
	"github.com/atlasdb/atlasdb/internal/queue"
	"github.com/atlasdb/atlasdb/internal/telemetry"
	"github.com/atlasdb/atlasdb/pkg/models"
)

type Ingester struct {
	producer *queue.Producer
	cfg      config.IngestConfig
	metrics  *telemetry.Metrics
	logger   zerolog.Logger
}

func NewIngester(
	producer *queue.Producer,
	cfg config.IngestConfig,
	metrics *telemetry.Metrics,
	logger zerolog.Logger,
) *Ingester {
	return &Ingester{
		producer: producer,
		cfg:      cfg,
		metrics:  metrics,
		logger:   logger,
	}
}

func (ing *Ingester) Ingest(ctx context.Context, inputs []models.EventInput) (*models.IngestResponse, error) {
	if len(inputs) == 0 {
		return nil, fmt.Errorf("no events provided")
	}

	if len(inputs) > ing.cfg.MaxBatchSize {
		return nil, fmt.Errorf("batch size %d exceeds maximum %d", len(inputs), ing.cfg.MaxBatchSize)
	}

	// Backpressure check
	depth, err := ing.producer.QueueDepth(ctx)
	if err != nil {
		ing.logger.Warn().Err(err).Msg("Failed to check queue depth")
	} else if depth > ing.cfg.QueueBackpressure {
		return nil, fmt.Errorf("queue backpressure: depth %d exceeds threshold %d", depth, ing.cfg.QueueBackpressure)
	}

	// Validate and convert
	var allErrors []ValidationError
	events := make([]models.Event, 0, len(inputs))

	now := time.Now().UTC()

	for i, input := range inputs {
		if errs := ValidateEventInput(i, input, ing.cfg.MaxEventSizeBytes); len(errs) > 0 {
			allErrors = append(allErrors, errs...)
			continue
		}

		eventID := ulid.MustNew(ulid.Timestamp(now), rand.Reader).String()

		ts := now
		if input.Timestamp != nil {
			ts = *input.Timestamp
		}

		severity := input.Severity
		if severity == "" {
			severity = models.SeverityInfo
		}

		events = append(events, models.Event{
			EventID:    eventID,
			Source:     input.Source,
			EventType:  input.EventType,
			Severity:   severity,
			Timestamp:  ts,
			ReceivedAt: now,
			Data:       input.Data,
			Tags:       input.Tags,
			Metadata:   input.Metadata,
		})
	}

	if len(allErrors) > 0 && len(events) == 0 {
		return nil, fmt.Errorf("all events failed validation: %v", allErrors)
	}

	if err := ing.producer.Publish(ctx, events); err != nil {
		return nil, fmt.Errorf("publish events: %w", err)
	}

	// Record metrics
	ing.metrics.IngestBatchSize.Observe(float64(len(events)))
	for _, e := range events {
		ing.metrics.EventsIngestedTotal.WithLabelValues(e.Source).Inc()
		ing.metrics.EventsIngestedBytes.Add(float64(len(e.Data)))
	}

	eventIDs := make([]string, len(events))
	for i, e := range events {
		eventIDs[i] = e.EventID
	}

	return &models.IngestResponse{
		Accepted: len(events),
		EventIDs: eventIDs,
	}, nil
}
