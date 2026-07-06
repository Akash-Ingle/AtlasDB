package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
)

type Cache struct {
	client  *redis.Client
	prefix  string
	logger  zerolog.Logger
	hits    prometheus.Counter
	misses  prometheus.Counter
}

func New(client *redis.Client, prefix string, logger zerolog.Logger) *Cache {
	c := &Cache{
		client: client,
		prefix: prefix,
		logger: logger,
		hits: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "atlas_cache_hits_total",
			Help: "Total cache hits",
		}),
		misses: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "atlas_cache_misses_total",
			Help: "Total cache misses",
		}),
	}
	prometheus.Register(c.hits)
	prometheus.Register(c.misses)
	return c
}

func (c *Cache) key(k string) string {
	return fmt.Sprintf("%s:%s", c.prefix, k)
}

// Get retrieves a cached value. Returns false on miss.
func (c *Cache) Get(ctx context.Context, key string, dest interface{}) bool {
	data, err := c.client.Get(ctx, c.key(key)).Bytes()
	if err != nil {
		c.misses.Inc()
		return false
	}

	if err := json.Unmarshal(data, dest); err != nil {
		c.misses.Inc()
		return false
	}

	c.hits.Inc()
	return true
}

// Set stores a value with TTL.
func (c *Cache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) {
	data, err := json.Marshal(value)
	if err != nil {
		return
	}
	c.client.Set(ctx, c.key(key), data, ttl)
}

// Delete removes a cached value.
func (c *Cache) Delete(ctx context.Context, key string) {
	c.client.Del(ctx, c.key(key))
}

// DeletePattern removes all keys matching a pattern.
func (c *Cache) DeletePattern(ctx context.Context, pattern string) {
	iter := c.client.Scan(ctx, 0, c.key(pattern), 100).Iterator()
	for iter.Next(ctx) {
		c.client.Del(ctx, iter.Val())
	}
}
