-- 001_initial.sql: Core schema for AtlasDB

-- Enable required extensions
CREATE EXTENSION IF NOT EXISTS "pgcrypto";
CREATE EXTENSION IF NOT EXISTS "vector";

-- =============================================================================
-- Events (partitioned by day)
-- =============================================================================
CREATE TABLE IF NOT EXISTS events (
    event_id    TEXT        NOT NULL,
    source      TEXT        NOT NULL,
    event_type  TEXT        NOT NULL,
    severity    TEXT        NOT NULL DEFAULT 'info',
    timestamp   TIMESTAMPTZ NOT NULL,
    received_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    data        JSONB       NOT NULL DEFAULT '{}',
    tags        TEXT[]      DEFAULT '{}',
    metadata    JSONB       DEFAULT '{}',

    search_vector TSVECTOR,
    embedding     vector(384),

    PRIMARY KEY (event_id, timestamp)
) PARTITION BY RANGE (timestamp);

-- Trigger to auto-populate search_vector on INSERT
CREATE OR REPLACE FUNCTION events_search_vector_update() RETURNS trigger AS $$
BEGIN
    NEW.search_vector :=
        setweight(to_tsvector('english', coalesce(NEW.source, '')), 'A') ||
        setweight(to_tsvector('english', coalesce(NEW.event_type, '')), 'A') ||
        setweight(to_tsvector('english', coalesce(NEW.severity, '')), 'B') ||
        setweight(to_tsvector('english', coalesce(NEW.data::text, '')), 'C');
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Indices on the parent table (inherited by partitions)
CREATE INDEX IF NOT EXISTS idx_events_source
    ON events (source, timestamp DESC);

CREATE INDEX IF NOT EXISTS idx_events_severity
    ON events (severity, timestamp DESC)
    WHERE severity IN ('error', 'fatal');

CREATE INDEX IF NOT EXISTS idx_events_search
    ON events USING GIN (search_vector);

CREATE INDEX IF NOT EXISTS idx_events_tags
    ON events USING GIN (tags);

CREATE INDEX IF NOT EXISTS idx_events_data_status
    ON events ((data->>'status'), timestamp DESC);

-- =============================================================================
-- Aggregation tables
-- =============================================================================
CREATE TABLE IF NOT EXISTS event_counts_1m (
    bucket      TIMESTAMPTZ NOT NULL,
    source      TEXT        NOT NULL,
    event_type  TEXT        NOT NULL,
    severity    TEXT        NOT NULL,
    count       BIGINT      NOT NULL DEFAULT 0,
    PRIMARY KEY (bucket, source, event_type, severity)
) PARTITION BY RANGE (bucket);

CREATE TABLE IF NOT EXISTS latency_stats_1m (
    bucket      TIMESTAMPTZ NOT NULL,
    source      TEXT        NOT NULL,
    endpoint    TEXT        NOT NULL,
    count       BIGINT      NOT NULL,
    sum_ms      DOUBLE PRECISION NOT NULL,
    min_ms      DOUBLE PRECISION NOT NULL,
    max_ms      DOUBLE PRECISION NOT NULL,
    p50_ms      DOUBLE PRECISION,
    p95_ms      DOUBLE PRECISION,
    p99_ms      DOUBLE PRECISION,
    PRIMARY KEY (bucket, source, endpoint)
) PARTITION BY RANGE (bucket);

-- =============================================================================
-- Users and Auth
-- =============================================================================
CREATE TABLE IF NOT EXISTS users (
    user_id       UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    email         TEXT        UNIQUE NOT NULL,
    password_hash TEXT        NOT NULL,
    role          TEXT        NOT NULL DEFAULT 'viewer',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS api_keys (
    key_id     UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID        NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    key_hash   TEXT        NOT NULL,
    key_prefix TEXT        NOT NULL,
    name       TEXT        NOT NULL,
    scopes     TEXT[]      NOT NULL,
    last_used_at TIMESTAMPTZ,
    expires_at   TIMESTAMPTZ,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_api_keys_prefix ON api_keys (key_prefix);

-- =============================================================================
-- Alert Rules and History
-- =============================================================================
CREATE TABLE IF NOT EXISTS alert_rules (
    rule_id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    name             TEXT        NOT NULL,
    description      TEXT,
    condition        JSONB       NOT NULL,
    severity         TEXT        NOT NULL DEFAULT 'warning',
    channels         JSONB       NOT NULL DEFAULT '[]',
    enabled          BOOLEAN     NOT NULL DEFAULT true,
    cooldown_seconds INTEGER     NOT NULL DEFAULT 300,
    created_by       UUID        REFERENCES users(user_id),
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS alert_events (
    alert_event_id UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    rule_id        UUID        NOT NULL REFERENCES alert_rules(rule_id) ON DELETE CASCADE,
    status         TEXT        NOT NULL,
    fired_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    resolved_at    TIMESTAMPTZ,
    value          DOUBLE PRECISION,
    context        JSONB,
    notified       BOOLEAN     NOT NULL DEFAULT false
);

CREATE INDEX IF NOT EXISTS idx_alert_events_rule
    ON alert_events (rule_id, fired_at DESC);

CREATE INDEX IF NOT EXISTS idx_alert_events_status
    ON alert_events (status, fired_at DESC);

-- =============================================================================
-- Partition management function
-- =============================================================================
CREATE OR REPLACE FUNCTION create_daily_partition(
    parent_table TEXT,
    partition_date DATE
) RETURNS void AS $$
DECLARE
    partition_name TEXT;
    start_date DATE;
    end_date DATE;
BEGIN
    start_date := partition_date;
    end_date := partition_date + INTERVAL '1 day';
    partition_name := parent_table || '_' || to_char(partition_date, 'YYYY_MM_DD');

    EXECUTE format(
        'CREATE TABLE IF NOT EXISTS %I PARTITION OF %I FOR VALUES FROM (%L) TO (%L)',
        partition_name, parent_table, start_date, end_date
    );

    -- Attach the search vector trigger to each event partition
    IF parent_table = 'events' THEN
        EXECUTE format(
            'CREATE TRIGGER trg_%I_search_vector BEFORE INSERT ON %I FOR EACH ROW EXECUTE FUNCTION events_search_vector_update()',
            partition_name, partition_name
        );
    END IF;
END;
$$ LANGUAGE plpgsql;

-- Create partitions for today and the next 7 days
DO $$
DECLARE
    d DATE;
BEGIN
    FOR d IN SELECT generate_series(
        CURRENT_DATE - INTERVAL '1 day',
        CURRENT_DATE + INTERVAL '7 days',
        INTERVAL '1 day'
    )::date
    LOOP
        PERFORM create_daily_partition('events', d);
        PERFORM create_daily_partition('event_counts_1m', d);
        PERFORM create_daily_partition('latency_stats_1m', d);
    END LOOP;
END $$;
