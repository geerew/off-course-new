package models

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Param defines the model for application parameters (table: params)
type Param struct {
	BaseModel
	Key   string
	Value string
}
