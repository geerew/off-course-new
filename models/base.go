package models

import (
	"github.com/geerew/off-course/utils/security"
	"github.com/geerew/off-course/utils/types"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// BaseModel defines the base model for all models
type BaseModel struct {
	ID        string
	CreatedAt types.DateTime
	UpdatedAt types.DateTime
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// RefreshId generates and sets a new model ID
func (b *BaseModel) RefreshId() {
	b.ID = security.PseudorandomString(10)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// SetId sets the model ID
func (b *BaseModel) SetId(id string) {
	b.ID = id
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// RefreshCreated updates the Created At field to the current date/time
func (b *BaseModel) RefreshCreatedAt() {
	b.CreatedAt = types.NowDateTime()
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// RefreshUpdatedAt updates the Updated At field to the current date/time
func (b *BaseModel) RefreshUpdatedAt() {
	b.UpdatedAt = types.NowDateTime()
}
