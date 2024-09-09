package models

import "time"

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Course defines the model for a course  (table: courses)
type Course struct {
	BaseModel
	Title     string
	Path      string
	CardPath  string
	Available bool

	// --------------------------------
	// Not in this table, but added via join
	// --------------------------------

	// Scan status
	ScanStatus string

	// Course Progress
	Started           bool
	StartedAt         time.Time
	Percent           int
	CompletedAt       time.Time
	ProgressUpdatedAt time.Time
}
