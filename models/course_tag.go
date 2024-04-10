package models

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// Defines a model for the table `courses_tags`
// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// CourseTag defines the model for a course tag
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
