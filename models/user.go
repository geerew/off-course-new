package models

import (
	"github.com/geerew/off-course/utils/schema"
	"github.com/geerew/off-course/utils/types"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Course defines the model for a course
type User struct {
	Base `db:":nested"`

	Username     string         `db:"col:username;not null"`
	PasswordHash string         `db:"col:password_hash;not null"`
	Role         types.UserRole `db:"col:role;not null"`
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

var (
	USER_TABLE         = "users"
	USER_USERNAME      = "username"
	USER_PASSWORD_HASH = "password_hash"
	USER_ROLE          = "role"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Table implements the `schema.Modeler` interface by returning the table name
func (u *User) Table() string {
	return USER_TABLE
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Fields implements the `schema.Modeler` interface by defining the model fields
func (u *User) Define(s *schema.ModelConfig) {
	s.Embedded("Base")

	// Common fields
	s.Field("Username").Column(USER_USERNAME).NotNull()
	s.Field("PasswordHash").Column(USER_PASSWORD_HASH).NotNull()
	s.Field("Role").Column(USER_ROLE).NotNull()
}
