package redis

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"

	"github.com/atlasdb/atlasdb/internal/config"
)

func NewClient(ctx context.Context, cfg config.RedisConfig, logger zerolog.Logger) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
		PoolSize: 20,
	})

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("ping redis: %w", err)
	}

	logger.Info().
		Str("addr", cfg.Addr).
		Msg("Connected to Redis")

	return client, nil
}
