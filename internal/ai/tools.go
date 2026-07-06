package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/atlasdb/atlasdb/internal/alerts"
	"github.com/atlasdb/atlasdb/internal/analytics"
	"github.com/atlasdb/atlasdb/internal/storage/postgres"
	"github.com/atlasdb/atlasdb/pkg/models"
	"github.com/rs/zerolog"
)

// ToolRegistry provides the set of tools available to the investigation agent.
type ToolRegistry struct {
	eventStore *postgres.EventStore
	analytics  *analytics.Engine
	alertStore *alerts.Store
	logger     zerolog.Logger
}

func NewToolRegistry(es *postgres.EventStore, ae *analytics.Engine, as *alerts.Store, logger zerolog.Logger) *ToolRegistry {
	return &ToolRegistry{
		eventStore: es,
		analytics:  ae,
		alertStore: as,
		logger:     logger,
	}
}

func (r *ToolRegistry) Definitions() []ToolDef {
	return []ToolDef{
		{
			Name:        "query_events",
			Description: "Query the event store with filters. Returns matching events.",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"source":     map[string]string{"type": "string", "description": "Filter by event source/service name"},
					"event_type": map[string]string{"type": "string", "description": "Filter by event type"},
					"severity":   map[string]string{"type": "string", "description": "Filter by severity: debug, info, warn, error, fatal"},
					"start_time": map[string]string{"type": "string", "description": "Start time in RFC3339 format"},
					"end_time":   map[string]string{"type": "string", "description": "End time in RFC3339 format"},
					"limit":      map[string]string{"type": "integer", "description": "Max events to return (default 20)"},
				},
			},
		},
		{
			Name:        "search_events",
			Description: "Full-text search events by keyword. Use for searching event data content.",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"query":      map[string]string{"type": "string", "description": "Search query text"},
					"source":     map[string]string{"type": "string", "description": "Filter by source"},
					"severity":   map[string]string{"type": "string", "description": "Filter by severity"},
					"start_time": map[string]string{"type": "string", "description": "Start time in RFC3339"},
					"end_time":   map[string]string{"type": "string", "description": "End time in RFC3339"},
					"limit":      map[string]string{"type": "integer", "description": "Max results (default 10)"},
				},
				"required": []string{"query"},
			},
		},
		{
			Name:        "get_metrics",
			Description: "Get time-series metrics for a service. Returns event counts over time.",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"source":     map[string]string{"type": "string", "description": "Service/source name"},
					"severity":   map[string]string{"type": "string", "description": "Filter by severity"},
					"start_time": map[string]string{"type": "string", "description": "Start time in RFC3339"},
					"end_time":   map[string]string{"type": "string", "description": "End time in RFC3339"},
					"resolution": map[string]string{"type": "string", "description": "1m or 1h"},
				},
			},
		},
		{
			Name:        "get_summary",
			Description: "Get a summary of event analytics: total events, error rate, top sources, severity breakdown.",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"start_time": map[string]string{"type": "string", "description": "Start time in RFC3339"},
					"end_time":   map[string]string{"type": "string", "description": "End time in RFC3339"},
				},
			},
		},
		{
			Name:        "get_alert_history",
			Description: "Get recent alert events (fired alerts). Shows what alerts have triggered.",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"limit": map[string]string{"type": "integer", "description": "Max alerts to return (default 20)"},
				},
			},
		},
		{
			Name:        "get_alert_rules",
			Description: "List all configured alert rules.",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{},
			},
		},
	}
}

func (r *ToolRegistry) Execute(ctx context.Context, name string, argsJSON string) (string, error) {
	var args map[string]interface{}
	if argsJSON != "" {
		if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
			return "", fmt.Errorf("parse tool args: %w", err)
		}
	}

	r.logger.Debug().Str("tool", name).Str("args", argsJSON).Msg("Executing tool")

	switch name {
	case "query_events":
		return r.queryEvents(ctx, args)
	case "search_events":
		return r.searchEvents(ctx, args)
	case "get_metrics":
		return r.getMetrics(ctx, args)
	case "get_summary":
		return r.getSummary(ctx, args)
	case "get_alert_history":
		return r.getAlertHistory(ctx, args)
	case "get_alert_rules":
		return r.getAlertRules(ctx)
	default:
		return "", fmt.Errorf("unknown tool: %s", name)
	}
}

