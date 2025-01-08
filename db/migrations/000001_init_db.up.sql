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

-- Table for operations
CREATE TABLE operations (
    id TEXT NOT NULL PRIMARY KEY,
    key_id TEXT NOT NULL,
    date TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    quota_value DOUBLE PRECISION NOT NULL DEFAULT 0,
    msg TEXT NOT NULL,
    FOREIGN KEY (key_id) REFERENCES keys (id)
);

-- Table for settings
CREATE TABLE settings (
    id TEXT NOT NULL PRIMARY KEY,
    data JSONB NOT NULL,
    updated TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

END;
