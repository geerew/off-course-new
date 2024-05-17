package models

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

import "github.com/geerew/off-course/utils/types"

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// AssetProgress defines the model for a course progress (table: assets_progress)
type AssetProgress struct {
	BaseModel

	AssetID     string
	CourseID    string
	VideoPos    int
	Completed   bool
	CompletedAt types.DateTime
}