func (r *ToolRegistry) queryEvents(ctx context.Context, args map[string]interface{}) (string, error) {
	query := models.EventQuery{
		Limit: intArg(args, "limit", 20),
	}
	if v, ok := args["source"].(string); ok {
		query.Source = v
	}
	if v, ok := args["severity"].(string); ok {
		query.Severity = models.Severity(v)
	}
	if v, ok := args["start_time"].(string); ok {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			query.StartTime = &t
		}
	}
	if v, ok := args["end_time"].(string); ok {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			query.EndTime = &t
		}
	}

	page, err := r.eventStore.Query(ctx, query)
	if err != nil {
		return "", err
	}

	result, _ := json.Marshal(map[string]interface{}{
		"count":  len(page.Data),
		"events": page.Data,
	})
	return string(result), nil
}

func (r *ToolRegistry) searchEvents(ctx context.Context, args map[string]interface{}) (string, error) {
	query := models.EventQuery{
		Limit: intArg(args, "limit", 10),
	}
	if v, ok := args["source"].(string); ok {
		query.Source = v
	}
	if v, ok := args["severity"].(string); ok {
		query.Severity = models.Severity(v)
	}
	if v, ok := args["start_time"].(string); ok {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			query.StartTime = &t
		}
	}
	if v, ok := args["end_time"].(string); ok {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			query.EndTime = &t
		}
	}

	searchText, _ := args["query"].(string)

	page, err := r.eventStore.Query(ctx, query)
	if err != nil {
		return "", err
	}

	result, _ := json.Marshal(map[string]interface{}{
		"search_query": searchText,
		"count":        len(page.Data),
		"events":       page.Data,
	})
	return string(result), nil
}

func (r *ToolRegistry) getMetrics(ctx context.Context, args map[string]interface{}) (string, error) {
	end := time.Now().UTC()
	start := end.Add(-1 * time.Hour)

	if v, ok := args["start_time"].(string); ok {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			start = t
		}
	}
	if v, ok := args["end_time"].(string); ok {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			end = t
		}
	}

	req := analytics.TimeSeriesRequest{
		StartTime:  start,
		EndTime:    end,
		Resolution: strArg(args, "resolution", "1m"),
		Source:     strArg(args, "source", ""),
		Severity:   strArg(args, "severity", ""),
	}

	data, err := r.analytics.TimeSeries(ctx, req)
	if err != nil {
		return "", err
	}

	result, _ := json.Marshal(map[string]interface{}{
		"resolution":  req.Resolution,
		"data_points": len(data),
		"data":        data,
	})
	return string(result), nil
}

func (r *ToolRegistry) getSummary(ctx context.Context, args map[string]interface{}) (string, error) {
	end := time.Now().UTC()
	start := end.Add(-1 * time.Hour)

	if v, ok := args["start_time"].(string); ok {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			start = t
		}
	}
	if v, ok := args["end_time"].(string); ok {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			end = t
		}
	}

	summary, err := r.analytics.Summary(ctx, start, end)
	if err != nil {
		return "", err
	}

	result, _ := json.Marshal(summary)
	return string(result), nil
}

func (r *ToolRegistry) getAlertHistory(ctx context.Context, args map[string]interface{}) (string, error) {
	limit := intArg(args, "limit", 20)
	events, err := r.alertStore.ListAlertEvents(ctx, limit)
	if err != nil {
		return "", err
	}
	result, _ := json.Marshal(map[string]interface{}{
		"count":  len(events),
		"alerts": events,
	})
	return string(result), nil
}

func (r *ToolRegistry) getAlertRules(ctx context.Context) (string, error) {
	rules, err := r.alertStore.ListRules(ctx)
	if err != nil {
		return "", err
	}
	result, _ := json.Marshal(map[string]interface{}{
		"count": len(rules),
		"rules": rules,
	})
	return string(result), nil
}

// Argument helpers
func strArg(args map[string]interface{}, key, def string) string {
	if v, ok := args[key].(string); ok {
		return v
	}
	return def
}

func intArg(args map[string]interface{}, key string, def int) int {
	if v, ok := args[key].(float64); ok {
		return int(v)
	}
	return def
}
