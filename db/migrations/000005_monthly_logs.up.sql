-- monthly_logs materialized view

-- monthly
CREATE MATERIALIZED VIEW monthly_logs
WITH (timescaledb.continuous) AS
SELECT 
    time_bucket('1 month', day) AS month,
    key_id,
    SUM(request_count) AS request_count,
    SUM(failed_requests) AS failed_requests,
    SUM(failed_quota) AS failed_quota,
    SUM(used_quota) AS used_quota
FROM daily_logs
GROUP BY month, key_id
WITH NO DATA;

-- add continuous aggregate policy, recalculate every day
SELECT add_continuous_aggregate_policy('monthly_logs',
    start_offset => INTERVAL '3 month',
    end_offset => INTERVAL '1 day',
    schedule_interval => INTERVAL '1 day');    

-- turn realtime select feature on
ALTER MATERIALIZED VIEW monthly_logs set (timescaledb.materialized_only = false);    

CREATE INDEX idx_monthly_logs_key_id ON monthly_logs (key_id);
