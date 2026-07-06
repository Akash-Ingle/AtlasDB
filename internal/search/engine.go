package search

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"

	"github.com/atlasdb/atlasdb/pkg/models"
)

type SearchRequest struct {
	Query     string
	Source    string
	Severity  string
	StartTime *time.Time
	EndTime   *time.Time
	Limit     int
	Cursor    string
}

type SearchResult struct {
	Event models.Event `json:"event"`
	Rank  float64      `json:"rank"`
}

type Engine struct {
	pool   *pgxpool.Pool
	logger zerolog.Logger
}

func NewEngine(pool *pgxpool.Pool, logger zerolog.Logger) *Engine {
	return &Engine{pool: pool, logger: logger}
}

func (e *Engine) Search(ctx context.Context, req SearchRequest) (*models.PaginatedResponse[SearchResult], error) {
	limit := req.Limit
	if limit <= 0 || limit > models.MaxPageSize {
		limit = models.DefaultPageSize
	}

	// Parse the query string into structured filters + free text
	parsed := Parse(req.Query)

	var conditions []string
	var args []interface{}
	argIdx := 1

	// Apply parsed field filters (source:x, severity:y, etc.)
	parsedConds, parsedArgs, argIdx := parsed.BuildSQL(argIdx)
	conditions = append(conditions, parsedConds...)
	args = append(args, parsedArgs...)

	// Override from explicit request params (backward compat)
	if req.Source != "" {
		parsed.Filters["source"] = req.Source
		conditions = append(conditions, fmt.Sprintf("source = $%d", argIdx))
		args = append(args, req.Source)
		argIdx++
	}
	if req.Severity != "" {
		conditions = append(conditions, fmt.Sprintf("severity = $%d", argIdx))
		args = append(args, req.Severity)
		argIdx++
	}

	// Track where the tsquery arg is for ranking
	tsqueryArg := 0
	hasFreeText := false

	if parsed.FreeText != "" {
		hasFreeText = true
		tsqueryArg = argIdx
		// Already added by BuildSQL — but BuildSQL uses to_tsquery.
		// We need the arg index for the ranking function.
		// BuildSQL already added the condition, so we just track the index.
		// Actually, BuildSQL adds it. Let's check if it was added.
	}

	// If BuildSQL didn't add a text condition (query was all field filters),
	// and there's no free text, we skip the tsquery entirely.
	// Re-build cleanly:
	conditions = nil
	args = nil
	argIdx = 1

	// Time range (default last 24h for partition pruning)
	if req.StartTime != nil {
		conditions = append(conditions, fmt.Sprintf("timestamp >= $%d", argIdx))
		args = append(args, *req.StartTime)
		argIdx++
	} else {
		defaultStart := time.Now().Add(-24 * time.Hour)
		conditions = append(conditions, fmt.Sprintf("timestamp >= $%d", argIdx))
		args = append(args, defaultStart)
		argIdx++
	}

	if req.EndTime != nil {
		conditions = append(conditions, fmt.Sprintf("timestamp < $%d", argIdx))
		args = append(args, *req.EndTime)
		argIdx++
	}

	// Field filters from parser
	for field, value := range parsed.Filters {
		col, ok := searchableFields[field]
		if !ok {
			continue
		}
		conditions = append(conditions, fmt.Sprintf("%s = $%d", col, argIdx))
		args = append(args, value)
		argIdx++
	}

	// Explicit request params override parser
	if req.Source != "" {
		if _, already := parsed.Filters["source"]; !already {
			conditions = append(conditions, fmt.Sprintf("source = $%d", argIdx))
			args = append(args, req.Source)
			argIdx++
		}
	}
	if req.Severity != "" {
		if _, already := parsed.Filters["severity"]; !already {
			conditions = append(conditions, fmt.Sprintf("severity = $%d", argIdx))
			args = append(args, req.Severity)
			argIdx++
		}
	}

	// Free-text search
	hasFreeText = parsed.FreeText != ""
	if hasFreeText {
		tsqueryArg = argIdx
		tsq := toTSQuery(parsed.FreeText)
		conditions = append(conditions, fmt.Sprintf("search_vector @@ to_tsquery('english', $%d)", argIdx))
		args = append(args, tsq)
		argIdx++
	}

	// Cursor
	if req.Cursor != "" {
		cursor, err := models.DecodeCursor(req.Cursor)
		if err != nil {
			return nil, fmt.Errorf("invalid cursor: %w", err)
		}
		conditions = append(conditions,
			fmt.Sprintf("(timestamp, event_id) < ($%d, $%d)", argIdx, argIdx+1))
		args = append(args, cursor.Timestamp, cursor.EventID)
		argIdx += 2
	}

	where := ""
	if len(conditions) > 0 {
		where = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Build ranking expression
	rankExpr := "0"
	if hasFreeText {
		rankExpr = fmt.Sprintf("ts_rank_cd(search_vector, to_tsquery('english', $%d))", tsqueryArg)
	}

	orderBy := "timestamp DESC, event_id DESC"
	if hasFreeText {
		orderBy = "rank DESC, " + orderBy
	}

	query := fmt.Sprintf(`
		SELECT event_id, source, event_type, severity, timestamp, received_at, data, tags, metadata,
		       %s AS rank
		FROM events
		%s
		ORDER BY %s
		LIMIT %d
	`, rankExpr, where, orderBy, limit+1)

	rows, err := e.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("search query: %w", err)
	}
	defer rows.Close()

	var results []SearchResult
	for rows.Next() {
		var sr SearchResult
		var sev string
		var data, metadata []byte

		err := rows.Scan(
			&sr.Event.EventID, &sr.Event.Source, &sr.Event.EventType, &sev,
			&sr.Event.Timestamp, &sr.Event.ReceivedAt, &data, &sr.Event.Tags, &metadata,
			&sr.Rank,
		)
		if err != nil {
			return nil, fmt.Errorf("scan search result: %w", err)
		}

		sr.Event.Severity = models.Severity(sev)
		sr.Event.Data = json.RawMessage(data)
		if metadata != nil {
			sr.Event.Metadata = json.RawMessage(metadata)
		}
		results = append(results, sr)
	}

	hasMore := len(results) > limit
	if hasMore {
		results = results[:limit]
	}

	var cursor string
	if hasMore && len(results) > 0 {
		last := results[len(results)-1].Event
		cursor = models.EncodeCursor(last.Timestamp, last.EventID)
	}

	return &models.PaginatedResponse[SearchResult]{
		Data:    results,
		Cursor:  cursor,
		HasMore: hasMore,
	}, nil
}
