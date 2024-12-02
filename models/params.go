package models

import "github.com/geerew/off-course/utils/schema"

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Param defines the model for a parameter
type Param struct {
	Base
	Key   string
	Value string
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

var (
	PARAM_TABLE = "params"
	PARAM_KEY   = "key"
	PARAM_VALUE = "value"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Table implements the `schema.Modeler` interface by returning the table name
func (p *Param) Table() string {
	return PARAM_TABLE
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Fields implements the `schema.Modeler` interface by defining the model fields
func (p *Param) Define(s *schema.ModelConfig) {
	s.Embedded("Base")

	// Common fields
	s.Field("Key").Column(PARAM_KEY).NotNull()
	s.Field("Value").Column(PARAM_VALUE).NotNull().Mutable()
}
