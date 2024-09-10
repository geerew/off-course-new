package models

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Param defines the model for application parameters (table: params)
type Param struct {
	BaseModel `db:":nested"`

	Key   string `db:"key:required"`
	Value string `db:"value:required"`
}
