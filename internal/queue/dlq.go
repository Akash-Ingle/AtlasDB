package queue

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
)

type DLQManager struct {
	client *redis.Client
	logger zerolog.Logger
}

func NewDLQManager(client *redis.Client, logger zerolog.Logger) *DLQManager {
	return &DLQManager{client: client, logger: logger}
}

type DLQMessage struct {
	ID     string            `json:"id"`
	Stream string            `json:"stream"`
	Data   map[string]string `json:"data"`
}

type DLQStats struct {
	Stream string `json:"stream"`
	Length int64  `json:"length"`
}

const dlqStreamPrefix = "dlq:"

// Stats returns the count of messages in each DLQ stream.
func (d *DLQManager) Stats(ctx context.Context, partitions int) ([]DLQStats, error) {
	var stats []DLQStats
	for i := 0; i < partitions; i++ {
		stream := fmt.Sprintf("%s%d", dlqStreamPrefix, i)
		length, err := d.client.XLen(ctx, stream).Result()
		if err != nil {
			continue
		}
		if length > 0 {
			stats = append(stats, DLQStats{Stream: stream, Length: length})
		}
	}
	return stats, nil
}

// List returns messages from a DLQ stream.
func (d *DLQManager) List(ctx context.Context, stream string, count int64) ([]DLQMessage, error) {
	if count <= 0 {
		count = 50
	}
	msgs, err := d.client.XRange(ctx, stream, "-", "+").Result()
	if err != nil {
		return nil, fmt.Errorf("xrange %s: %w", stream, err)
	}

	var result []DLQMessage
	for i, msg := range msgs {
		if int64(i) >= count {
			break
		}
		data := make(map[string]string)
		for k, v := range msg.Values {
			data[k] = fmt.Sprintf("%v", v)
		}
		result = append(result, DLQMessage{
			ID:     msg.ID,
			Stream: stream,
			Data:   data,
		})
	}
	return result, nil
}

// Retry moves messages from DLQ back to the source stream for reprocessing.
func (d *DLQManager) Retry(ctx context.Context, dlqStream string, messageIDs []string, sourcePartition int) (int, error) {
	sourceStream := fmt.Sprintf("events:%d", sourcePartition)
	retried := 0

	for _, id := range messageIDs {
		// Read the message from DLQ
		msgs, err := d.client.XRange(ctx, dlqStream, id, id).Result()
		if err != nil || len(msgs) == 0 {
			continue
		}

		// Re-publish to source stream
		args := make([]interface{}, 0, len(msgs[0].Values)*2)
		for k, v := range msgs[0].Values {
			args = append(args, k, v)
		}
		_, err = d.client.XAdd(ctx, &redis.XAddArgs{
			Stream: sourceStream,
			Values: msgs[0].Values,
		}).Result()
		if err != nil {
			d.logger.Error().Err(err).Str("id", id).Msg("Failed to retry DLQ message")
			continue
		}

		// Remove from DLQ
		d.client.XDel(ctx, dlqStream, id)
		retried++
	}

	d.logger.Info().Int("retried", retried).Str("dlq", dlqStream).Msg("DLQ messages retried")
	return retried, nil
}

// Purge removes all messages from a DLQ stream.
func (d *DLQManager) Purge(ctx context.Context, stream string) (int64, error) {
	length, _ := d.client.XLen(ctx, stream).Result()
	if err := d.client.Del(ctx, stream).Err(); err != nil {
		return 0, fmt.Errorf("purge %s: %w", stream, err)
	}
	d.logger.Info().Str("stream", stream).Int64("purged", length).Msg("DLQ purged")
	return length, nil
}
