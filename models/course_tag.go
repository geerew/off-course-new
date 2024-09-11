package models

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// CourseTag defines the model for a course tag  (table: courses_tags)
type CourseTag struct {
	BaseModel `db:":nested"`

	TagId    string `db:"tag_id:required"`
	CourseId string `db:"course_id:required"`

	// --------------------------------
	// Added via join
	// --------------------------------

	Course string `db_join:"courses:title"`
	Tag    string `db_join:"tags:tag"`
}
