package models

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Log defines the model for a log (table: logs)
type Log struct {
	BaseModel

	Level   int
	Message string
}
