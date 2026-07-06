package queue

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	"github.com/vmihailenco/msgpack/v5"

	"github.com/atlasdb/atlasdb/pkg/models"
)

const (
	StreamPrefix = "events:"
	DLQStream    = "dead-letter:events"
)

func StreamKey(partition int) string {
	return fmt.Sprintf("%s%d", StreamPrefix, partition)
}

type Producer struct {
	client        *redis.Client
	numPartitions int
	logger        zerolog.Logger
}

func NewProducer(client *redis.Client, numPartitions int, logger zerolog.Logger) *Producer {
	return &Producer{
		client:        client,
		numPartitions: numPartitions,
		logger:        logger,
	}
}

func (p *Producer) Publish(ctx context.Context, events []models.Event) error {
	pipe := p.client.Pipeline()

	for _, event := range events {
		partition := partitionFor(event.Source, p.numPartitions)
		key := StreamKey(partition)

		data, err := msgpack.Marshal(event)
		if err != nil {
			return fmt.Errorf("marshal event %s: %w", event.EventID, err)
		}

		pipe.XAdd(ctx, &redis.XAddArgs{
			Stream: key,
			Values: map[string]interface{}{
				"data": data,
			},
		})
	}

	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("publish events: %w", err)
	}

	return nil
}

func (p *Producer) QueueDepth(ctx context.Context) (int64, error) {
	var total int64
	for i := 0; i < p.numPartitions; i++ {
		length, err := p.client.XLen(ctx, StreamKey(i)).Result()
		if err != nil {
			return 0, fmt.Errorf("xlen %s: %w", StreamKey(i), err)
		}
		total += length
	}
	return total, nil
}

func partitionFor(source string, numPartitions int) int {
	var hash uint32
	for _, c := range source {
		hash = hash*31 + uint32(c)
	}
	return int(hash % uint32(numPartitions))
}
