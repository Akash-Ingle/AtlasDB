package alerts

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/rs/zerolog"
)

type WebhookChannel struct {
	Type string `json:"type"`
	URL  string `json:"url"`
}

type WebhookPayload struct {
	Alert     string          `json:"alert"`
	Severity  string          `json:"severity"`
	Status    string          `json:"status"`
	Value     float64         `json:"value"`
	FiredAt   time.Time       `json:"fired_at"`
	RuleID    string          `json:"rule_id"`
	Context   json.RawMessage `json:"context,omitempty"`
}

type Notifier struct {
	httpClient *http.Client
	logger     zerolog.Logger
}

func NewNotifier(logger zerolog.Logger) *Notifier {
	return &Notifier{
		httpClient: &http.Client{Timeout: 10 * time.Second},
		logger:     logger,
	}
}

func (n *Notifier) Notify(ctx context.Context, rule AlertRule, event AlertEvent) error {
	var channels []WebhookChannel
	if err := json.Unmarshal(rule.Channels, &channels); err != nil {
		return fmt.Errorf("parse channels: %w", err)
	}

	payload := WebhookPayload{
		Alert:    rule.Name,
		Severity: rule.Severity,
		Status:   event.Status,
		Value:    event.Value,
		FiredAt:  event.FiredAt,
		RuleID:   rule.RuleID.String(),
		Context:  event.Context,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}

	var lastErr error
	for _, ch := range channels {
		if ch.Type != "webhook" {
			continue
		}

		for attempt := 0; attempt < 3; attempt++ {
			req, err := http.NewRequestWithContext(ctx, "POST", ch.URL, bytes.NewReader(body))
			if err != nil {
				lastErr = err
				break
			}
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("User-Agent", "AtlasDB-Alerting/1.0")

			resp, err := n.httpClient.Do(req)
			if err != nil {
				lastErr = err
				n.logger.Warn().Err(err).
					Str("url", ch.URL).
					Int("attempt", attempt+1).
					Msg("Webhook delivery failed, retrying")
				time.Sleep(time.Duration(attempt+1) * time.Second)
				continue
			}
			resp.Body.Close()

			if resp.StatusCode >= 200 && resp.StatusCode < 300 {
				n.logger.Info().
					Str("alert", rule.Name).
					Str("url", ch.URL).
					Int("status", resp.StatusCode).
					Msg("Webhook delivered")
				lastErr = nil
				break
			}

			lastErr = fmt.Errorf("webhook returned status %d", resp.StatusCode)
			time.Sleep(time.Duration(attempt+1) * time.Second)
		}
	}

	return lastErr
}
