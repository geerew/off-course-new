package models

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Attachment defines the model for an attachment (table: attachments)
type Attachment struct {
	BaseModel `db:":nested"`

	CourseID string `db:"course_id:required"`
	AssetID  string `db:"asset_id:required"`
	Title    string `db:"title:required"`
	Path     string `db:"path:required"`
}
