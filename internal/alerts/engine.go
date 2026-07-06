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

type AlertRule struct {
	RuleID          uuid.UUID       `json:"rule_id"`
	Name            string          `json:"name"`
	Description     string          `json:"description,omitempty"`
	Condition       AlertCondition  `json:"condition"`
	Severity        string          `json:"severity"`
	Channels        json.RawMessage `json:"channels"`
	Enabled         bool            `json:"enabled"`
	CooldownSeconds int             `json:"cooldown_seconds"`
	CreatedBy       *uuid.UUID      `json:"created_by,omitempty"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
}

type AlertCondition struct {
	Metric    string  `json:"metric"`    // "error_rate", "event_count", "latency_p99"
	Operator  string  `json:"operator"`  // ">", "<", ">=", "<=", "=="
	Threshold float64 `json:"threshold"`
	Window    string  `json:"window"`    // "1m", "5m", "15m", "1h"
	Source    string  `json:"source,omitempty"`
}

type AlertEvent struct {
	AlertEventID uuid.UUID  `json:"alert_event_id"`
	RuleID       uuid.UUID  `json:"rule_id"`
	RuleName     string     `json:"rule_name,omitempty"`
	Status       string     `json:"status"` // "firing", "resolved"
	FiredAt      time.Time  `json:"fired_at"`
	ResolvedAt   *time.Time `json:"resolved_at,omitempty"`
	Value        float64    `json:"value"`
	Context      json.RawMessage `json:"context,omitempty"`
	Notified     bool       `json:"notified"`
}

type Store struct {
	pool   *pgxpool.Pool
	logger zerolog.Logger
}

func NewStore(pool *pgxpool.Pool, logger zerolog.Logger) *Store {
	return &Store{pool: pool, logger: logger}
}

// --- Alert Rules CRUD ---

func (s *Store) CreateRule(ctx context.Context, rule *AlertRule) error {
	condJSON, _ := json.Marshal(rule.Condition)
	_, err := s.pool.Exec(ctx, `
		INSERT INTO alert_rules (rule_id, name, description, condition, severity, channels, enabled, cooldown_seconds, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, rule.RuleID, rule.Name, rule.Description, condJSON,
		rule.Severity, rule.Channels, rule.Enabled, rule.CooldownSeconds, rule.CreatedBy)
	return err
}

func (s *Store) GetRule(ctx context.Context, ruleID uuid.UUID) (*AlertRule, error) {
	var rule AlertRule
	var condJSON []byte
	err := s.pool.QueryRow(ctx, `
		SELECT rule_id, name, description, condition, severity, channels,
		       enabled, cooldown_seconds, created_by, created_at, updated_at
		FROM alert_rules WHERE rule_id = $1
	`, ruleID).Scan(
		&rule.RuleID, &rule.Name, &rule.Description, &condJSON,
		&rule.Severity, &rule.Channels, &rule.Enabled, &rule.CooldownSeconds,
		&rule.CreatedBy, &rule.CreatedAt, &rule.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	json.Unmarshal(condJSON, &rule.Condition)
	return &rule, nil
}

func (s *Store) ListRules(ctx context.Context) ([]AlertRule, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT rule_id, name, description, condition, severity, channels,
		       enabled, cooldown_seconds, created_by, created_at, updated_at
		FROM alert_rules ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rules []AlertRule
	for rows.Next() {
		var r AlertRule
		var condJSON []byte
		err := rows.Scan(
			&r.RuleID, &r.Name, &r.Description, &condJSON,
			&r.Severity, &r.Channels, &r.Enabled, &r.CooldownSeconds,
			&r.CreatedBy, &r.CreatedAt, &r.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		json.Unmarshal(condJSON, &r.Condition)
		rules = append(rules, r)
	}
	return rules, nil
}

func (s *Store) UpdateRule(ctx context.Context, rule *AlertRule) error {
	condJSON, _ := json.Marshal(rule.Condition)
	tag, err := s.pool.Exec(ctx, `
		UPDATE alert_rules
		SET name = $2, description = $3, condition = $4, severity = $5,
		    channels = $6, enabled = $7, cooldown_seconds = $8, updated_at = NOW()
		WHERE rule_id = $1
	`, rule.RuleID, rule.Name, rule.Description, condJSON,
		rule.Severity, rule.Channels, rule.Enabled, rule.CooldownSeconds)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("rule not found")
	}
	return nil
}

func (s *Store) DeleteRule(ctx context.Context, ruleID uuid.UUID) error {
	tag, err := s.pool.Exec(ctx, "DELETE FROM alert_rules WHERE rule_id = $1", ruleID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("rule not found")
	}
	return nil
}

// --- Alert Events ---

func (s *Store) FireAlert(ctx context.Context, event *AlertEvent) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO alert_events (alert_event_id, rule_id, status, fired_at, value, context)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, event.AlertEventID, event.RuleID, event.Status, event.FiredAt, event.Value, event.Context)
	return err
}

func (s *Store) ListAlertEvents(ctx context.Context, limit int) ([]AlertEvent, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := s.pool.Query(ctx, `
		SELECT ae.alert_event_id, ae.rule_id, ar.name, ae.status,
		       ae.fired_at, ae.resolved_at, ae.value, ae.context, ae.notified
		FROM alert_events ae
		JOIN alert_rules ar ON ar.rule_id = ae.rule_id
		ORDER BY ae.fired_at DESC
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []AlertEvent
	for rows.Next() {
		var e AlertEvent
		err := rows.Scan(
			&e.AlertEventID, &e.RuleID, &e.RuleName, &e.Status,
			&e.FiredAt, &e.ResolvedAt, &e.Value, &e.Context, &e.Notified,
		)
		if err != nil {
			return nil, err
		}
		events = append(events, e)
	}
	return events, nil
}

func (s *Store) MarkNotified(ctx context.Context, alertEventID uuid.UUID) error {
	_, err := s.pool.Exec(ctx,
		"UPDATE alert_events SET notified = true WHERE alert_event_id = $1",
		alertEventID)
	return err
}

// LastFiringTime returns the most recent firing time for a rule, for cooldown checks.
func (s *Store) LastFiringTime(ctx context.Context, ruleID uuid.UUID) (*time.Time, error) {
	var t *time.Time
	err := s.pool.QueryRow(ctx,
		"SELECT MAX(fired_at) FROM alert_events WHERE rule_id = $1 AND status = 'firing'",
		ruleID).Scan(&t)
	if err != nil {
		return nil, err
	}
	return t, nil
}
