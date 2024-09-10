package models

import "github.com/geerew/off-course/utils/types"

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// User defines the model for a user (table: users)
type User struct {
	BaseModel `db:":nested"`

	Username     string         `db:"username:required"`
	PasswordHash string         `db:"password_hash:required"`
	Role         types.UserRole `db:"role:required"`
}
