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

----------------------------------------------------
-- ALTER SCANS TABLE TO ADD STATUS COLUMN DEFAULT --
----------------------------------------------------

-- Create a new temporary table with the default for status
CREATE TABLE scans_temp (
    id          TEXT PRIMARY KEY NOT NULL,
    course_id   TEXT UNIQUE NOT NULL,
    status      TEXT NOT NULL DEFAULT 'waiting',
    created_at  TEXT NOT NULL DEFAULT (STRFTIME('%Y-%m-%d %H:%M:%f', 'NOW')),
    updated_at  TEXT NOT NULL DEFAULT (STRFTIME('%Y-%m-%d %H:%M:%f', 'NOW')),
    FOREIGN KEY (course_id) REFERENCES courses (id) ON DELETE CASCADE
);

-- Copy data from the original table to the temporary table
INSERT INTO scans_temp (id, course_id, status, created_at, updated_at)
SELECT id, course_id, status, created_at, updated_at FROM scans;

-- Drop the original table
DROP TABLE scans;

-- Rename the temporary table to the original name
ALTER TABLE scans_temp RENAME TO scans;