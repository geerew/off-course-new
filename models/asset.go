package models

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

import (
	"database/sql"

	"github.com/geerew/off-course/utils/schema"
	"github.com/geerew/off-course/utils/types"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Asset defines the model for an asset
type Asset struct {
	Base
	CourseID string
	Title    string
	Prefix   sql.NullInt16
	Chapter  string
	Type     types.Asset
	Path     string
	Hash     string

	// Relations
	Progress    *AssetProgress
	Attachments []*Attachment
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

var (
	ASSET_TABLE          = "assets"
	ASSET_COURSE_ID      = "course_id"
	ASSET_TITLE          = "title"
	ASSET_PREFIX         = "prefix"
	ASSET_CHAPTER        = "chapter"
	ASSET_TYPE           = "type"
	ASSET_PATH           = "path"
	ASSET_HASH           = "hash"
	ASSET_VIDEO_POSITION = "video_pos"
	ASSET_COMPLETED      = "completed"
	ASSET_COMPLETED_AT   = "completed_at"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Table implements the `schema.Modeler` interface by returning the table name
func (a *Asset) Table() string {
	return ASSET_TABLE
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Fields implements the `schema.Modeler` interface by defining the model fields
func (a *Asset) Define(s *schema.ModelConfig) {
	s.Embedded("Base")

	// Common fields
	s.Field("CourseID").Column(ASSET_COURSE_ID).NotNull()
	s.Field("Title").Column(COURSE_TITLE).NotNull().Mutable()
	s.Field("Prefix").Column(ASSET_PREFIX).Mutable()
	s.Field("Chapter").Column(ASSET_CHAPTER).Mutable()
	s.Field("Type").Column(ASSET_TYPE).NotNull().Mutable()
	s.Field("Path").Column(ASSET_PATH).NotNull().Mutable()
	s.Field("Hash").Column(ASSET_HASH).NotNull().Mutable()

	// Relation fields
	s.Relation("Progress").MatchOn(ASSET_PROGRESS_ASSET_ID)
	s.Relation("Attachments").MatchOn(ATTACHMENT_ASSET_ID)
}
