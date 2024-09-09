-- +goose Up

--- Parameters table for application settings
CREATE TABLE params (
    id           TEXT PRIMARY KEY NOT NULL,
    key          TEXT UNIQUE NOT NULL,
    value        TEXT NOT NULL,
    created_at   TEXT NOT NULL DEFAULT (STRFTIME('%Y-%m-%d %H:%M:%f', 'NOW')),
    updated_at   TEXT NOT NULL DEFAULT (STRFTIME('%Y-%m-%d %H:%M:%f', 'NOW'))
);

-- Insert initial parameter to track admin creation
INSERT INTO params (id, key, value) VALUES (
    'wV0418r0Rr',
    'hasAdmin',
    'false'
);

--- User information
CREATE TABLE users (
    id            TEXT PRIMARY KEY NOT NULL,
    username      TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    role          TEXT NOT NULL CHECK(role IN ('admin', 'user')),
    created_at    TEXT NOT NULL DEFAULT (STRFTIME('%Y-%m-%d %H:%M:%f', 'NOW')),
    updated_at    TEXT NOT NULL DEFAULT (STRFTIME('%Y-%m-%d %H:%M:%f', 'NOW'))
);