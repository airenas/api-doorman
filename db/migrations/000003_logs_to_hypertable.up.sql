-- make logs as hypertable

BEGIN; 

SELECT create_hypertable('logs', 'date', if_not_exists => TRUE);
-- add retention policy
SELECT add_retention_policy('logs', INTERVAL '200 days');

END;

