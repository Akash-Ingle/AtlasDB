-- 002_hourly_rollup.sql: Hourly aggregation table for analytics rollups

CREATE TABLE IF NOT EXISTS event_counts_1h (
    bucket      TIMESTAMPTZ NOT NULL,
    source      TEXT        NOT NULL,
    event_type  TEXT        NOT NULL,
    severity    TEXT        NOT NULL,
    count       BIGINT      NOT NULL DEFAULT 0,
    PRIMARY KEY (bucket, source, event_type, severity)
) PARTITION BY RANGE (bucket);

-- Create partitions for recent days
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
        PERFORM create_daily_partition('event_counts_1h', d);
    END LOOP;
END $$;
