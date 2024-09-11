package models

import "github.com/geerew/off-course/utils/types"

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Course defines the model for a course  (table: courses)
type Course struct {
	BaseModel `db:":nested"`

	Title     string `db:"title:required"`
	Path      string `db:"path:required"`
	CardPath  string `db:"card_path"`
	Available bool   `db:"available"`

	// --------------------------------
	// Added via join
	// --------------------------------

	// Scan status
	ScanStatus string `db_join:"scans:status:scan_status"`

	// Course Progress
	Started           bool           `db_join:"courses_progress:started"`
	StartedAt         types.DateTime `db_join:"courses_progress:started_at"`
	Percent           int            `db_join:"courses_progress:percent"`
	CompletedAt       types.DateTime `db_join:"courses_progress:completed_at"`
	ProgressUpdatedAt types.DateTime `db_join:"courses_progress:updated_at:progress_updated_at"`
}
