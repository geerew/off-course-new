package models

import (
	"fmt"

	"github.com/geerew/off-course/utils/schema"
	"github.com/geerew/off-course/utils/types"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Course defines the model for a course
type Course struct {
	Base
	Title     string
	Path      string
	CardPath  string
	Available bool

	// Joins
	ScanStatus types.ScanStatus

	// Relations
	Progress CourseProgress
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

var (
	COURSE_TABLE       = "courses"
	COURSE_TITLE       = "title"
	COURSE_PATH        = "path"
	COURSE_CARD_PATH   = "card_path"
	COURSE_AVAILABLE   = "available"
	COURSE_SCAN_STATUS = "status"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Table implements the `schema.Modeler` interface by returning the table name
func (c *Course) Table() string {
	return COURSE_TABLE
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Fields implements the `schema.Modeler` interface by defining the model fields
func (c *Course) Define(s *schema.ModelConfig) {
	s.Embedded("Base")

	// Common fields
	s.Field("Title").Column(COURSE_TITLE).NotNull()
	s.Field("Path").Column(COURSE_PATH).NotNull()
	s.Field("CardPath").Column(COURSE_CARD_PATH).Mutable()
	s.Field("Available").Column(COURSE_AVAILABLE).Mutable()

	// Join fields
	s.Field("ScanStatus").JoinTable(SCAN_TABLE).Column(COURSE_SCAN_STATUS).Alias("scan_status")

	// Relation fields
	s.Relation("Progress").MatchOn(COURSE_PROGRESS_COURSE_ID)

	// Joins
	s.LeftJoin(SCAN_TABLE).On(fmt.Sprintf("%s.%s = %s.%s", COURSE_TABLE, BASE_ID, SCAN_TABLE, SCAN_COURSE_ID))
}
