
-- SELECT drop_continuous_aggregate_policy('monthly_logs');
DROP INDEX IF EXISTS idx_monthly_logs_key_id;
DROP MATERIALIZED VIEW IF EXISTS monthly_logs;