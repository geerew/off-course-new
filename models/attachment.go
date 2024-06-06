package models

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Attachment defines the model for an attachment (table: attachments)
type Attachment struct {
	BaseModel

	CourseID string
	AssetID  string
	Title    string
	Path     string
}
