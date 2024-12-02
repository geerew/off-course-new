package models

import "github.com/geerew/off-course/utils/schema"

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Attachment defines the model for an attachment
type Attachment struct {
	Base
	AssetID string
	Title   string
	Path    string
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

var (
	ATTACHMENT_TABLE    = "attachments"
	ATTACHMENT_ASSET_ID = "asset_id"
	ATTACHMENT_TITLE    = "title"
	ATTACHMENT_PATH     = "path"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Table implements the `schema.Modeler` interface by returning the table name
func (a *Attachment) Table() string {
	return ATTACHMENT_TABLE
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Fields implements the `schema.Modeler` interface by defining the model fields
func (a *Attachment) Define(s *schema.ModelConfig) {
	s.Embedded("Base")

	s.Field("AssetID").Column(ATTACHMENT_ASSET_ID).NotNull()
	s.Field("Title").Column(ATTACHMENT_TITLE).NotNull().Mutable()
	s.Field("Path").Column(ATTACHMENT_PATH).NotNull().Mutable()
}
