-- initial db schema

BEGIN;

-- Table for keys
CREATE TABLE keys (
    id TEXT NOT NULL PRIMARY KEY,
    project TEXT NOT NULL,
    manual BOOLEAN NOT NULL,
    key_hash TEXT NOT NULL,
    valid_to TIMESTAMPTZ,
    quota_limit DOUBLE PRECISION NOT NULL DEFAULT 0,
    quota_value DOUBLE PRECISION NOT NULL DEFAULT 0,
    quota_value_failed DOUBLE PRECISION NOT NULL DEFAULT 0,
    last_used TIMESTAMPTZ,
    reset_at TIMESTAMPTZ,
    last_ip TEXT,
    disabled BOOLEAN NOT NULL DEFAULT FALSE,
    ip_white_list TEXT,
    description TEXT,
    tags TEXT[], 
    external_id TEXT,

    created TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX uidx_keys_project_key_manual ON keys (project, key_hash, manual);

-- Table for logs
CREATE TABLE logs (
    key_id TEXT NOT NULL,
    url TEXT,
    quota_value DOUBLE PRECISION NOT NULL DEFAULT 0,
    date TIMESTAMPTZ NOT NULL,
    ip TEXT,
    value TEXT,
    fail BOOLEAN NOT NULL DEFAULT FALSE,
    response_code INT,
    request_id TEXT,
    error_msg TEXT, 
    FOREIGN KEY (key_id) REFERENCES keys (id)
);

CREATE INDEX idx_logs_date ON logs (date);
CREATE INDEX idx_logs_request_id ON logs (request_id);

-- -- Table for keyMapRecord
-- CREATE TABLE key_map_record (
--     key_id TEXT NOT NULL,
--     key_hash TEXT NOT NULL,
--     external_id TEXT,
--     project TEXT NOT NULL,
--     created TIMESTAMPTZ
-- );

-- -- Table for oldKey (used as a separate table to model the array of old keys)
-- CREATE TABLE old_key (
--     key_map_record_id SERIAL PRIMARY KEY, -- Links to key_map_record
--     key_hash TEXT NOT NULL,
--     changed_on TIMESTAMPTZ
-- );

-- Table for operations
CREATE TABLE operations (
    id TEXT NOT NULL PRIMARY KEY,
    key_id TEXT NOT NULL,
    date TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    quota_value DOUBLE PRECISION NOT NULL DEFAULT 0,
    msg TEXT NOT NULL,
    FOREIGN KEY (key_id) REFERENCES keys (id)
);

-- -- Table for settingsRecord
-- CREATE TABLE settings_record (
--     reset_started TIMESTAMPTZ,
--     next_reset TIMESTAMPTZ,
--     updated TIMESTAMPTZ
-- );


-- type oldKey struct {
-- 	KeyHash   string    `bson:"keyHash"`
-- 	ChangedOn time.Time `bson:"changedOn,omitempty"`
-- }

-- type settingsRecord struct {
-- 	ResetStarted time.Time `bson:"resetStarted,omitempty"`
-- 	NextReset    time.Time `bson:"nextReset,omitempty"`
-- 	Updated      time.Time `bson:"updated,omitempty"`
-- }

END;
