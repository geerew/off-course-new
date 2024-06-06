package models

import "time"

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// AssetProgress defines the model for a course progress (table: assets_progress)
type AssetProgress struct {
	BaseModel

	AssetID     string
	CourseID    string
	VideoPos    int
	Completed   bool
	CompletedAt time.Time
}
