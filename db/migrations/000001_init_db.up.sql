-- initial db schema

BEGIN;

-- Table for keys
CREATE TABLE keys (
    id TEXT NOT NULL PRIMARY KEY,
    project TEXT NOT NULL,
    manual BOOLEAN NOT NULL,
    key TEXT NOT NULL,
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

CREATE UNIQUE INDEX uidx_keys_project_key_manual ON keys (project, key, manual);

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

-- -- Table for operationRecord
-- CREATE TABLE operation_record (
--     key_id TEXT NOT NULL,
--     operation_id TEXT NOT NULL,
--     date TIMESTAMPTZ,
--     quota_value FLOAT8,
--     msg TEXT
-- );

-- -- Table for settingsRecord
-- CREATE TABLE settings_record (
--     reset_started TIMESTAMPTZ,
--     next_reset TIMESTAMPTZ,
--     updated TIMESTAMPTZ
-- );

-- type keyRecord struct {
-- 	Key              string    `bson:"key"`
-- 	KeyID            string    `bson:"keyID,omitempty"`
-- 	Manual           bool      `bson:"manual"`
-- 	ValidTo          time.Time `bson:"validTo,omitempty"`
-- 	Limit            float64   `bson:"limit,omitempty"`
-- 	QuotaValue       float64   `bson:"quotaValue"`
-- 	QuotaValueFailed float64   `bson:"quotaValueFailed,omitempty"`
-- 	Created          time.Time `bson:"created,omitempty"`
-- 	Updated          time.Time `bson:"updated,omitempty"`
-- 	LastUsed         time.Time `bson:"lastUsed,omitempty"`
-- 	ResetAt          time.Time `bson:"resetAt,omitempty"`
-- 	LastIP           string    `bson:"lastIP,omitempty"`
-- 	Disabled         bool      `bson:"disabled,omitempty"`
-- 	IPWhiteList      string    `bson:"IPWhiteList,omitempty"`
-- 	Description      string    `bson:"description,omitempty"`
-- 	Tags             []string  `bson:"tags,omitempty"`
-- 	ExternalID       string    `bson:"externalID,omitempty"`
-- }

-- type logRecord struct {
-- 	Key          string    `bson:"key,omitempty"`
-- 	KeyID        string    `bson:"keyID,omitempty"`
-- 	URL          string    `bson:"url,omitempty"`
-- 	QuotaValue   float64   `bson:"quotaValue,omitempty"`
-- 	Date         time.Time `bson:"date,omitempty"`
-- 	IP           string    `bson:"ip,omitempty"`
-- 	Value        string    `bson:"value,omitempty"`
-- 	Fail         bool      `bson:"fail,omitempty"`
-- 	ResponseCode int       `bson:"responseCode,omitempty"`

-- 	RequestID string `bson:"requestID,omitempty"`
-- 	ErrorMsg  string `bson:"errorMsg,omitempty"`
-- }

-- type keyMapRecord struct {
-- 	KeyID      string    `bson:"keyID"`
-- 	KeyHash    string    `bson:"keyHash"`
-- 	ExternalID string    `bson:"externalID"`
-- 	Project    string    `bson:"project"`
-- 	Created    time.Time `bson:"created,omitempty"`
-- 	Old        []oldKey  `bson:"old,omitempty"`
-- }

-- type oldKey struct {
-- 	KeyHash   string    `bson:"keyHash"`
-- 	ChangedOn time.Time `bson:"changedOn,omitempty"`
-- }

-- type operationRecord struct {
-- 	KeyID       string    `bson:"keyID"`
-- 	OperationID string    `bson:"operationID"`
-- 	Date        time.Time `bson:"date,omitempty"`
-- 	QuotaValue  float64   `bson:"quotaValue,omitempty"`
-- 	Msg         string    `bson:"msg,omitempty"`
-- }

-- type settingsRecord struct {
-- 	ResetStarted time.Time `bson:"resetStarted,omitempty"`
-- 	NextReset    time.Time `bson:"nextReset,omitempty"`
-- 	Updated      time.Time `bson:"updated,omitempty"`
-- }

END;
