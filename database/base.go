package database

import (
	"database/sql"

	"github.com/Masterminds/squirrel"
	"github.com/geerew/off-course/utils/appFs"
	"github.com/geerew/off-course/utils/pagination"
	_ "modernc.org/sqlite"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Defines the sql functions
type (
	ExecFn     = func(query string, args ...interface{}) (sql.Result, error)
	QueryFn    = func(query string, args ...interface{}) (*sql.Rows, error)
	QueryRowFn = func(query string, args ...interface{}) *sql.Row
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Database defines the interface for a database
type Database interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
	Begin(opts *sql.TxOptions) (*sql.Tx, error)
	RunInTransaction(txFunc func(*sql.Tx) error) error
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// DatabaseParams defines optional params for a database query
type DatabaseParams struct {
	// A slice of columns to order by (ex ["id DESC", "title ASC"])
	OrderBy []string

	// A slice of columns to select (ex ["id", "title", "courses.col"])
	Columns []string

	// Any valid squirrel WHERE expression
	//
	//
	// Examples:
	//
	//   EQ:   sq.Eq{"id": "123"}
	//   IN:   sq.Eq{"id": []string{"123", "456"}}
	//   OR:   sq.Or{sq.Expr("id = ?", "123"), sq.Expr("id = ?", "456")}
	//   AND:  sq.And{sq.Eq{"id": "123"}, sq.Eq{"title": "devops"}}
	//   LIKE: sq.Like{"title": "%dev%"}
	//   NOT:  sq.NotEq{"id": "123"}
	Where any

	// Columns to group by
	GroupBys []string

	// Limit the results
	Having squirrel.Sqlizer

	// Used to paginate the results
	Pagination *pagination.Pagination

	// Control which related entities to fetch
	IncludeRelations []string

	// Whether to use case-insensitive search
	CaseInsensitive bool
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// DatabaseManager manages the database connections
type DatabaseManager struct {
	DataDb Database
	LogsDb Database
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// DatabaseConfig defines the configuration for a database
type DatabaseConfig struct {
	IsDebug    bool
	DataDir    string
	DSN        string
	MigrateDir string
	AppFs      *appFs.AppFs
	InMemory   bool
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// NewSqliteDBManager returns a new DatabaseManager
func NewSqliteDBManager(config *DatabaseConfig) (*DatabaseManager, error) {
	manager := &DatabaseManager{}

	dataConfig := &DatabaseConfig{
		IsDebug:    config.IsDebug,
		DataDir:    config.DataDir,
		DSN:        "data.db",
		MigrateDir: "data",
		AppFs:      config.AppFs,
		InMemory:   config.InMemory,
	}

	if dataDb, err := NewSqliteDB(dataConfig); err != nil {
		return nil, err
	} else {
		manager.DataDb = dataDb
	}

	logsConfig := &DatabaseConfig{
		IsDebug:    config.IsDebug,
		DataDir:    config.DataDir,
		DSN:        "logs.db",
		MigrateDir: "logs",
		AppFs:      config.AppFs,
		InMemory:   config.InMemory,
	}

	if logsDB, err := NewSqliteDB(logsConfig); err != nil {
		return nil, err
	} else {
		manager.LogsDb = logsDB
	}

	return manager, nil
}
