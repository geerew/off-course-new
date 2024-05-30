package database

import (
	"database/sql"
	"os"
	"path/filepath"

	"github.com/geerew/off-course/migrations"
	"github.com/pressly/goose/v3"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// SqliteDb defines an sqlite database
type SqliteDb struct {
	DB     *sql.DB
	config *DatabaseConfig
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// NewSqliteDB creates a new SqliteDb
func NewSqliteDB(config *DatabaseConfig) (*SqliteDb, error) {
	sqliteDB := &SqliteDb{
		config: config,
	}

	if err := sqliteDB.bootstrap(); err != nil {
		return nil, err
	}

	if err := sqliteDB.migrate(); err != nil {
		return nil, err
	}

	return sqliteDB, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Query executes a query that returns rows, typically a SELECT statement
//
// It implements the Database interface
func (db *SqliteDb) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return db.DB.Query(query, args...)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// QueryRow executes a query that is expected to return at most one row
//
// It implements the Database interface
func (db *SqliteDb) QueryRow(query string, args ...interface{}) *sql.Row {
	return db.DB.QueryRow(query, args...)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Exec executes a query without returning any rows
//
// It implements the Database interface
func (db *SqliteDb) Exec(query string, args ...interface{}) (sql.Result, error) {
	return db.DB.Exec(query, args...)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Begin starts a new transaction
//
// It implements the Database interface
func (db *SqliteDb) Begin(opts *sql.TxOptions) (*sql.Tx, error) {
	return db.DB.Begin()
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// RunInTransaction runs a function in a transaction
//
// It implements the Database interface
func (db *SqliteDb) RunInTransaction(txFunc func(*sql.Tx) error) error {
	tx, err := db.DB.Begin()
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if err != nil {
			tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	err = txFunc(tx)
	return err
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// bootstrap initializes the sqlite database connect and sets db.DB
func (db *SqliteDb) bootstrap() error {
	if err := db.config.AppFs.Fs.MkdirAll(db.config.DataDir, os.ModePerm); err != nil {
		return err
	}

	dsn := filepath.Join(db.config.DataDir, db.config.DSN)
	if db.config.InMemory {
		dsn = "file::memory:"
	}

	conn, err := sql.Open("sqlite", dsn)
	if err != nil {
		return err
	}

	// TODO: make this better (use semaphore to block/continue)
	conn.SetMaxIdleConns(1)
	conn.SetMaxOpenConns(1)

	db.DB = conn

	if err := db.setPragma(); err != nil {
		return err
	}

	return nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// migrate runs the goose migrations
func (db *SqliteDb) migrate() error {
	goose.SetLogger(goose.NopLogger())

	goose.SetBaseFS(migrations.EmbedMigrations)

	if err := goose.SetDialect("sqlite3"); err != nil {
		return err
	}

	if err := goose.Up(db.DB, db.config.MigrateDir); err != nil {
		return err
	}

	return nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// setPragma sets the default PRAGMA values for the DB
func (db *SqliteDb) setPragma() error {
	// Note: busy_timeout needs to be set BEFORE journal_mode
	_, err := db.Exec(`
	PRAGMA busy_timeout       = 10000;
	PRAGMA journal_mode       = WAL;
	PRAGMA journal_size_limit = 200000000;
	PRAGMA synchronous        = NORMAL;
	PRAGMA foreign_keys       = ON;
	PRAGMA cache_size         = -16000;
`)

	return err
}
