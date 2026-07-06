package ai

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
)

type Anomaly struct {
	Metric     string    `json:"metric"`
	Source     string    `json:"source"`
	Value      float64   `json:"value"`
	Mean       float64   `json:"mean"`
	StdDev     float64   `json:"std_dev"`
	ZScore     float64   `json:"z_score"`
	DetectedAt time.Time `json:"detected_at"`
}

type AnomalyDetector struct {
	pool      *pgxpool.Pool
	provider  LLMProvider
	logger    zerolog.Logger
	threshold float64 // Z-score threshold, default 3.0
}

func NewAnomalyDetector(pool *pgxpool.Pool, provider LLMProvider, logger zerolog.Logger) *AnomalyDetector {
	return &AnomalyDetector{
		pool:      pool,
		provider:  provider,
		logger:    logger,
		threshold: 3.0,
	}
}

type AnomalyReport struct {
	Anomalies   []Anomaly `json:"anomalies"`
	Explanation string    `json:"explanation"`
}

// DetectAnomalies runs Z-score anomaly detection on the last window of data
// compared to a 24-hour baseline.
func (d *AnomalyDetector) DetectAnomalies(ctx context.Context) (*AnomalyReport, error) {
	now := time.Now().UTC()
	currentEnd := now
	currentStart := now.Add(-5 * time.Minute)
	baselineStart := now.Add(-24 * time.Hour)
	baselineEnd := currentStart

	var anomalies []Anomaly

	// Detect per-source error rate anomalies
	sources, err := d.getActiveSources(ctx, baselineStart, currentEnd)
	if err != nil {
		return nil, fmt.Errorf("get sources: %w", err)
	}

	for _, source := range sources {
		// Error rate anomaly
		anomaly, err := d.checkMetric(ctx, source, "error_rate", baselineStart, baselineEnd, currentStart, currentEnd)
		if err != nil {
			d.logger.Warn().Err(err).Str("source", source).Msg("Anomaly check failed")
			continue
		}
		if anomaly != nil {
			anomalies = append(anomalies, *anomaly)
		}

		// Event volume anomaly
		anomaly, err = d.checkMetric(ctx, source, "event_count", baselineStart, baselineEnd, currentStart, currentEnd)
		if err != nil {
			d.logger.Warn().Err(err).Str("source", source).Msg("Anomaly check failed")
			continue
		}
		if anomaly != nil {
			anomalies = append(anomalies, *anomaly)
		}
	}

	report := &AnomalyReport{Anomalies: anomalies}

	if len(anomalies) > 0 && d.provider != nil {
		explanation, err := d.explainAnomalies(ctx, anomalies)
		if err != nil {
			d.logger.Warn().Err(err).Msg("Anomaly explanation failed")
		} else {
			report.Explanation = explanation
		}
	}

	return report, nil
}

