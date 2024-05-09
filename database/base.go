package database

import (
	"database/sql"
	"os"
	"path/filepath"

	"github.com/Masterminds/squirrel"
	"github.com/geerew/off-course/migrations"
	"github.com/geerew/off-course/utils/appFs"
	"github.com/geerew/off-course/utils/pagination"
	"github.com/pressly/goose/v3"
	"github.com/rs/zerolog/log"
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
	Bootstrap() error
	Exec(query string, args ...interface{}) (sql.Result, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
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

// SqliteDb defines an sqlite storage
type SqliteDb struct {
	db       *sql.DB
	isDebug  bool
	dataDir  string
	appFs    *appFs.AppFs
	inMemory bool
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// SqliteDbConfig defines the config when creating a new sqlite storage
type SqliteDbConfig struct {
	IsDebug  bool
	DataDir  string
	AppFs    *appFs.AppFs
	InMemory bool
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// NewSqlite creates a new SqliteDb
func NewSqliteDB(config *SqliteDbConfig) *SqliteDb {
	return &SqliteDb{
		isDebug:  config.IsDebug,
		dataDir:  config.DataDir,
		appFs:    config.AppFs,
		inMemory: config.InMemory,
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Bootstrap initializes an sqlite DB connection and migrates the models, if required
func (s *SqliteDb) Bootstrap() error {
	// Create the data dir (if it does not exist)
	if err := s.appFs.Fs.MkdirAll(s.dataDir, os.ModePerm); err != nil {
		return err
	}

	dsn := filepath.Join(s.dataDir, "data.db")
	if s.inMemory {
		dsn = "file::memory:"
	}

	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return err
	}

	// TODO: make this better (use semaphore to block/continue)
	db.SetMaxIdleConns(1)
	db.SetMaxOpenConns(1)

	// Setup the default DB connection
	//
	// Note: busy_timeout needs to be set BEFORE journal_mode
	_, err = db.Exec(`
		PRAGMA busy_timeout       = 10000;
		PRAGMA journal_mode       = WAL;
		PRAGMA journal_size_limit = 200000000;
		PRAGMA synchronous        = NORMAL;
		PRAGMA foreign_keys       = ON;
		PRAGMA cache_size         = -16000;
	`)

	if err != nil {
		return err
	}

	s.db = db

	// Do the migrate
	if err := s.migrate(); err != nil {
		return err
	}

	return nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// logQuery logs the SQL command if debug mode is enabled
func (s *SqliteDb) logQuery(query string, args ...interface{}) {
	if s.isDebug {
		log.Debug().Msgf("SQL Query: %s; Arguments: %v", query, args)
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Exec executes a query without returning any rows
func (s *SqliteDb) Exec(query string, args ...interface{}) (sql.Result, error) {
	s.logQuery(query, args...)
	return s.db.Exec(query, args...)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Query executes a query that returns rows, typically a SELECT
func (s *SqliteDb) Query(query string, args ...interface{}) (*sql.Rows, error) {
	s.logQuery(query, args...)
	return s.db.Query(query, args...)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// QueryRow executes a query that is expected to return at most one row
func (s *SqliteDb) QueryRow(query string, args ...interface{}) *sql.Row {
	s.logQuery(query, args...)
	return s.db.QueryRow(query, args...)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (s *SqliteDb) RunInTransaction(txFunc func(*sql.Tx) error) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}

	if err := txFunc(tx); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Migrate runs the DB migrations
func (s *SqliteDb) migrate() error {
	// Disable goose logging
	//
	// TODO: Handle this better
	goose.SetLogger(goose.NopLogger())

	goose.SetBaseFS(migrations.EmbedMigrations)

	if err := goose.SetDialect("sqlite3"); err != nil {
		return err
	}

	if err := goose.Up(s.db, "."); err != nil {
		return err
	}

	return nil
}
