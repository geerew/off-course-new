package models

import (
	"github.com/geerew/off-course/utils/schema"
	"github.com/geerew/off-course/utils/types"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Log defines the model for a log
type Log struct {
	Base
	Level   int
	Message string
	Data    types.JsonMap
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

var (
	LOG_TABLE   = "logs"
	LOG_LEVEL   = "level"
	LOG_MESSAGE = "message"
	LOG_DATA    = "data"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Table implements the `schema.Modeler` interface by returning the table name
func (l *Log) Table() string {
	return LOG_TABLE
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Fields implements the `schema.Modeler` interface by defining the model fields
func (l *Log) Define(s *schema.ModelConfig) {
	s.Embedded("Base")

	// Common fields
	s.Field("Level").Column(LOG_LEVEL)
	s.Field("Message").Column(LOG_MESSAGE).NotNull()
	s.Field("Data").Column(LOG_DATA).NotNull()
}
