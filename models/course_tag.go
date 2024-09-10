package models

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// CourseTag defines the model for a course tag  (table: courses_tags)
type CourseTag struct {
	BaseModel `db:":nested"`

	TagId    string `db:"tag_id:required"`
	CourseId string `db:"course_id:required"`

	// --------------------------------
	// Not in this table, but added via join
	// --------------------------------

	Course string
	Tag    string
}
