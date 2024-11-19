package models

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

import (
	"fmt"

	"github.com/geerew/off-course/utils/schema"
	"github.com/geerew/off-course/utils/types"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Scan defines the model for a scan
type Scan struct {
	Base
	CourseID string
	Status   types.ScanStatus

	// Joins
	CoursePath string
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

var (
	SCAN_TABLE       = "scans"
	SCAN_COURSE_ID   = "course_id"
	SCAN_STATUS      = "status"
	SCAN_COURSE_PATH = "path"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Table implements the `schema.Modeler` interface by returning the table name
func (s *Scan) Table() string {
	return SCAN_TABLE
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Define implements the `schema.Modeler` interface by defining the model
func (s *Scan) Define(c *schema.ModelConfig) {
	c.Embedded("Base")

	// Common fields
	c.Field("CourseID").Column(SCAN_COURSE_ID)
	c.Field("Status").Column(SCAN_STATUS).Mutable().IgnoreIfNull()

	// Join fields
	c.Field("CoursePath").JoinTable(COURSE_TABLE).Column(SCAN_COURSE_PATH).Alias("course_path")

	// Joins
	c.LeftJoin(COURSE_TABLE).On(fmt.Sprintf("%s.%s = %s.%s", SCAN_TABLE, SCAN_COURSE_ID, COURSE_TABLE, BASE_ID))
}
