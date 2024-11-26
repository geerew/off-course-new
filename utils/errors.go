package utils

import "errors"

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

var (
	// Generic
	ErrNilPtr        = errors.New("nil pointer")
	ErrNotPtr        = errors.New("requires a pointer")
	ErrNotModeler    = errors.New("does not implement Modeler interface")
	ErrEmbedded      = errors.New("embedded struct does not implement Definer interface")
	ErrInvalidValue  = errors.New("invalid value")
	ErrInvalidColumn = errors.New("invalid column")
	ErrInvalidPluck  = errors.New("pluck is only valid when selecting a single column")
	ErrNotStruct     = errors.New("not a struct")
	ErrNotSlice      = errors.New("not a slice")
	ErrNoTable       = errors.New("table name cannot be empty")

	// DB
	ErrInvalidWhere = errors.New("where clause cannot be empty")

	// Model
	ErrInvalidId  = errors.New("id cannot be empty")
	ErrInvalidKey = errors.New("key cannot be empty")
)
