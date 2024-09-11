package daos

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

import (
	"errors"

	"github.com/Masterminds/squirrel"
	"github.com/geerew/off-course/database"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Errors
var (
	ErrEmptyId         = errors.New("id cannot be empty")
	ErrMissingCourseId = errors.New("course id cannot be empty")
	ErrMissingWhere    = errors.New("where clause cannot be empty")
	ErrInvalidPrefix   = errors.New("prefix must be greater than 0")
	ErrNilTransaction  = errors.New("transaction cannot be nil")
	ErrMissingTag      = errors.New("tag cannot be empty")
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Scannable is an interface for a database row that can be scanned into a struct
type Scannable interface {
	Scan(dest ...interface{}) error
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// ScanFn is a type representing a generic function that scans a Scannable into
// a struct
type ScanFn[T any] func(Scannable) (*T, error)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// BaseDao
// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// BaseDao is the base data access object that all other daos inherit from
type BaseDao struct {
	db    database.Database
	table string
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Db returns the database
func (dao *BaseDao) Db() database.Database {
	return dao.db
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Table returns the table name
func (dao *BaseDao) Table() string {
	return dao.table
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Count satisfies the daoer interface for Daos that do not require a count method
func (dao *BaseDao) Count(dbParams *database.DatabaseParams, tx *database.Tx) (int, error) {
	return 0, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// countSelect returns the default count select builder
//
// Note: The columns are removed, so you must specify the columns with `.Columns(...)` when using
// this select builder
func (dao *BaseDao) countSelect() squirrel.SelectBuilder {
	return squirrel.
		StatementBuilder.
		PlaceholderFormat(squirrel.Question).
		Select("").
		From(dao.Table()).
		RemoveColumns()
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// baseSelect returns the default select builder
//
// Note: The columns are removed, so you must specify the columns with `.Columns(...)` when using
// this select builder
func (dao *BaseDao) baseSelect() squirrel.SelectBuilder {
	return dao.countSelect()
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// daoer is an interface for a data access object to be able to use the generic
// functions
type daoer interface {
	Db() database.Database
	Table() string
	Count(*database.DatabaseParams, *database.Tx) (int, error)
	countSelect() squirrel.SelectBuilder
	baseSelect() squirrel.SelectBuilder
}
