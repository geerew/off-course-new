package models

import "github.com/geerew/off-course/utils/types"

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Log defines the model for a log (table: logs)
type Log struct {
	BaseModel

	Data    types.JsonMap
	Level   int
	Message string
}