func (d *AnomalyDetector) getActiveSources(ctx context.Context, start, end time.Time) ([]string, error) {
	rows, err := d.pool.Query(ctx, `
		SELECT DISTINCT source FROM event_counts_1m
		WHERE bucket >= $1 AND bucket < $2
	`, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sources []string
	for rows.Next() {
		var s string
		if err := rows.Scan(&s); err != nil {
			return nil, err
		}
		sources = append(sources, s)
	}
	return sources, nil
}

func (d *AnomalyDetector) checkMetric(ctx context.Context, source, metric string, baseStart, baseEnd, curStart, curEnd time.Time) (*Anomaly, error) {
	// Compute baseline stats (mean + stddev) over 5-minute windows in the baseline period
	var baselineMean, baselineStdDev float64

	switch metric {
	case "error_rate":
		baselineMean, baselineStdDev = d.computeErrorRateStats(ctx, source, baseStart, baseEnd)
		currentValue := d.computeCurrentErrorRate(ctx, source, curStart, curEnd)

		if baselineStdDev == 0 {
			return nil, nil
		}

		zScore := (currentValue - baselineMean) / baselineStdDev
		if zScore > d.threshold {
			return &Anomaly{
				Metric:     metric,
				Source:     source,
				Value:      currentValue,
				Mean:       baselineMean,
				StdDev:     baselineStdDev,
				ZScore:     zScore,
				DetectedAt: time.Now().UTC(),
			}, nil
		}

	case "event_count":
		baselineMean, baselineStdDev = d.computeVolumeStats(ctx, source, baseStart, baseEnd)
		currentValue := d.computeCurrentVolume(ctx, source, curStart, curEnd)

		if baselineStdDev == 0 {
			return nil, nil
		}

		zScore := (currentValue - baselineMean) / baselineStdDev
		if math.Abs(zScore) > d.threshold {
			return &Anomaly{
				Metric:     metric,
				Source:     source,
				Value:      currentValue,
				Mean:       baselineMean,
				StdDev:     baselineStdDev,
				ZScore:     zScore,
				DetectedAt: time.Now().UTC(),
			}, nil
		}
	}

	return nil, nil
}

func (d *AnomalyDetector) computeErrorRateStats(ctx context.Context, source string, start, end time.Time) (mean, stddev float64) {
	// Compute error rate per 5-minute window
	rows, err := d.pool.Query(ctx, `
		WITH windows AS (
			SELECT
				date_trunc('minute', bucket) - (EXTRACT(minute FROM bucket)::int % 5) * INTERVAL '1 minute' AS window,
				SUM(count) AS total,
				SUM(CASE WHEN severity IN ('error','fatal') THEN count ELSE 0 END) AS errors
			FROM event_counts_1m
			WHERE bucket >= $1 AND bucket < $2 AND source = $3
			GROUP BY window
			HAVING SUM(count) > 0
		)
		SELECT AVG(errors::float / total), COALESCE(STDDEV(errors::float / total), 0)
		FROM windows
	`, start, end, source)
	if err != nil {
		return 0, 0
	}
	defer rows.Close()

	if rows.Next() {
		var m, s *float64
		rows.Scan(&m, &s)
		if m != nil {
			mean = *m
		}
		if s != nil {
			stddev = *s
		}
	}
	return
}

func (d *AnomalyDetector) computeCurrentErrorRate(ctx context.Context, source string, start, end time.Time) float64 {
	var total, errors int64
	d.pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(count), 0),
		       COALESCE(SUM(CASE WHEN severity IN ('error','fatal') THEN count ELSE 0 END), 0)
		FROM event_counts_1m
		WHERE bucket >= $1 AND bucket < $2 AND source = $3
	`, start, end, source).Scan(&total, &errors)

	if total == 0 {
		return 0
	}
	return float64(errors) / float64(total)
}

func (d *AnomalyDetector) computeVolumeStats(ctx context.Context, source string, start, end time.Time) (mean, stddev float64) {
	rows, err := d.pool.Query(ctx, `
		WITH windows AS (
			SELECT
				date_trunc('minute', bucket) - (EXTRACT(minute FROM bucket)::int % 5) * INTERVAL '1 minute' AS window,
				SUM(count) AS total
			FROM event_counts_1m
			WHERE bucket >= $1 AND bucket < $2 AND source = $3
			GROUP BY window
		)
		SELECT AVG(total), COALESCE(STDDEV(total), 0)
		FROM windows
	`, start, end, source)
	if err != nil {
		return 0, 0
	}
	defer rows.Close()

	if rows.Next() {
		var m, s *float64
		rows.Scan(&m, &s)
		if m != nil {
			mean = *m
		}
		if s != nil {
			stddev = *s
		}
	}
	return
}

func (d *AnomalyDetector) computeCurrentVolume(ctx context.Context, source string, start, end time.Time) float64 {
	var count int64
	d.pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(count), 0) FROM event_counts_1m
		WHERE bucket >= $1 AND bucket < $2 AND source = $3
	`, start, end, source).Scan(&count)
	return float64(count)
}

func (d *AnomalyDetector) explainAnomalies(ctx context.Context, anomalies []Anomaly) (string, error) {
	prompt := fmt.Sprintf(`You are an SRE analyzing anomalies detected by AtlasDB's Z-score anomaly detection.

Current time: %s

The following anomalies were detected:
`, time.Now().UTC().Format(time.RFC3339))

	for i, a := range anomalies {
		prompt += fmt.Sprintf(`
Anomaly %d:
- Source: %s
- Metric: %s
- Current value: %.4f
- 24h baseline mean: %.4f
- 24h baseline std dev: %.4f
- Z-score: %.2f (threshold: %.1f)
`, i+1, a.Source, a.Metric, a.Value, a.Mean, a.StdDev, a.ZScore, d.threshold)
	}

	prompt += `
Provide a concise explanation of:
1. What happened (in plain English)
2. Potential root causes
3. Recommended actions

Be specific and reference the data points above.`

	resp, err := d.provider.Complete(ctx, CompletionRequest{
		Messages: []Message{
			{Role: RoleUser, Content: prompt},
		},
		MaxTokens:   1024,
		Temperature: 0.3,
	})
	if err != nil {
		return "", err
	}

	return resp.Content, nil
}
