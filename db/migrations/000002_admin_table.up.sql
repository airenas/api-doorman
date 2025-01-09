-- administrators table

BEGIN;

-- Keeps users that can issue keys
CREATE TABLE administrators (
    id TEXT NOT NULL PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,
    key_hash TEXT NOT NULL,

    max_valid_to TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    max_limit DOUBLE PRECISION NOT NULL DEFAULT 0,
    projects TEXT[], 

    disabled BOOLEAN NOT NULL DEFAULT FALSE,
    
    created TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- add column to keys table
ALTER TABLE keys ADD COLUMN adm_id TEXT;
ALTER TABLE keys ADD CONSTRAINT fk_administrator FOREIGN KEY (adm_id) REFERENCES administrators (id);
CREATE UNIQUE INDEX uidx_administrators_key_hash ON administrators (key_hash);

END;

