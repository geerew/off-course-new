package models

import "time"

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// CourseProgress defines the model for a course progress (table: courses_progress)
type CourseProgress struct {
	BaseModel

	CourseID    string
	Started     bool
	StartedAt   time.Time
	Percent     int
	CompletedAt time.Time
}
