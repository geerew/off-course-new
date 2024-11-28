package models

import (
	"github.com/geerew/off-course/utils/schema"
	"github.com/geerew/off-course/utils/types"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// AssetProgress defines the model for an asset progress
type AssetProgress struct {
	Base
	AssetID     string
	VideoPos    int
	Completed   bool
	CompletedAt types.DateTime
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

var (
	ASSET_PROGRESS_TABLE        = "assets_progress"
	ASSET_PROGRESS_ASSET_ID     = "asset_id"
	ASSET_PROGRESS_VIDEO_POS    = "video_pos"
	ASSET_PROGRESS_COMPLETED    = "completed"
	ASSET_PROGRESS_COMPLETED_AT = "completed_at"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Table implements the `schema.Modeler` interface by returning the table name
func (a *AssetProgress) Table() string {
	return ASSET_PROGRESS_TABLE
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Fields implements the `schema.Modeler` interface by defining the model fields
func (a *AssetProgress) Define(s *schema.ModelConfig) {
	s.Embedded("Base")

	// Common fields
	s.Field("AssetID").Column(ASSET_PROGRESS_ASSET_ID).NotNull()
	s.Field("VideoPos").Column(ASSET_PROGRESS_VIDEO_POS).Mutable()
	s.Field("Completed").Column(ASSET_PROGRESS_COMPLETED).Mutable()
	s.Field("CompletedAt").Column(ASSET_PROGRESS_COMPLETED_AT).Mutable()
}
