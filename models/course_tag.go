package models

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// CourseTag defines the model for a course tag  (table: courses_tags)
type CourseTag struct {
	BaseModel

	TagId    string
	CourseId string

	// --------------------------------
	// Not in this table, but added via join
	// --------------------------------

	Course string
	Tag    string
}
