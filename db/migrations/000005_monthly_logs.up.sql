-- monthly_logs materialized view

-- montly
CREATE MATERIALIZED VIEW monthly_logs
WITH (timescaledb.continuous) AS
SELECT 
    time_bucket('1 month', date) AS month,
    key_id,
    COUNT(*) AS request_count,
    COUNT(*) FILTER (WHERE fail) AS failed_requests,
    SUM(quota_value) FILTER (WHERE fail) AS failed_quota,
    SUM(quota_value) FILTER (WHERE NOT fail) AS used_quota
FROM logs
GROUP BY month, key_id
WITH NO DATA;

-- add continuous aggregate policy, recalculate every day
SELECT add_continuous_aggregate_policy('monthly_logs',
    start_offset => INTERVAL '3 month',
    end_offset => INTERVAL '1 day',
    schedule_interval => INTERVAL '1 day');    

CREATE INDEX idx_monthly_logs_key_id ON monthly_logs (key_id);
