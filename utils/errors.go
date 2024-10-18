package utils

import "errors"

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

var (
	ErrNilPtr       = errors.New("nil pointer")
	ErrNotPtr       = errors.New("requires a pointer")
	ErrNotModeler   = errors.New("does not implement Modeler interface")
	ErrEmbedded     = errors.New("embedded struct does not implement Fielder interface")
	ErrInvalidValue = errors.New("invalid value")

	// Model errors
	ErrInvalidId       = errors.New("id cannot be empty")
	ErrInvalidTag      = errors.New("tag cannot be empty")
	ErrInvalidCourseId = errors.New("course id cannot be empty")
)
