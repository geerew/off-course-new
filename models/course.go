package models

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// Defines a model for the table `courses`
// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

import (
	"github.com/geerew/off-course/utils/types"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Course defines the model for a course
type Course struct {
	BaseModel
	Title     string
	Path      string
	CardPath  string
	Available bool

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
