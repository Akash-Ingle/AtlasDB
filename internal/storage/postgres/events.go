package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"

	"github.com/atlasdb/atlasdb/pkg/models"
)

type EventStore struct {
	pool   *pgxpool.Pool
	logger zerolog.Logger
}

func NewEventStore(pool *pgxpool.Pool, logger zerolog.Logger) *EventStore {
	return &EventStore{pool: pool, logger: logger}
}

func (s *EventStore) InsertBatch(ctx context.Context, events []models.Event) error {
	if len(events) == 0 {
		return nil
	}

	query := `
		INSERT INTO events (event_id, source, event_type, severity, timestamp, received_at, data, tags, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (event_id, timestamp) DO NOTHING
	`

	batch := &pgx.Batch{}
	for _, e := range events {
		batch.Queue(query,
			e.EventID,
			e.Source,
			e.EventType,
			string(e.Severity),
			e.Timestamp,
			e.ReceivedAt,
			e.Data,
			e.Tags,
			e.Metadata,
		)
	}

	br := s.pool.SendBatch(ctx, batch)
	defer br.Close()

	for i := 0; i < len(events); i++ {
		if _, err := br.Exec(); err != nil {
			return fmt.Errorf("insert event %d: %w", i, err)
		}
	}

	return nil
}

func (s *EventStore) GetByID(ctx context.Context, eventID string) (*models.Event, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT event_id, source, event_type, severity, timestamp, received_at, data, tags, metadata
		FROM events
		WHERE event_id = $1
		LIMIT 1
	`, eventID)

	return scanEvent(row)
}

func (s *EventStore) Query(ctx context.Context, q models.EventQuery) (*models.PaginatedResponse[models.Event], error) {
	limit := q.Limit
	if limit <= 0 || limit > models.MaxPageSize {
		limit = models.DefaultPageSize
	}

	var conditions []string
	var args []interface{}
	argIdx := 1

	if q.StartTime != nil {
		conditions = append(conditions, fmt.Sprintf("timestamp >= $%d", argIdx))
		args = append(args, *q.StartTime)
		argIdx++
	} else {
		// Default: last 24 hours for partition pruning
		defaultStart := time.Now().Add(-24 * time.Hour)
		conditions = append(conditions, fmt.Sprintf("timestamp >= $%d", argIdx))
		args = append(args, defaultStart)
		argIdx++
	}

	if q.EndTime != nil {
		conditions = append(conditions, fmt.Sprintf("timestamp < $%d", argIdx))
		args = append(args, *q.EndTime)
		argIdx++
	}

	if q.Source != "" {
		conditions = append(conditions, fmt.Sprintf("source = $%d", argIdx))
		args = append(args, q.Source)
		argIdx++
	}

	if q.EventType != "" {
		conditions = append(conditions, fmt.Sprintf("event_type = $%d", argIdx))
		args = append(args, q.EventType)
		argIdx++
	}

	if q.Severity != "" {
		conditions = append(conditions, fmt.Sprintf("severity = $%d", argIdx))
		args = append(args, string(q.Severity))
		argIdx++
	}

	// Cursor-based pagination
	if q.Cursor != "" {
		cursor, err := models.DecodeCursor(q.Cursor)
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

	// Fetch limit+1 to determine has_more
	query := fmt.Sprintf(`
		SELECT event_id, source, event_type, severity, timestamp, received_at, data, tags, metadata
		FROM events
		%s
		ORDER BY timestamp DESC, event_id DESC
		LIMIT %d
	`, where, limit+1)

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query events: %w", err)
	}
	defer rows.Close()

	var events []models.Event
	for rows.Next() {
		var e models.Event
		var sev string
		var data, metadata []byte
		err := rows.Scan(
			&e.EventID, &e.Source, &e.EventType, &sev,
			&e.Timestamp, &e.ReceivedAt, &data, &e.Tags, &metadata,
		)
		if err != nil {
			return nil, fmt.Errorf("scan event: %w", err)
		}
		e.Severity = models.Severity(sev)
		e.Data = json.RawMessage(data)
		if metadata != nil {
			e.Metadata = json.RawMessage(metadata)
		}
		events = append(events, e)
	}

	hasMore := len(events) > limit
	if hasMore {
		events = events[:limit]
	}

	var cursor string
	if hasMore && len(events) > 0 {
		last := events[len(events)-1]
		cursor = models.EncodeCursor(last.Timestamp, last.EventID)
	}

	return &models.PaginatedResponse[models.Event]{
		Data:    events,
		Cursor:  cursor,
		HasMore: hasMore,
	}, nil
}

func (s *EventStore) EnsurePartition(ctx context.Context, date time.Time) error {
	_, err := s.pool.Exec(ctx,
		"SELECT create_daily_partition('events', $1::date)",
		date.Format("2006-01-02"),
	)
	if err != nil {
		return fmt.Errorf("create partition for %s: %w", date.Format("2006-01-02"), err)
	}
	return nil
}

func (s *EventStore) IncrementCounts(ctx context.Context, events []models.Event) error {
	if len(events) == 0 {
		return nil
	}

	query := `
		INSERT INTO event_counts_1m (bucket, source, event_type, severity, count)
		VALUES (date_trunc('minute', $1::timestamptz), $2, $3, $4, 1)
		ON CONFLICT (bucket, source, event_type, severity)
		DO UPDATE SET count = event_counts_1m.count + 1
	`

	batch := &pgx.Batch{}
	for _, e := range events {
		batch.Queue(query, e.Timestamp, e.Source, e.EventType, string(e.Severity))
	}

	br := s.pool.SendBatch(ctx, batch)
	defer br.Close()

	for i := 0; i < len(events); i++ {
		if _, err := br.Exec(); err != nil {
			s.logger.Warn().Err(err).Msg("Failed to increment event count")
		}
	}

	return nil
}

func scanEvent(row pgx.Row) (*models.Event, error) {
	var e models.Event
	var sev string
	var data, metadata []byte
	err := row.Scan(
		&e.EventID, &e.Source, &e.EventType, &sev,
		&e.Timestamp, &e.ReceivedAt, &data, &e.Tags, &metadata,
	)
	if err != nil {
		return nil, err
	}
	e.Severity = models.Severity(sev)
	e.Data = json.RawMessage(data)
	if metadata != nil {
		e.Metadata = json.RawMessage(metadata)
	}
	return &e, nil
}
