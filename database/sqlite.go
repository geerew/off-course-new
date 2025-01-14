package database

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/geerew/off-course/migrations"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pressly/goose/v3"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// NewSqliteDBManager returns a new DatabaseManager
func NewSqliteDBManager(config *DatabaseConfig) (*DatabaseManager, error) {
	manager := &DatabaseManager{}

	dataConfig := &DatabaseConfig{
		DataDir:    config.DataDir,
		DSN:        "data.db",
		MigrateDir: "data",
		AppFs:      config.AppFs,
		InMemory:   config.InMemory,
		Logger:     config.Logger,
	}

	if dataDb, err := NewSqliteDB(dataConfig); err != nil {
		return nil, err
	} else {
		manager.DataDb = dataDb
	}

	logsConfig := &DatabaseConfig{
		DataDir:    config.DataDir,
		DSN:        "logs.db",
		MigrateDir: "logs",
		AppFs:      config.AppFs,
		InMemory:   config.InMemory,

		// Never provider a logger for the logs DB as it will cause an infinite loop
		Logger: nil,
	}

	if logsDB, err := NewSqliteDB(logsConfig); err != nil {
		return nil, err
	} else {
		manager.LogsDb = logsDB
	}

	return manager, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Tx is a wrapper around sql.Tx that logs queries
type Tx struct {
	*sql.Tx
	db *SqliteDb
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Exec executes a query within a transaction without returning any rows
func (tx *Tx) Exec(query string, args ...any) (sql.Result, error) {
	tx.db.log(query, args...)
	return tx.Tx.Exec(query, args...)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Query executes a query within a transaction that returns rows, typically a SELECT statement
func (tx *Tx) Query(query string, args ...any) (*sql.Rows, error) {
	tx.db.log(query, args...)
	return tx.Tx.Query(query, args...)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// QueryRow executes a query within a transaction that is expected to return at most one row
func (tx *Tx) QueryRow(query string, args ...any) *sql.Row {
	tx.db.log(query, args...)
	return tx.Tx.QueryRow(query, args...)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// SqliteDb defines an sqlite database
type SqliteDb struct {
	conn   *sql.DB
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

// DB returns the underlying sql.DB
func (db *SqliteDb) DB() *sql.DB {
	return db.conn
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Query executes a query that returns rows, typically a SELECT statement
//
// It implements the Database interface
func (db *SqliteDb) Query(query string, args ...any) (*sql.Rows, error) {
	db.log(query, args...)
	return db.conn.Query(query, args...)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// QueryRow executes a query that is expected to return at most one row
//
// It implements the Database interface
func (db *SqliteDb) QueryRow(query string, args ...any) *sql.Row {
	db.log(query, args...)
	return db.conn.QueryRow(query, args...)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Exec executes a query without returning any rows
//
// It implements the Database interface
func (db *SqliteDb) Exec(query string, args ...any) (sql.Result, error) {
	db.log(query, args...)
	return db.conn.Exec(query, args...)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// RunInTransaction runs a function in a transaction
//
// It implements the Database interface
func (db *SqliteDb) RunInTransaction(ctx context.Context, txFunc func(context.Context) error) (err error) {
	// Check if there's an existing querier in the context
	existingQuerier := QuerierFromContext(ctx, nil)
	if existingQuerier != nil {
		return txFunc(ctx)
	}

	slqTx, err := db.conn.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	tx := &Tx{
		Tx: slqTx,
		db: db,
	}

	// Set the querier in the context to use the transaction
	txCtx := WithQuerier(ctx, tx)

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

	return txFunc(txCtx)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (db *SqliteDb) SetLogger(l *slog.Logger) {
	db.config.Logger = l
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// bootstrap initializes the sqlite database connect and sets db.conn
func (db *SqliteDb) bootstrap() error {
	if err := db.config.AppFs.Fs.MkdirAll(db.config.DataDir, os.ModePerm); err != nil {
		return err
	}

	dsn := filepath.Join(db.config.DataDir, db.config.DSN)
	if db.config.InMemory {
		dsn = "file::memory:"
	}

	conn, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return err
	}

	// TODO: make this better (use semaphore to block/continue)
	conn.SetMaxIdleConns(1)
	conn.SetMaxOpenConns(1)

	db.conn = conn

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

	if err := goose.Up(db.conn, db.config.MigrateDir); err != nil {
		return err
	}

	return nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (db *SqliteDb) log(query string, args ...any) {
	if db.config.Logger != nil {
		attrs := make([]any, 0, len(args))
		attrs = append(attrs, loggerType)

		for i, arg := range args {
			attrs = append(attrs, slog.Any(fmt.Sprintf("arg %d", i+1), arg))
		}

		db.config.Logger.Debug(
			query,
			attrs...,
		)
	}

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
