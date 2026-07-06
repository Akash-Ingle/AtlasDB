package alerts

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
)

type Evaluator struct {
	store    *Store
	pool     *pgxpool.Pool
	notifier *Notifier
	logger   zerolog.Logger
}

func NewEvaluator(store *Store, pool *pgxpool.Pool, notifier *Notifier, logger zerolog.Logger) *Evaluator {
	return &Evaluator{
		store:    store,
		pool:     pool,
		notifier: notifier,
		logger:   logger,
	}
}

// EvaluateAll checks all enabled alert rules against current data.
func (e *Evaluator) EvaluateAll(ctx context.Context) error {
	rules, err := e.store.ListRules(ctx)
	if err != nil {
		return fmt.Errorf("list rules: %w", err)
	}

	for _, rule := range rules {
		if !rule.Enabled {
			continue
		}

		if err := e.evaluateRule(ctx, rule); err != nil {
			e.logger.Error().Err(err).
				Str("rule", rule.Name).
				Stringer("rule_id", rule.RuleID).
				Msg("Rule evaluation failed")
		}
	}
	return nil
}

func (e *Evaluator) evaluateRule(ctx context.Context, rule AlertRule) error {
	// Check cooldown
	lastFired, err := e.store.LastFiringTime(ctx, rule.RuleID)
	if err != nil {
		return err
	}
	if lastFired != nil {
		cooldown := time.Duration(rule.CooldownSeconds) * time.Second
		if time.Since(*lastFired) < cooldown {
			return nil // Still in cooldown
		}
	}

	// Compute metric value
	value, err := e.computeMetric(ctx, rule.Condition)
	if err != nil {
		return fmt.Errorf("compute metric: %w", err)
	}

	// Compare against threshold
	triggered := compareValue(value, rule.Condition.Operator, rule.Condition.Threshold)

	if !triggered {
		return nil
	}

	e.logger.Warn().
		Str("rule", rule.Name).
		Float64("value", value).
		Float64("threshold", rule.Condition.Threshold).
		Msg("Alert triggered")

	// Create alert event
	contextData, _ := json.Marshal(map[string]interface{}{
		"metric":    rule.Condition.Metric,
		"value":     value,
		"threshold": rule.Condition.Threshold,
		"source":    rule.Condition.Source,
		"window":    rule.Condition.Window,
	})

	alertEvent := &AlertEvent{
		AlertEventID: uuid.New(),
		RuleID:       rule.RuleID,
		Status:       "firing",
		FiredAt:      time.Now().UTC(),
		Value:        value,
		Context:      contextData,
	}

	if err := e.store.FireAlert(ctx, alertEvent); err != nil {
		return fmt.Errorf("fire alert: %w", err)
	}

	// Notify
	if e.notifier != nil {
		go func() {
			if err := e.notifier.Notify(context.Background(), rule, *alertEvent); err != nil {
				e.logger.Error().Err(err).Str("rule", rule.Name).Msg("Notification failed")
			} else {
				e.store.MarkNotified(context.Background(), alertEvent.AlertEventID)
			}
		}()
	}

	return nil
}

func (e *Evaluator) computeMetric(ctx context.Context, cond AlertCondition) (float64, error) {
	window := parseWindow(cond.Window)
	end := time.Now().UTC()
	start := end.Add(-window)

	switch cond.Metric {
	case "error_rate":
		return e.errorRate(ctx, start, end, cond.Source)
	case "event_count":
		return e.eventCount(ctx, start, end, cond.Source)
	case "error_count":
		return e.errorCount(ctx, start, end, cond.Source)
	default:
		return 0, fmt.Errorf("unknown metric: %s", cond.Metric)
	}
}

func (e *Evaluator) errorRate(ctx context.Context, start, end time.Time, source string) (float64, error) {
	var total, errors int64
	query := `
		SELECT COALESCE(SUM(count), 0),
		       COALESCE(SUM(CASE WHEN severity IN ('error','fatal') THEN count ELSE 0 END), 0)
		FROM event_counts_1m
		WHERE bucket >= $1 AND bucket < $2
	`
	args := []interface{}{start, end}
	if source != "" {
		query += " AND source = $3"
		args = append(args, source)
	}

	err := e.pool.QueryRow(ctx, query, args...).Scan(&total, &errors)
	if err != nil {
		return 0, err
	}
	if total == 0 {
		return 0, nil
	}
	return float64(errors) / float64(total), nil
}

func (e *Evaluator) eventCount(ctx context.Context, start, end time.Time, source string) (float64, error) {
	query := `SELECT COALESCE(SUM(count), 0) FROM event_counts_1m WHERE bucket >= $1 AND bucket < $2`
	args := []interface{}{start, end}
	if source != "" {
		query += " AND source = $3"
		args = append(args, source)
	}
	var count int64
	err := e.pool.QueryRow(ctx, query, args...).Scan(&count)
	return float64(count), err
}

func (e *Evaluator) errorCount(ctx context.Context, start, end time.Time, source string) (float64, error) {
	query := `SELECT COALESCE(SUM(count), 0) FROM event_counts_1m WHERE bucket >= $1 AND bucket < $2 AND severity IN ('error','fatal')`
	args := []interface{}{start, end}
	if source != "" {
		query += " AND source = $3"
		args = append(args, source)
	}
	var count int64
	err := e.pool.QueryRow(ctx, query, args...).Scan(&count)
	return float64(count), err
}

func compareValue(value float64, operator string, threshold float64) bool {
	switch operator {
	case ">":
		return value > threshold
	case ">=":
		return value >= threshold
	case "<":
		return value < threshold
	case "<=":
		return value <= threshold
	case "==":
		return value == threshold
	default:
		return false
	}
}

func parseWindow(w string) time.Duration {
	switch w {
	case "1m":
		return time.Minute
	case "5m":
		return 5 * time.Minute
	case "15m":
		return 15 * time.Minute
	case "1h":
		return time.Hour
	default:
		return 5 * time.Minute
	}
}
