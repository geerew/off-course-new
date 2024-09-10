package models

import "github.com/geerew/off-course/utils/types"

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// AssetProgress defines the model for a course progress (table: assets_progress)
type AssetProgress struct {
	BaseModel `db:":nested"`

	AssetID     string         `db:"asset_id:required"`
	CourseID    string         `db:"course_id:required"`
	VideoPos    int            `db:"video_pos"`
	Completed   bool           `db:"completed"`
	CompletedAt types.DateTime `db:"completed_at"`
}
