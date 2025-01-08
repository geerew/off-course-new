-- +goose Up

--- Course
CREATE TABLE courses (
	id           TEXT PRIMARY KEY NOT NULL,
	title        TEXT NOT NULL,
	path         TEXT UNIQUE NOT NULL,
	card_path    TEXT,
	available    BOOLEAN NOT NULL DEFAULT FALSE,
	created_at   TEXT NOT NULL DEFAULT (STRFTIME('%Y-%m-%d %H:%M:%f', 'NOW')),
	updated_at   TEXT NOT NULL DEFAULT (STRFTIME('%Y-%m-%d %H:%M:%f', 'NOW'))
);

--- Progress of courses
CREATE TABLE courses_progress (
	id           TEXT PRIMARY KEY NOT NULL,
	course_id    TEXT NOT NULL UNIQUE,
	started      BOOLEAN NOT NULL DEFAULT FALSE,
	started_at   TEXT,
	percent      INTEGER NOT NULL DEFAULT 0,
	completed_at TEXT,
	created_at   TEXT NOT NULL DEFAULT (STRFTIME('%Y-%m-%d %H:%M:%f', 'NOW')),
	updated_at   TEXT NOT NULL DEFAULT (STRFTIME('%Y-%m-%d %H:%M:%f', 'NOW')),
	---
	FOREIGN KEY (course_id) REFERENCES courses (id) ON DELETE CASCADE
);

--- Assets
CREATE TABLE assets (
	id           TEXT PRIMARY KEY NOT NULL,
	course_id    TEXT NOT NULL,
	title        TEXT NOT NULL,
	prefix       INTEGER NOT NULL,
	chapter      TEXT,
	type         TEXT NOT NULL,
	path         TEXT UNIQUE NOT NULL,
	hash	     TEXT NOT NULL,
	created_at   TEXT NOT NULL DEFAULT (STRFTIME('%Y-%m-%d %H:%M:%f', 'NOW')),
	updated_at   TEXT NOT NULL DEFAULT (STRFTIME('%Y-%m-%d %H:%M:%f', 'NOW')),
	---
	FOREIGN KEY (course_id) REFERENCES courses (id) ON DELETE CASCADE
);

--- Progress of assets
CREATE TABLE assets_progress (
	id           TEXT PRIMARY KEY NOT NULL,
	asset_id     TEXT NOT NULL UNIQUE,
	video_pos    INTEGER NOT NULL DEFAULT 0,
	completed	 BOOLEAN NOT NULL DEFAULT FALSE,
	completed_at TEXT,
	created_at   TEXT NOT NULL DEFAULT (STRFTIME('%Y-%m-%d %H:%M:%f', 'NOW')),
	updated_at   TEXT NOT NULL DEFAULT (STRFTIME('%Y-%m-%d %H:%M:%f', 'NOW')),
	---
	FOREIGN KEY (asset_id) REFERENCES assets (id) ON DELETE CASCADE
);

--- Attachments
CREATE TABLE attachments (
	id          TEXT PRIMARY KEY NOT NULL,
	asset_id    TEXT NOT NULL,
	title       TEXT NOT NULL,
	path        TEXT UNIQUE NOT NULL,
	created_at  TEXT NOT NULL DEFAULT (STRFTIME('%Y-%m-%d %H:%M:%f', 'NOW')),
	updated_at  TEXT NOT NULL DEFAULT (STRFTIME('%Y-%m-%d %H:%M:%f', 'NOW')),
	---
	FOREIGN KEY (asset_id) REFERENCES assets (id) ON DELETE CASCADE
);

--- Scan jobs
CREATE TABLE scans (
	id          TEXT PRIMARY KEY NOT NULL,
	course_id   TEXT UNIQUE NOT NULL,
    status      TEXT NOT NULL DEFAULT 'waiting',
	created_at  TEXT NOT NULL DEFAULT (STRFTIME('%Y-%m-%d %H:%M:%f', 'NOW')),
	updated_at  TEXT NOT NULL DEFAULT (STRFTIME('%Y-%m-%d %H:%M:%f', 'NOW')),
	---
	FOREIGN KEY (course_id) REFERENCES courses (id) ON DELETE CASCADE
);

--- Tags
CREATE TABLE tags (
	id          TEXT PRIMARY KEY NOT NULL,
	tag         TEXT NOT NULL UNIQUE COLLATE NOCASE,
	created_at  TEXT NOT NULL DEFAULT (STRFTIME('%Y-%m-%d %H:%M:%f', 'NOW')),
	updated_at  TEXT NOT NULL DEFAULT (STRFTIME('%Y-%m-%d %H:%M:%f', 'NOW'))
);

--- Course tags (join table)
CREATE TABLE courses_tags (
	id          TEXT PRIMARY KEY NOT NULL,
	tag_id      TEXT NOT NULL,
	course_id   TEXT NOT NULL,
	created_at  TEXT NOT NULL DEFAULT (STRFTIME('%Y-%m-%d %H:%M:%f', 'NOW')),
	updated_at  TEXT NOT NULL DEFAULT (STRFTIME('%Y-%m-%d %H:%M:%f', 'NOW')),
	---
	FOREIGN KEY (tag_id) REFERENCES tags (id) ON DELETE CASCADE,
	FOREIGN KEY (course_id) REFERENCES courses (id) ON DELETE CASCADE,
	---
	CONSTRAINT unique_course_tag UNIQUE (tag_id, course_id)
);

--- Parameters (for application settings)
CREATE TABLE params (
    id           TEXT PRIMARY KEY NOT NULL,
    key          TEXT UNIQUE NOT NULL,
    value        TEXT NOT NULL,
    created_at   TEXT NOT NULL DEFAULT (STRFTIME('%Y-%m-%d %H:%M:%f', 'NOW')),
    updated_at   TEXT NOT NULL DEFAULT (STRFTIME('%Y-%m-%d %H:%M:%f', 'NOW'))
);

--- Users
CREATE TABLE users (
    id            TEXT PRIMARY KEY NOT NULL,
    username      TEXT UNIQUE NOT NULL COLLATE NOCASE,
	display_name  TEXT NOT NULL,
    password_hash TEXT NOT NULL,
    role          TEXT NOT NULL CHECK(role IN ('admin', 'user')),
    created_at    TEXT NOT NULL DEFAULT (STRFTIME('%Y-%m-%d %H:%M:%f', 'NOW')),
    updated_at    TEXT NOT NULL DEFAULT (STRFTIME('%Y-%m-%d %H:%M:%f', 'NOW'))
);

--- Sessions
CREATE TABLE sessions (
    id      TEXT PRIMARY KEY NOT NULL,
	data    BLOB NOT NULL,
	expires BIGINT NOT NULL
);