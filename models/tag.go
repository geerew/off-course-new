package models

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Tag defines the model for a tag (table: tags)
type Tag struct {
	BaseModel

	Tag string

	// --------------------------------
	// Not in this table, but added via a join
	// --------------------------------

	// Courses
	CourseCount int
	CourseTags  []*CourseTag
}
