package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Server    ServerConfig
	Postgres  PostgresConfig
	Redis     RedisConfig
	Auth      AuthConfig
	Ingest    IngestConfig
	Processor ProcessorConfig
	AI        AIConfig
	OTel      OTelConfig
	Log       LogConfig
}

type AIConfig struct {
	Provider       string // "openai" or "anthropic"
	OpenAIKey      string
	OpenAIModel    string
	OpenAIBaseURL  string
	AnthropicKey   string
	AnthropicModel string
	EmbedModel     string
	Enabled        bool
}

type ServerConfig struct {
	Host            string
	Port            int
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ShutdownTimeout time.Duration
}

type PostgresConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Database string
	SSLMode  string
	MaxConns int32
	MinConns int32
}

func (p PostgresConfig) DSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		p.User, p.Password, p.Host, p.Port, p.Database, p.SSLMode,
	)
}

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

type AuthConfig struct {
	JWTSecret        string
	AccessTokenTTL   time.Duration
	RefreshTokenTTL  time.Duration
}

type IngestConfig struct {
	MaxBatchSize       int
	MaxEventSizeBytes  int
	QueueBackpressure  int64
	NumPartitions      int
}

type ProcessorConfig struct {
	Workers        int
	BatchSize      int
	ClaimTimeout   time.Duration
	MaxRetries     int
	ConsumerGroup  string
}

type OTelConfig struct {
	Enabled      bool
	OTLPEndpoint string
}

type LogConfig struct {
	Level  string
	Format string // "json" or "console"
}

func Load() (*Config, error) {
	cfg := &Config{
		Server: ServerConfig{
			Host:            envStr("SERVER_HOST", "0.0.0.0"),
			Port:            envInt("SERVER_PORT", 8080),
			ReadTimeout:     envDuration("SERVER_READ_TIMEOUT", 15*time.Second),
			WriteTimeout:    envDuration("SERVER_WRITE_TIMEOUT", 15*time.Second),
			ShutdownTimeout: envDuration("SERVER_SHUTDOWN_TIMEOUT", 10*time.Second),
		},
		Postgres: PostgresConfig{
			Host:     envStr("POSTGRES_HOST", "localhost"),
			Port:     envInt("POSTGRES_PORT", 5432),
			User:     envStr("POSTGRES_USER", "atlas"),
			Password: envStr("POSTGRES_PASSWORD", "atlas"),
			Database: envStr("POSTGRES_DB", "atlasdb"),
			SSLMode:  envStr("POSTGRES_SSLMODE", "disable"),
			MaxConns: int32(envInt("POSTGRES_MAX_CONNS", 20)),
			MinConns: int32(envInt("POSTGRES_MIN_CONNS", 5)),
		},
		Redis: RedisConfig{
			Addr:     envStr("REDIS_ADDR", "localhost:6379"),
			Password: envStr("REDIS_PASSWORD", ""),
			DB:       envInt("REDIS_DB", 0),
		},
		Auth: AuthConfig{
			JWTSecret:       envStr("JWT_SECRET", "atlas-dev-secret-change-in-production"),
			AccessTokenTTL:  envDuration("JWT_ACCESS_TTL", 15*time.Minute),
			RefreshTokenTTL: envDuration("JWT_REFRESH_TTL", 7*24*time.Hour),
		},
		Ingest: IngestConfig{
			MaxBatchSize:      envInt("INGEST_MAX_BATCH", 1000),
			MaxEventSizeBytes: envInt("INGEST_MAX_EVENT_BYTES", 1048576), // 1MB
			QueueBackpressure: int64(envInt("INGEST_QUEUE_BACKPRESSURE", 100000)),
			NumPartitions:     envInt("INGEST_NUM_PARTITIONS", 4),
		},
		Processor: ProcessorConfig{
			Workers:       envInt("PROCESSOR_WORKERS", 4),
			BatchSize:     envInt("PROCESSOR_BATCH_SIZE", 100),
			ClaimTimeout:  envDuration("PROCESSOR_CLAIM_TIMEOUT", 30*time.Second),
			MaxRetries:    envInt("PROCESSOR_MAX_RETRIES", 3),
			ConsumerGroup: envStr("PROCESSOR_CONSUMER_GROUP", "atlas-processors"),
		},
		OTel: OTelConfig{
			Enabled:      envStr("OTEL_ENABLED", "false") == "true",
			OTLPEndpoint: envStr("OTEL_EXPORTER_OTLP_ENDPOINT", "localhost:4317"),
		},
		AI: AIConfig{
			Provider:       envStr("AI_PROVIDER", "openai"),
			OpenAIKey:      envStr("OPENAI_API_KEY", ""),
			OpenAIModel:    envStr("OPENAI_MODEL", "gpt-4o"),
			OpenAIBaseURL:  envStr("OPENAI_BASE_URL", "https://api.openai.com/v1"),
			AnthropicKey:   envStr("ANTHROPIC_API_KEY", ""),
			AnthropicModel: envStr("ANTHROPIC_MODEL", "claude-sonnet-4-20250514"),
			EmbedModel:     envStr("EMBED_MODEL", "text-embedding-3-small"),
			Enabled:        envStr("AI_ENABLED", "false") == "true",
		},
		Log: LogConfig{
			Level:  envStr("LOG_LEVEL", "info"),
			Format: envStr("LOG_FORMAT", "json"),
		},
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}
	return cfg, nil
}

func (c *Config) validate() error {
	if c.Server.Port < 1 || c.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", c.Server.Port)
	}
	if c.Ingest.NumPartitions < 1 {
		return fmt.Errorf("ingest partitions must be >= 1")
	}
	if c.Processor.Workers < 1 {
		return fmt.Errorf("processor workers must be >= 1")
	}
	return nil
}

func envStr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return fallback
}

func envDuration(key string, fallback time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return fallback
}
