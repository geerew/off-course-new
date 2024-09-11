package models

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

import (
	"database/sql"

	"github.com/geerew/off-course/utils/types"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Asset defines the model for an asset (table: assets)
type Asset struct {
	BaseModel `db:":nested"`

	CourseID string        `db:"course_id:required"`
	Title    string        `db:"title:required"`
	Prefix   sql.NullInt16 `db:"prefix"`
	Chapter  string        `db:"chapter:required"`
	Type     types.Asset   `db:"type:required"`
	Path     string        `db:"path:required"`
	Hash     string        `db:"hash:required"`

	// --------------------------------
	// Added via a join
	// --------------------------------

	// Asset Progress
	VideoPos    int            `db_join:"assets_progress:video_pos"`
	Completed   bool           `db_join:"assets_progress:completed"`
	CompletedAt types.DateTime `db_join:"assets_progress:completed_at"`

	// --------------------------------
	// Manually added
	// --------------------------------
	Attachments []*Attachment
}
