package utils

import "errors"

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

var (
	ErrNilType         = errors.New("nil type")
	ErrUnsupportedType = errors.New("unsupported type")
	ErrInvalidValue    = errors.New("invalid value")
	ErrNilTableMapper  = errors.New("nil table mapper")

	// ErrInvalidId        = errors.New("id cannot be empty")
	// ErrInvalidTable     = errors.New("table name cannot be empty")
	// ErrInvalidCourseId  = errors.New("course id cannot be empty")
	// ErrInvalidWhere     = errors.New("where clause cannot be empty")
	// ErrInvalidPrefix    = errors.New("prefix must be greater than 0")
	// ErrNilTransaction   = errors.New("transaction cannot be nil")
	// ErrInvalidTag       = errors.New("tag cannot be empty")
	// ErrNilModel         = errors.New("model cannot be nil")
	// ErrInvalidSliceType = errors.New("input must be a slice or array")
	// ErrUnaddressable    = errors.New("value is unaddressable")

)
