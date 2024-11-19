package models

import (
	"github.com/geerew/off-course/utils/schema"
	"github.com/geerew/off-course/utils/security"
	"github.com/geerew/off-course/utils/types"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

type Modeler interface {
	Table() string
	Id() string
	RefreshId()
	RefreshCreatedAt()
	RefreshUpdatedAt()
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Base defines the base model for all models
type Base struct {
	ID        string
	CreatedAt types.DateTime
	UpdatedAt types.DateTime
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

var (
	BASE_ID         = "id"
	BASE_CREATED_AT = "created_at"
	BASE_UPDATED_AT = "updated_at"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Define implements the `schema.Modeler` interface by defining the model
func (b *Base) Define(s *schema.ModelConfig) {
	// Common fields
	s.Field("ID").Column(BASE_ID).NotNull()
	s.Field("CreatedAt").Column(BASE_CREATED_AT).NotNull()
	s.Field("UpdatedAt").Column(BASE_UPDATED_AT).NotNull().Mutable()
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Id returns the model ID
func (b *Base) Id() string {
	return b.ID
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// RefreshId generates and sets a new model ID
func (b *Base) RefreshId() {
	b.ID = security.PseudorandomString(10)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// SetId sets the model ID
func (b *Base) SetId(id string) {
	b.ID = id
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// RefreshCreatedAt updates the Created At field to the current date/time
func (b *Base) RefreshCreatedAt() {
	b.CreatedAt = types.NowDateTime()
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// RefreshUpdatedAt updates the Updated At field to the current date/time
func (b *Base) RefreshUpdatedAt() {
	b.UpdatedAt = types.NowDateTime()
}
