package models

import "github.com/geerew/off-course/utils/types"

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Log defines the model for a log (table: logs)
type Log struct {
	BaseModel `db:":nested"`

	Data    types.JsonMap `db:"data"`
	Level   int           `db:"level"`
	Message string        `db:"message:required"`
}
