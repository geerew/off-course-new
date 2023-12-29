package database

import (
	"database/sql"
	"os"
	"path/filepath"
	"runtime"

	"github.com/geerew/off-course/utils/appFs"
	"github.com/geerew/off-course/utils/pagination"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/sqlitedialect"
	"github.com/uptrace/bun/driver/sqliteshim"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Database defines the interface for a database
type Database interface {
	Bootstrap() error
	DB() *bun.DB
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

type Where struct {
	Query  string
	Column string
	Value  any
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

type Relation struct {
	Struct  string
	Cols    []string
	OrderBy []string
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// DatabaseParams defines optional params for a database query
type DatabaseParams struct {
	OrderBy    []string
	Where      []Where
	Relation   []Relation
	Pagination *pagination.Pagination
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// SqliteDb defines an sqlite storage
type SqliteDb struct {
	db      *bun.DB
	isDebug bool
	dataDir string
	appFs   *appFs.AppFs
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// SqliteDbConfig defines the config when creating a new sqlite storage
type SqliteDbConfig struct {
	IsDebug bool
	DataDir string
	AppFs   *appFs.AppFs
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// NewSqlite creates a new SqliteDb
func NewSqliteDB(config *SqliteDbConfig) *SqliteDb {
	return &SqliteDb{
		isDebug: config.IsDebug,
		dataDir: config.DataDir,
		appFs:   config.AppFs,
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Bootstrap initializes an sqlite DB connection and migrates the models, if required
func (s *SqliteDb) Bootstrap() error {
	// Ensure that data dir exist
	if err := s.appFs.Fs.MkdirAll(s.dataDir, os.ModePerm); err != nil {
		return err
	}

	dsn := filepath.Join(s.dataDir, "data.db")
	if os.Getenv("OC_InMemDb") != "" {
		dsn = "file::memory:"
	}

	sqldb, err := sql.Open(sqliteshim.ShimName, dsn)
	if err != nil {
		panic(err)
	}

	db := bun.NewDB(sqldb, sqlitedialect.New())

	// Set the max open connections to 4x the number of CPUs
	maxOpenConns := 4 * runtime.GOMAXPROCS(0)
	db.SetMaxIdleConns(maxOpenConns)
	db.SetMaxOpenConns(maxOpenConns)

	// db.AddQueryHook(bundebug.NewQueryHook(bundebug.WithVerbose(true)))
	// db.AddQueryHook(hook)

	// Define a logger when debug is enabled
	// config := &gorm.Config{}
	// if s.isDebug {
	// 	config.Logger = NewZerologGormLogger(log.Logger)
	// } else {
	// 	config.Logger = NewZerologGormLogger(log.Logger)
	// 	// config.Logger = logger.Default.LogMode(logger.Silent)
	// }

	// Get a DB concurrent connection
	// db, err := gorm.Open(sqlite.Open(dsn), config)
	// if err != nil {
	// 	return err
	// }

	// Setup the default DB connection
	//
	// Note: busy_timeout needs to be set BEFORE journal_mode
	_, err = db.Exec(`
		PRAGMA busy_timeout       = 10000;
		PRAGMA journal_mode       = WAL;
		PRAGMA journal_size_limit = 200000000;
		PRAGMA synchronous        = NORMAL;
		PRAGMA foreign_keys       = ON;
		PRAGMA temp_store         = MEMORY;
		PRAGMA cache_size         = -16000;
	`)

	if err != nil {
		return err
	}

	s.db = db

	return nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// DB returns the DB connection
func (s *SqliteDb) DB() *bun.DB {
	return s.db
}
