package migrations

import (
	"context"

	"github.com/rs/zerolog/log"
	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		log.Debug().Str("file", "init").Msg("up migration")

		_, err := db.NewRaw(
			`CREATE TABLE courses (
				id          TEXT PRIMARY KEY NOT NULL,
				title       TEXT NOT NULL,
				path        TEXT UNIQUE NOT NULL,
				card_path   TEXT,
				started     BOOLEAN NOT NULL DEFAULT FALSE,
				finished    BOOLEAN NOT NULL DEFAULT FALSE,
				scan_status TEXT,
				created_at  TEXT NOT NULL DEFAULT (STRFTIME('%Y-%m-%d %H:%M:%f', 'NOW')),
				updated_at  TEXT NOT NULL DEFAULT (STRFTIME('%Y-%m-%d %H:%M:%f', 'NOW'))
			);

			CREATE TABLE assets (
				id          TEXT PRIMARY KEY NOT NULL,
				course_id   TEXT NOT NULL,
				title       TEXT NOT NULL,
				prefix      INTEGER NOT NULL,
				chapter     TEXT,
				type        TEXT NOT NULL,
				path        TEXT UNIQUE NOT NULL,
				progress    INTEGER DEFAULT 0,
				finished    BOOLEAN NOT NULL DEFAULT FALSE,
				created_at  TEXT NOT NULL DEFAULT (STRFTIME('%Y-%m-%d %H:%M:%f', 'NOW')),
				updated_at  TEXT NOT NULL DEFAULT (STRFTIME('%Y-%m-%d %H:%M:%f', 'NOW')),
				---
				FOREIGN KEY (course_id) REFERENCES courses (id) ON DELETE CASCADE
			);

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


			CREATE TABLE scans (
				id          TEXT PRIMARY KEY NOT NULL,
				course_id   TEXT UNIQUE NOT NULL,
				status      TEXT NOT NULL,
				created_at  TEXT NOT NULL DEFAULT (STRFTIME('%Y-%m-%d %H:%M:%f', 'NOW')),
				updated_at  TEXT NOT NULL DEFAULT (STRFTIME('%Y-%m-%d %H:%M:%f', 'NOW')),
				---
				FOREIGN KEY (course_id) REFERENCES courses (id) ON DELETE CASCADE
			);
		`).Exec(context.Background())

		return err

	}, nil)
}
