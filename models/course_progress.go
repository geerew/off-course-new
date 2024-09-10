package models

import "github.com/geerew/off-course/utils/types"

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// CourseProgress defines the model for a course progress (table: courses_progress)
type CourseProgress struct {
	BaseModel `db:":nested"`

	CourseID    string         `db:"course_id:required"`
	Started     bool           `db:"started"`
	StartedAt   types.DateTime `db:"started_at"`
	Percent     int            `db:"percent"`
	CompletedAt types.DateTime `db:"completed_at"`
}
