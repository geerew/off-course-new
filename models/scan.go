package models

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

import (
	"github.com/geerew/off-course/utils/types"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Scan defines the model for a scan  (table: scans)
type Scan struct {
	BaseModel `db:":nested"`

	CourseID string           `db:"course_id:required"`
	Status   types.ScanStatus `db:"status:required"`

	// --------------------------------
	// Added via join
	// --------------------------------

	CoursePath string `db_join:"courses:path:course_path"`
}
