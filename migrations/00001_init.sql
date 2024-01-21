-- +goose Up

--- Course information
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

--- Assets information
CREATE TABLE assets (
	id           TEXT PRIMARY KEY NOT NULL,
	course_id    TEXT NOT NULL,
	title        TEXT NOT NULL,
	prefix       INTEGER NOT NULL,
	chapter      TEXT,
	type         TEXT NOT NULL,
	path         TEXT UNIQUE NOT NULL,
	created_at   TEXT NOT NULL DEFAULT (STRFTIME('%Y-%m-%d %H:%M:%f', 'NOW')),
	updated_at   TEXT NOT NULL DEFAULT (STRFTIME('%Y-%m-%d %H:%M:%f', 'NOW')),
	---
	FOREIGN KEY (course_id) REFERENCES courses (id) ON DELETE CASCADE
);

--- Progress of assets
CREATE TABLE assets_progress (
	id           TEXT PRIMARY KEY NOT NULL,
	asset_id     TEXT NOT NULL UNIQUE,
	course_id     TEXT NOT NULL,
	video_pos    INTEGER NOT NULL DEFAULT 0,
	completed	 BOOLEAN NOT NULL DEFAULT FALSE,
	completed_at TEXT,
	created_at   TEXT NOT NULL DEFAULT (STRFTIME('%Y-%m-%d %H:%M:%f', 'NOW')),
	updated_at   TEXT NOT NULL DEFAULT (STRFTIME('%Y-%m-%d %H:%M:%f', 'NOW')),
	---
	FOREIGN KEY (course_id) REFERENCES courses (id) ON DELETE CASCADE,
	FOREIGN KEY (asset_id) REFERENCES assets (id) ON DELETE CASCADE

);

--- Attachments information
CREATE TABLE attachments (
	id          TEXT PRIMARY KEY NOT NULL,
	course_id   TEXT NOT NULL,
	asset_id    TEXT NOT NULL,
	title       TEXT NOT NULL,
	path        TEXT UNIQUE NOT NULL,
	created_at  TEXT NOT NULL DEFAULT (STRFTIME('%Y-%m-%d %H:%M:%f', 'NOW')),
	updated_at  TEXT NOT NULL DEFAULT (STRFTIME('%Y-%m-%d %H:%M:%f', 'NOW')),
	---
	FOREIGN KEY (course_id) REFERENCES courses (id) ON DELETE CASCADE,
	FOREIGN KEY (asset_id) REFERENCES assets (id) ON DELETE CASCADE
);

--- Scans information
CREATE TABLE scans (
	id          TEXT PRIMARY KEY NOT NULL,
	course_id   TEXT UNIQUE NOT NULL,
	status      TEXT NOT NULL,
	created_at  TEXT NOT NULL DEFAULT (STRFTIME('%Y-%m-%d %H:%M:%f', 'NOW')),
	updated_at  TEXT NOT NULL DEFAULT (STRFTIME('%Y-%m-%d %H:%M:%f', 'NOW')),
	---
	FOREIGN KEY (course_id) REFERENCES courses (id) ON DELETE CASCADE
);