package models

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// Defines a model for the table `attachments`
// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Attachment defines the model for an attachment
type Attachment struct {
	BaseModel

	CourseID string
	AssetID  string
	Title    string
	Path     string
	Md5      string
}
