package analytics

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
)

type Rollup struct {
	pool   *pgxpool.Pool
	logger zerolog.Logger
}

func NewRollup(pool *pgxpool.Pool, logger zerolog.Logger) *Rollup {
	return &Rollup{pool: pool, logger: logger}
}

// RollupHour aggregates event_counts_1m into event_counts_1h for a given hour.
// Idempotent: uses INSERT ON CONFLICT DO UPDATE.
func (r *Rollup) RollupHour(ctx context.Context, hour time.Time) error {
	hourStart := hour.Truncate(time.Hour)
	hourEnd := hourStart.Add(time.Hour)

	// Ensure target partition exists
	_, err := r.pool.Exec(ctx,
		"SELECT create_daily_partition('event_counts_1h', $1::date)",
		hourStart.Format("2006-01-02"))
	if err != nil {
		r.logger.Warn().Err(err).Msg("Failed to ensure 1h partition")
	}

	tag, err := r.pool.Exec(ctx, `
		INSERT INTO event_counts_1h (bucket, source, event_type, severity, count)
		SELECT
			date_trunc('hour', bucket) AS bucket,
			source,
			event_type,
			severity,
			SUM(count) AS count
		FROM event_counts_1m
		WHERE bucket >= $1 AND bucket < $2
		GROUP BY date_trunc('hour', bucket), source, event_type, severity
		ON CONFLICT (bucket, source, event_type, severity)
		DO UPDATE SET count = EXCLUDED.count
	`, hourStart, hourEnd)
	if err != nil {
		return fmt.Errorf("rollup hour %s: %w", hourStart, err)
	}

	r.logger.Info().
		Time("hour", hourStart).
		Int64("rows", tag.RowsAffected()).
		Msg("Hourly rollup complete")

	return nil
}

// RollupPendingHours rolls up all completed hours since the last rollup.
func (r *Rollup) RollupPendingHours(ctx context.Context) error {
	// Roll up the last 2 complete hours to catch any late-arriving data
	now := time.Now().UTC().Truncate(time.Hour)
	for i := 2; i >= 1; i-- {
		hour := now.Add(-time.Duration(i) * time.Hour)
		if err := r.RollupHour(ctx, hour); err != nil {
			r.logger.Error().Err(err).Time("hour", hour).Msg("Rollup failed")
		}
	}
	return nil
}
