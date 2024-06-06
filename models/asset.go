package models

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

import (
	"database/sql"
	"time"

	"github.com/geerew/off-course/utils/types"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Asset defines the model for an asset (table: assets)
type Asset struct {
	BaseModel

	CourseID string
	Title    string
	Prefix   sql.NullInt16
	Chapter  string
	Type     types.Asset
	Path     string
	Hash     string

	// --------------------------------
	// Not in this table, but added via a join
	// --------------------------------

	// Asset Progress
	VideoPos    int
	Completed   bool
	CompletedAt time.Time

	// Attachments
	Attachments []*Attachment
}
