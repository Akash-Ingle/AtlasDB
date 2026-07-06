package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

type QueryEngine struct {
	provider LLMProvider
	logger   zerolog.Logger
}

func NewQueryEngine(provider LLMProvider, logger zerolog.Logger) *QueryEngine {
	return &QueryEngine{provider: provider, logger: logger}
}

type NLQueryRequest struct {
	Question string `json:"question"`
	Stream   bool   `json:"stream"`
}

type NLQueryResponse struct {
	Answer     string          `json:"answer"`
	Query      *GeneratedQuery `json:"query,omitempty"`
	Confidence float64         `json:"confidence"`
	Usage      Usage           `json:"usage"`
}

type GeneratedQuery struct {
	Type       string            `json:"type"`
	Filters    map[string]string `json:"filters,omitempty"`
	SearchText string            `json:"search_text,omitempty"`
	StartTime  string            `json:"start_time,omitempty"`
	EndTime    string            `json:"end_time,omitempty"`
	OrderBy    string            `json:"order_by,omitempty"`
	Limit      int               `json:"limit,omitempty"`
	Metric     string            `json:"metric,omitempty"`
	GroupBy    string            `json:"group_by,omitempty"`
}

const schemaContext = `You are AtlasDB AI, an assistant for an event streaming and analytics platform.

Available data:
- events table: event_id (ULID), source (text), event_type (text), severity (debug/info/warn/error/fatal), timestamp (timestamptz), data (jsonb), tags (text[])
- Aggregation: event_counts_1m with columns: bucket, source, event_type, severity, count
- Analytics APIs: timeseries, summary, top-N queries
- Alert rules and alert event history

Current time: %s

When the user asks a question, generate a structured JSON query to answer it.
Respond with ONLY a JSON object with these fields:
{
  "type": "event_query" | "search" | "analytics" | "count",
  "filters": {"field": "value"},
  "search_text": "optional full-text search terms",
  "start_time": "ISO8601",
  "end_time": "ISO8601",
  "order_by": "field DESC",
  "limit": 100,
  "metric": "error_rate | event_count | error_count",
  "group_by": "source | event_type | severity",
  "explanation": "Brief explanation of what this query does"
}`

const summarizePrompt = `Based on the user's question and the query results below, provide a clear, concise answer.
Include specific numbers, event IDs, or timestamps where relevant.
If the results are empty, say so helpfully.
Rate your confidence from 0.0 to 1.0 based on how well the data answers the question.

User question: %s

Query executed: %s

Results (showing first %d):
%s

Respond with JSON:
{
  "answer": "Your natural language answer here",
  "confidence": 0.85
}`

func (e *QueryEngine) TranslateQuery(ctx context.Context, question string) (*GeneratedQuery, error) {
	systemPrompt := fmt.Sprintf(schemaContext, time.Now().UTC().Format(time.RFC3339))

	resp, err := e.provider.Complete(ctx, CompletionRequest{
		Messages: []Message{
			{Role: RoleSystem, Content: systemPrompt},
			{Role: RoleUser, Content: question},
		},
		MaxTokens:   1024,
		Temperature: 0.1,
		JSONMode:    true,
	})
	if err != nil {
		return nil, fmt.Errorf("llm query translation: %w", err)
	}

	// Parse the generated query
	content := strings.TrimSpace(resp.Content)
	var query GeneratedQuery
	if err := json.Unmarshal([]byte(content), &query); err != nil {
		return nil, fmt.Errorf("parse generated query: %w (raw: %s)", err, content)
	}

	// Validate
	if query.Type == "" {
		query.Type = "event_query"
	}
	if query.Limit <= 0 || query.Limit > 1000 {
		query.Limit = 100
	}

	return &query, nil
}

func (e *QueryEngine) Summarize(ctx context.Context, question string, query *GeneratedQuery, results interface{}) (*NLQueryResponse, error) {
	queryJSON, _ := json.Marshal(query)
	resultsJSON, _ := json.MarshalIndent(results, "", "  ")

	// Truncate results if too large
	resultsStr := string(resultsJSON)
	if len(resultsStr) > 8000 {
		resultsStr = resultsStr[:8000] + "\n... (truncated)"
	}

	prompt := fmt.Sprintf(summarizePrompt, question, string(queryJSON), 20, resultsStr)

	resp, err := e.provider.Complete(ctx, CompletionRequest{
		Messages: []Message{
			{Role: RoleUser, Content: prompt},
		},
		MaxTokens:   2048,
		Temperature: 0.3,
		JSONMode:    true,
	})
	if err != nil {
		return nil, fmt.Errorf("llm summarize: %w", err)
	}

	var summary struct {
		Answer     string  `json:"answer"`
		Confidence float64 `json:"confidence"`
	}
	if err := json.Unmarshal([]byte(resp.Content), &summary); err != nil {
		// Fallback: use raw content
		summary.Answer = resp.Content
		summary.Confidence = 0.5
	}

	return &NLQueryResponse{
		Answer:     summary.Answer,
		Query:      query,
		Confidence: summary.Confidence,
		Usage:      resp.Usage,
	}, nil
}

// TranslateAndStream translates the question and streams the response via SSE.
func (e *QueryEngine) TranslateStream(ctx context.Context, question string) (<-chan StreamChunk, *GeneratedQuery, error) {
	// First, translate the question (non-streaming — this is fast)
	query, err := e.TranslateQuery(ctx, question)
	if err != nil {
		return nil, nil, err
	}

	return nil, query, nil
}
