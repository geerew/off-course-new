package models

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// Defines a model for the table `tags`
// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Tag defines the model for a tag
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
