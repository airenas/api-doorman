-- daily_logs materialized view

CREATE MATERIALIZED VIEW daily_logs
WITH (timescaledb.continuous) AS
SELECT 
    time_bucket('1 day', date) AS day,
    key_id,
    COUNT(*) AS request_count,
    COUNT(*) FILTER (WHERE fail) AS failed_requests,
    SUM(quota_value) FILTER (WHERE fail) AS failed_quota,
    SUM(quota_value) FILTER (WHERE NOT fail) AS used_quota
FROM logs
GROUP BY day, key_id
WITH NO DATA;

-- add continuous aggregate policy, recalculate every hour
SELECT add_continuous_aggregate_policy('daily_logs',
    start_offset => INTERVAL '3 day',
    end_offset => INTERVAL '1 hour',
    schedule_interval => INTERVAL '1 hour');

-- turn realtime select feature on
ALTER MATERIALIZED VIEW daily_logs set (timescaledb.materialized_only = false);

CREATE INDEX idx_daily_logs_key_id ON daily_logs (key_id);