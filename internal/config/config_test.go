package config

import (
	"os"
	"testing"
)

func TestLoadDefaults(t *testing.T) {
	// Clear any env vars that might interfere
	os.Unsetenv("SERVER_PORT")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if cfg.Server.Port != 8080 {
		t.Errorf("Server.Port = %d, want 8080", cfg.Server.Port)
	}

	if cfg.Postgres.Host != "localhost" {
		t.Errorf("Postgres.Host = %q, want localhost", cfg.Postgres.Host)
	}

	if cfg.Ingest.NumPartitions != 4 {
		t.Errorf("Ingest.NumPartitions = %d, want 4", cfg.Ingest.NumPartitions)
	}

	if cfg.Processor.Workers != 4 {
		t.Errorf("Processor.Workers = %d, want 4", cfg.Processor.Workers)
	}
}

func TestLoadEnvOverride(t *testing.T) {
	os.Setenv("SERVER_PORT", "9090")
	defer os.Unsetenv("SERVER_PORT")

	os.Setenv("POSTGRES_HOST", "db.example.com")
	defer os.Unsetenv("POSTGRES_HOST")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if cfg.Server.Port != 9090 {
		t.Errorf("Server.Port = %d, want 9090", cfg.Server.Port)
	}

	if cfg.Postgres.Host != "db.example.com" {
		t.Errorf("Postgres.Host = %q, want db.example.com", cfg.Postgres.Host)
	}
}

func TestPostgresDSN(t *testing.T) {
	cfg := PostgresConfig{
		Host:     "localhost",
		Port:     5432,
		User:     "atlas",
		Password: "secret",
		Database: "atlasdb",
		SSLMode:  "disable",
	}

	want := "postgres://atlas:secret@localhost:5432/atlasdb?sslmode=disable"
	if got := cfg.DSN(); got != want {
		t.Errorf("DSN() = %q, want %q", got, want)
	}
}

func TestValidation(t *testing.T) {
	cfg := &Config{}
	cfg.Server.Port = -1
	cfg.Ingest.NumPartitions = 4
	cfg.Processor.Workers = 4

	if err := cfg.validate(); err == nil {
		t.Error("expected validation error for invalid port")
	}

	cfg.Server.Port = 8080
	cfg.Ingest.NumPartitions = 0
	if err := cfg.validate(); err == nil {
		t.Error("expected validation error for zero partitions")
	}
}
