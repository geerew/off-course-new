package models

import "github.com/geerew/off-course/utils/types"

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// User defines the model for a user (table: users)
type User struct {
	BaseModel

	Username     string
	PasswordHash string
	Role         types.UserRole
}
