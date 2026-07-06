package queue

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	"github.com/vmihailenco/msgpack/v5"

	"github.com/atlasdb/atlasdb/pkg/models"
)

type MessageHandler func(ctx context.Context, events []models.Event) error

type Consumer struct {
	client       *redis.Client
	group        string
	consumerName string
	streams      []string
	batchSize    int64
	claimTimeout time.Duration
	maxRetries   int
	logger       zerolog.Logger
}

func NewConsumer(
	client *redis.Client,
	group string,
	consumerName string,
	numPartitions int,
	batchSize int,
	claimTimeout time.Duration,
	maxRetries int,
	logger zerolog.Logger,
) *Consumer {
	streams := make([]string, numPartitions)
	for i := 0; i < numPartitions; i++ {
		streams[i] = StreamKey(i)
	}

	return &Consumer{
		client:       client,
		group:        group,
		consumerName: consumerName,
		streams:      streams,
		batchSize:    int64(batchSize),
		claimTimeout: claimTimeout,
		maxRetries:   maxRetries,
		logger:       logger,
	}
}

func (c *Consumer) EnsureGroups(ctx context.Context) error {
	for _, stream := range c.streams {
		err := c.client.XGroupCreateMkStream(ctx, stream, c.group, "0").Err()
		if err != nil && err.Error() != "BUSYGROUP Consumer Group name already exists" {
			return fmt.Errorf("create consumer group for %s: %w", stream, err)
		}
	}
	return nil
}

func (c *Consumer) Run(ctx context.Context, handler MessageHandler) error {
	if err := c.EnsureGroups(ctx); err != nil {
		return err
	}

	c.logger.Info().
		Str("group", c.group).
		Str("consumer", c.consumerName).
		Strs("streams", c.streams).
		Msg("Starting consumer")

	// First process any pending (unacknowledged) messages
	if err := c.processPending(ctx, handler); err != nil {
		c.logger.Warn().Err(err).Msg("Error processing pending messages")
	}

	// Then read new messages
	for {
		select {
		case <-ctx.Done():
			c.logger.Info().Msg("Consumer shutting down")
			return ctx.Err()
		default:
		}

		if err := c.readAndProcess(ctx, handler); err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			c.logger.Error().Err(err).Msg("Error reading from stream")
			time.Sleep(time.Second)
		}
	}
}

func (c *Consumer) readAndProcess(ctx context.Context, handler MessageHandler) error {
	// Build XREADGROUP args: stream1 stream2 ... > > ...
	readStreams := make([]string, 0, len(c.streams)*2)
	readStreams = append(readStreams, c.streams...)
	for range c.streams {
		readStreams = append(readStreams, ">")
	}

	results, err := c.client.XReadGroup(ctx, &redis.XReadGroupArgs{
		Group:    c.group,
		Consumer: c.consumerName,
		Streams:  readStreams,
		Count:    c.batchSize,
		Block:    2 * time.Second,
	}).Result()
	if err != nil {
		if err == redis.Nil {
			return nil
		}
		return fmt.Errorf("xreadgroup: %w", err)
	}

	for _, stream := range results {
		events, msgIDs, err := c.decodeMessages(stream.Messages)
		if err != nil {
			c.logger.Error().Err(err).Str("stream", stream.Stream).Msg("Failed to decode messages")
			continue
		}

		if len(events) == 0 {
			continue
		}

		if err := handler(ctx, events); err != nil {
			c.logger.Error().Err(err).
				Str("stream", stream.Stream).
				Int("count", len(events)).
				Msg("Handler failed, messages will be retried")
			continue
		}

		// ACK successfully processed messages
		if err := c.client.XAck(ctx, stream.Stream, c.group, msgIDs...).Err(); err != nil {
			c.logger.Error().Err(err).Str("stream", stream.Stream).Msg("Failed to ACK messages")
		}
	}

	return nil
}

func (c *Consumer) processPending(ctx context.Context, handler MessageHandler) error {
	for _, stream := range c.streams {
		pending, err := c.client.XPendingExt(ctx, &redis.XPendingExtArgs{
			Stream: stream,
			Group:  c.group,
			Start:  "-",
			End:    "+",
			Count:  c.batchSize,
		}).Result()
		if err != nil {
			continue
		}

		for _, p := range pending {
			if p.Idle < c.claimTimeout {
				continue
			}

			if p.RetryCount >= int64(c.maxRetries) {
				c.moveToDLQ(ctx, stream, p.ID)
				continue
			}

			// Claim the stale message
			claimed, err := c.client.XClaim(ctx, &redis.XClaimArgs{
				Stream:   stream,
				Group:    c.group,
				Consumer: c.consumerName,
				MinIdle:  c.claimTimeout,
				Messages: []string{p.ID},
			}).Result()
			if err != nil {
				continue
			}

			events, msgIDs, err := c.decodeMessages(claimed)
			if err != nil {
				continue
			}

			if err := handler(ctx, events); err != nil {
				c.logger.Warn().Err(err).Str("msg_id", p.ID).Msg("Retry handler failed")
				continue
			}

			c.client.XAck(ctx, stream, c.group, msgIDs...)
		}
	}
	return nil
}

func (c *Consumer) moveToDLQ(ctx context.Context, stream, msgID string) {
	msgs, err := c.client.XRange(ctx, stream, msgID, msgID).Result()
	if err != nil || len(msgs) == 0 {
		return
	}

	c.client.XAdd(ctx, &redis.XAddArgs{
		Stream: DLQStream,
		Values: map[string]interface{}{
			"original_stream": stream,
			"original_id":     msgID,
			"data":            msgs[0].Values["data"],
		},
	})

	c.client.XAck(ctx, stream, c.group, msgID)

	c.logger.Warn().
		Str("stream", stream).
		Str("msg_id", msgID).
		Msg("Message moved to DLQ after max retries")
}

func (c *Consumer) decodeMessages(msgs []redis.XMessage) ([]models.Event, []string, error) {
	events := make([]models.Event, 0, len(msgs))
	ids := make([]string, 0, len(msgs))

	for _, msg := range msgs {
		raw, ok := msg.Values["data"].(string)
		if !ok {
			continue
		}

		var event models.Event
		if err := msgpack.Unmarshal([]byte(raw), &event); err != nil {
			c.logger.Warn().Err(err).Str("msg_id", msg.ID).Msg("Failed to unmarshal event")
			continue
		}

		events = append(events, event)
		ids = append(ids, msg.ID)
	}

	return events, ids, nil
}
