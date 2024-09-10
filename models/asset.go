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
	// Not in this table, but added via a join
	// --------------------------------

	// Asset Progress
	VideoPos    int
	Completed   bool
	CompletedAt types.DateTime

	// Attachments
	Attachments []*Attachment
}
