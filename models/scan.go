package models

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

import (
	"github.com/geerew/off-course/utils/types"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Scan defines the model for a scan  (table: scans)
type Scan struct {
	BaseModel

	CourseID string
	Status   types.ScanStatus

	// --------------------------------
	// Not in this table, but added via join
	// --------------------------------

	CoursePath string
}
