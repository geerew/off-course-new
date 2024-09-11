package models

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Tag defines the model for a tag (table: tags)
type Tag struct {
	BaseModel `db:":nested"`

	Tag string `db:"tag:required"`

	// --------------------------------
	// Added via a join
	// --------------------------------

	CourseCount int `db_join:"courses_tags:id:course_count:COALESCE(COUNT(courses_tags.id), 0)"`

	// --------------------------------
	// Manually added
	// --------------------------------
	CourseTags []*CourseTag
}
