package models

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// Defines a model for the table `assets`
// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

import (
	"database/sql"

	"github.com/geerew/off-course/utils/types"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Asset defines the model for an asset
type Asset struct {
	BaseModel

	CourseID string
	Title    string
	Prefix   sql.NullInt16
	Chapter  string
	Type     types.Asset
	Path     string

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
