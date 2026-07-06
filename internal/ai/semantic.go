package ai

import (
	"context"
	"fmt"
	"math"
	"sort"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"

	"github.com/atlasdb/atlasdb/pkg/models"
)

type SemanticSearch struct {
	pool     *pgxpool.Pool
	provider LLMProvider
	logger   zerolog.Logger
}

func NewSemanticSearch(pool *pgxpool.Pool, provider LLMProvider, logger zerolog.Logger) *SemanticSearch {
	return &SemanticSearch{pool: pool, provider: provider, logger: logger}
}

type SemanticResult struct {
	Event      models.Event `json:"event"`
	Score      float64      `json:"score"`
	SearchType string       `json:"search_type"` // "keyword", "semantic", "hybrid"
}

// VectorSearch finds events similar to the query using cosine distance on embeddings.
func (s *SemanticSearch) VectorSearch(ctx context.Context, query string, limit int) ([]SemanticResult, error) {
	// Generate embedding for the query
	embeddings, err := s.provider.Embed(ctx, []string{query})
	if err != nil {
		return nil, fmt.Errorf("embed query: %w", err)
	}
	if len(embeddings) == 0 || len(embeddings[0]) == 0 {
		return nil, fmt.Errorf("empty embedding returned")
	}

	queryEmbed := embeddings[0]

	// Build pgvector query
	rows, err := s.pool.Query(ctx, `
		SELECT event_id, source, event_type, severity, timestamp, received_at,
		       data, tags, metadata,
		       1 - (embedding <=> $1::vector) AS score
		FROM events
		WHERE embedding IS NOT NULL
		ORDER BY embedding <=> $1::vector
		LIMIT $2
	`, pgvectorString(queryEmbed), limit)
	if err != nil {
		return nil, fmt.Errorf("vector search: %w", err)
	}
	defer rows.Close()

	var results []SemanticResult
	for rows.Next() {
		var r SemanticResult
		err := rows.Scan(
			&r.Event.EventID, &r.Event.Source, &r.Event.EventType,
			&r.Event.Severity, &r.Event.Timestamp, &r.Event.ReceivedAt,
			&r.Event.Data, &r.Event.Tags, &r.Event.Metadata,
			&r.Score,
		)
		if err != nil {
			return nil, err
		}
		r.SearchType = "semantic"
		results = append(results, r)
	}

	return results, nil
}

// HybridSearch combines keyword (full-text) and semantic (vector) search using
// Reciprocal Rank Fusion (RRF) to merge the two ranked lists.
func (s *SemanticSearch) HybridSearch(ctx context.Context, query string, limit int) ([]SemanticResult, error) {
	k := 60 // RRF constant

	// Run keyword search (ts_rank)
	keywordResults, err := s.keywordSearch(ctx, query, limit*2)
	if err != nil {
		s.logger.Warn().Err(err).Msg("Keyword search failed, falling back to vector only")
		return s.VectorSearch(ctx, query, limit)
	}

	// Run vector search
	vectorResults, err := s.VectorSearch(ctx, query, limit*2)
	if err != nil {
		s.logger.Warn().Err(err).Msg("Vector search failed, falling back to keyword only")
		return keywordResults, nil
	}

	// RRF merge
	scores := make(map[string]float64) // event_id -> fused score
	eventMap := make(map[string]SemanticResult)

	for rank, r := range keywordResults {
		id := r.Event.EventID
		scores[id] += 1.0 / float64(k+rank+1)
		r.SearchType = "hybrid"
		eventMap[id] = r
	}
	for rank, r := range vectorResults {
		id := r.Event.EventID
		scores[id] += 1.0 / float64(k+rank+1)
		r.SearchType = "hybrid"
		if _, exists := eventMap[id]; !exists {
			eventMap[id] = r
		}
	}

	// Sort by fused score
	type scored struct {
		id    string
		score float64
	}
	var sorted_ []scored
	for id, score := range scores {
		sorted_ = append(sorted_, scored{id, score})
	}
	sort.Slice(sorted_, func(i, j int) bool {
		return sorted_[i].score > sorted_[j].score
	})

	var results []SemanticResult
	for i, s := range sorted_ {
		if i >= limit {
			break
		}
		r := eventMap[s.id]
		r.Score = s.score
		results = append(results, r)
	}

	return results, nil
}

func (s *SemanticSearch) keywordSearch(ctx context.Context, query string, limit int) ([]SemanticResult, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT event_id, source, event_type, severity, timestamp, received_at,
		       data, tags, metadata,
		       ts_rank(search_vector, plainto_tsquery('english', $1)) AS score
		FROM events
		WHERE search_vector @@ plainto_tsquery('english', $1)
		ORDER BY score DESC
		LIMIT $2
	`, query, limit)
	if err != nil {
		return nil, fmt.Errorf("keyword search: %w", err)
	}
	defer rows.Close()

	var results []SemanticResult
	for rows.Next() {
		var r SemanticResult
		err := rows.Scan(
			&r.Event.EventID, &r.Event.Source, &r.Event.EventType,
			&r.Event.Severity, &r.Event.Timestamp, &r.Event.ReceivedAt,
			&r.Event.Data, &r.Event.Tags, &r.Event.Metadata,
			&r.Score,
		)
		if err != nil {
			return nil, err
		}
		r.SearchType = "keyword"
		results = append(results, r)
	}

	return results, nil
}

// pgvectorString converts a float32 slice to pgvector's text format: "[1,2,3]"
func pgvectorString(v []float32) string {
	s := "["
	for i, f := range v {
		if i > 0 {
			s += ","
		}
		s += fmt.Sprintf("%g", math.Float64frombits(uint64(math.Float32bits(f))<<32>>32))
	}
	s += "]"
	return s
}
