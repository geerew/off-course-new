package models

import "github.com/geerew/off-course/utils/types"

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Course defines the model for a course  (table: courses)
type Course struct {
	BaseModel `db:":nested"`

	Title     string `db:"title:required"`
	Path      string `db:"path:required"`
	CardPath  string `db:"card_path"`
	Available bool   `db:"available"`

	// --------------------------------
	// Not in this table, but added via join
	// --------------------------------

	// Scan status
	ScanStatus string

	// Course Progress
	Started           bool
	StartedAt         types.DateTime
	Percent           int
	CompletedAt       types.DateTime
	ProgressUpdatedAt types.DateTime
}
