package models

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

import "github.com/geerew/off-course/utils/types"

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// CourseProgress defines the model for a course progress (table: courses_progress)
type CourseProgress struct {
	BaseModel

	CourseID    string
	Started     bool
	StartedAt   types.DateTime
	Percent     int
	CompletedAt types.DateTime
}
