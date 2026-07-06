package analytics

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
)

type Engine struct {
	pool   *pgxpool.Pool
	logger zerolog.Logger
}

func NewEngine(pool *pgxpool.Pool, logger zerolog.Logger) *Engine {
	return &Engine{pool: pool, logger: logger}
}

// --- Time-series ---

type TimeSeriesPoint struct {
	Bucket   time.Time `json:"bucket"`
	Count    int64     `json:"count"`
	Source   string    `json:"source,omitempty"`
	Severity string    `json:"severity,omitempty"`
}

type TimeSeriesRequest struct {
	StartTime  time.Time
	EndTime    time.Time
	Resolution string // "1m" or "1h"
	Source     string
	Severity   string
	GroupBy    string // "source", "severity", or ""
}

func (e *Engine) TimeSeries(ctx context.Context, req TimeSeriesRequest) ([]TimeSeriesPoint, error) {
	table := "event_counts_1m"
	if req.Resolution == "1h" {
		table = "event_counts_1h"
	}

	selectCols := "bucket, SUM(count) AS count"
	groupCols := "bucket"
	if req.GroupBy == "source" {
		selectCols = "bucket, source, SUM(count) AS count"
		groupCols = "bucket, source"
	} else if req.GroupBy == "severity" {
		selectCols = "bucket, severity, SUM(count) AS count"
		groupCols = "bucket, severity"
	}

	var conditions []string
	var args []interface{}
	idx := 1

	conditions = append(conditions, fmt.Sprintf("bucket >= $%d", idx))
	args = append(args, req.StartTime)
	idx++

	conditions = append(conditions, fmt.Sprintf("bucket < $%d", idx))
	args = append(args, req.EndTime)
	idx++

	if req.Source != "" {
		conditions = append(conditions, fmt.Sprintf("source = $%d", idx))
		args = append(args, req.Source)
		idx++
	}
	if req.Severity != "" {
		conditions = append(conditions, fmt.Sprintf("severity = $%d", idx))
		args = append(args, req.Severity)
		idx++
	}

	query := fmt.Sprintf(`
		SELECT %s FROM %s
		WHERE %s
		GROUP BY %s
		ORDER BY bucket ASC
	`, selectCols, table, strings.Join(conditions, " AND "), groupCols)

	rows, err := e.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("timeseries query: %w", err)
	}
	defer rows.Close()

	var points []TimeSeriesPoint
	for rows.Next() {
		var p TimeSeriesPoint
		switch req.GroupBy {
		case "source":
			err = rows.Scan(&p.Bucket, &p.Source, &p.Count)
		case "severity":
			err = rows.Scan(&p.Bucket, &p.Severity, &p.Count)
		default:
			err = rows.Scan(&p.Bucket, &p.Count)
		}
		if err != nil {
			return nil, fmt.Errorf("scan timeseries: %w", err)
		}
		points = append(points, p)
	}
	return points, nil
}

// --- Summary ---

type Summary struct {
	TotalEvents   int64           `json:"total_events"`
	ErrorCount    int64           `json:"error_count"`
	ErrorRate     float64         `json:"error_rate"`
	ActiveSources int             `json:"active_sources"`
	TopSources    []SourceCount   `json:"top_sources"`
	BySeverity    []SeverityCount `json:"by_severity"`
}

type SourceCount struct {
	Source string `json:"source"`
	Count  int64  `json:"count"`
}

type SeverityCount struct {
	Severity string `json:"severity"`
	Count    int64  `json:"count"`
}

func (e *Engine) Summary(ctx context.Context, start, end time.Time) (*Summary, error) {
	s := &Summary{}

	// Total + error count
	err := e.pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(count), 0),
		       COALESCE(SUM(CASE WHEN severity IN ('error','fatal') THEN count ELSE 0 END), 0)
		FROM event_counts_1m
		WHERE bucket >= $1 AND bucket < $2
	`, start, end).Scan(&s.TotalEvents, &s.ErrorCount)
	if err != nil {
		return nil, fmt.Errorf("summary totals: %w", err)
	}

	if s.TotalEvents > 0 {
		s.ErrorRate = float64(s.ErrorCount) / float64(s.TotalEvents)
	}

	// Active sources
	err = e.pool.QueryRow(ctx, `
		SELECT COUNT(DISTINCT source) FROM event_counts_1m
		WHERE bucket >= $1 AND bucket < $2
	`, start, end).Scan(&s.ActiveSources)
	if err != nil {
		return nil, fmt.Errorf("summary sources: %w", err)
	}

	// Top sources
	rows, err := e.pool.Query(ctx, `
		SELECT source, SUM(count) AS total FROM event_counts_1m
		WHERE bucket >= $1 AND bucket < $2
		GROUP BY source ORDER BY total DESC LIMIT 10
	`, start, end)
	if err != nil {
		return nil, fmt.Errorf("top sources: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var sc SourceCount
		if err := rows.Scan(&sc.Source, &sc.Count); err != nil {
			return nil, err
		}
		s.TopSources = append(s.TopSources, sc)
	}

	// By severity
	rows2, err := e.pool.Query(ctx, `
		SELECT severity, SUM(count) AS total FROM event_counts_1m
		WHERE bucket >= $1 AND bucket < $2
		GROUP BY severity ORDER BY total DESC
	`, start, end)
	if err != nil {
		return nil, fmt.Errorf("by severity: %w", err)
	}
	defer rows2.Close()
	for rows2.Next() {
		var sc SeverityCount
		if err := rows2.Scan(&sc.Severity, &sc.Count); err != nil {
			return nil, err
		}
		s.BySeverity = append(s.BySeverity, sc)
	}

	return s, nil
}

// --- Top-N ---

type TopNRequest struct {
	StartTime time.Time
	EndTime   time.Time
	GroupBy   string // "source", "event_type"
	Limit     int
}

type TopNResult struct {
	Key   string `json:"key"`
	Count int64  `json:"count"`
}

func (e *Engine) TopN(ctx context.Context, req TopNRequest) ([]TopNResult, error) {
	col := "source"
	if req.GroupBy == "event_type" {
		col = "event_type"
	}
	limit := req.Limit
	if limit <= 0 || limit > 100 {
		limit = 10
	}

	query := fmt.Sprintf(`
		SELECT %s, SUM(count) AS total FROM event_counts_1m
		WHERE bucket >= $1 AND bucket < $2
		GROUP BY %s ORDER BY total DESC LIMIT %d
	`, col, col, limit)

	rows, err := e.pool.Query(ctx, query, req.StartTime, req.EndTime)
	if err != nil {
		return nil, fmt.Errorf("topn query: %w", err)
	}
	defer rows.Close()

	var results []TopNResult
	for rows.Next() {
		var r TopNResult
		if err := rows.Scan(&r.Key, &r.Count); err != nil {
			return nil, err
		}
		results = append(results, r)
	}
	return results, nil
}
