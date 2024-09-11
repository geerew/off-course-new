package models

import (
	"github.com/geerew/off-course/utils/security"
	"github.com/geerew/off-course/utils/types"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

const (
	DB_TAG      = "db"
	DB_JOIN_TAG = "db_join"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// BaseModel defines the base model for all models
type BaseModel struct {
	ID        string         `db:"id:required"`
	CreatedAt types.DateTime `db:"created_at:required"`
	UpdatedAt types.DateTime `db:"updated_at:required"`
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

// RefreshCreatedAt updates the Created At field to the current date/time
func (b *BaseModel) RefreshCreatedAt() {
	b.CreatedAt = types.NowDateTime()
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// RefreshUpdatedAt updates the Updated At field to the current date/time
func (b *BaseModel) RefreshUpdatedAt() {
	b.UpdatedAt = types.NowDateTime()
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Scannable is an interface for a database row that can be scanned into a struct
type Scannable interface {
	Scan(dest ...interface{}) error
}
