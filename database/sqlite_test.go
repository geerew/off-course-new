package database

import (
	"context"
	"fmt"
	"testing"

	"github.com/geerew/off-course/utils/appFs"
	"github.com/geerew/off-course/utils/logger"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func setupSqliteDB(t *testing.T) *DatabaseManager {
	t.Helper()

	// Logger
	logger, _, err := logger.InitLogger(&logger.BatchOptions{
		BatchSize: 1,
		WriteFn:   logger.NilWriteFn(),
	})
	require.NoError(t, err, "Failed to initialize logger")

	// Filesystem
	appFs := appFs.NewAppFs(afero.NewMemMapFs(), logger)

	// DB
	dbManager, err := NewSqliteDBManager(&DatabaseConfig{
		DataDir:  "./oc_data",
		AppFs:    appFs,
		InMemory: true,
	})

	require.NoError(t, err)
	require.NotNil(t, dbManager)

	// Test table
	_, err = dbManager.DataDb.Exec("CREATE TABLE test (id INTEGER PRIMARY KEY, name TEXT)")
	require.NoError(t, err)

	return dbManager
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestSqliteDb_Bootstrap(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		logger, _, err := logger.InitLogger(&logger.BatchOptions{
			BatchSize: 1,
			WriteFn:   logger.NilWriteFn(),
		})
		require.NoError(t, err)

		appFs := appFs.NewAppFs(afero.NewMemMapFs(), logger)

		db, err := NewSqliteDB(&DatabaseConfig{
			DataDir:    "./oc_data",
			DSN:        "data.db",
			MigrateDir: "data",
			AppFs:      appFs,
			InMemory:   true,
		})

		require.NoError(t, err)
		require.NotNil(t, db)

	})

	t.Run("error creating data dir", func(t *testing.T) {
		logger, _, err := logger.InitLogger(&logger.BatchOptions{
			BatchSize: 1,
			WriteFn:   logger.NilWriteFn(),
		})
		require.NoError(t, err)

		appFs := appFs.NewAppFs(afero.NewReadOnlyFs(afero.NewMemMapFs()), logger)

		db, err := NewSqliteDB(&DatabaseConfig{
			DataDir:    "./oc_data",
			DSN:        "data.db",
			MigrateDir: "data",
			AppFs:      appFs,
			InMemory:   true,
		})

		require.NotNil(t, err)
		require.EqualError(t, err, "operation not permitted")
		require.Nil(t, db)
	})

	t.Run("invalid migration", func(t *testing.T) {
		logger, _, err := logger.InitLogger(&logger.BatchOptions{
			BatchSize: 1,
			WriteFn:   logger.NilWriteFn(),
		})
		require.NoError(t, err)

		appFs := appFs.NewAppFs(afero.NewMemMapFs(), logger)

		db, err := NewSqliteDB(&DatabaseConfig{
			DataDir:    "./oc_data",
			DSN:        "data.db",
			MigrateDir: "test",
			AppFs:      appFs,
			InMemory:   true,
		})

		require.NotNil(t, err)
		require.EqualError(t, err, "test directory does not exist")
		require.Nil(t, db)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestSqliteDb_Query(t *testing.T) {
	dbManager := setupSqliteDB(t)

	_, err := dbManager.DataDb.Exec("INSERT INTO test (name) VALUES ('test')")
	require.NoError(t, err)

	rows, err := dbManager.DataDb.Query("SELECT * FROM test")
	require.NoError(t, err)
	defer rows.Close()

	for rows.Next() {
		var id int
		var name string
		err = rows.Scan(&id, &name)
		require.NoError(t, err)
		require.Equal(t, 1, id)
		require.Equal(t, "test", name)
	}

	require.Nil(t, rows.Err())
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestSqliteDb_QueryRow(t *testing.T) {
	dbManager := setupSqliteDB(t)

	_, err := dbManager.DataDb.Exec("INSERT INTO test (name) VALUES ('test')")
	require.NoError(t, err)

	var id int
	var name string
	err = dbManager.DataDb.QueryRow("SELECT * FROM test").Scan(&id, &name)

	require.NoError(t, err)
	require.Equal(t, "test", name)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestSqliteDb_Exec(t *testing.T) {
	dbManager := setupSqliteDB(t)

	result, err := dbManager.DataDb.Exec("INSERT INTO test (name) VALUES ('test')")
	require.NoError(t, err)

	rowAffected, err := result.RowsAffected()
	require.NoError(t, err)
	require.Equal(t, int64(1), rowAffected)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestSqliteDb_RunInTransaction(t *testing.T) {

	t.Run(("error"), func(t *testing.T) {
		dbManager := setupSqliteDB(t)

		ctx := context.Background()
		err := dbManager.DataDb.RunInTransaction(ctx, func(txCtx context.Context) error {
			q := QuerierFromContext(txCtx, dbManager.DataDb)
			_, err := q.Exec("INSERT INTO test (name) VALUES ('test')")
			if err != nil {
				return err
			}

			// Simulate error
			if true {
				return fmt.Errorf("error")
			}

			return nil
		})

		require.Error(t, err)

		var count int
		err = dbManager.DataDb.QueryRow("SELECT COUNT(*) FROM test").Scan(&count)
		require.NoError(t, err)
		require.Equal(t, 0, count)
	})

	t.Run(("success"), func(t *testing.T) {
		dbManager := setupSqliteDB(t)

		ctx := context.Background()
		err := dbManager.DataDb.RunInTransaction(ctx, func(txCtx context.Context) error {
			q := QuerierFromContext(txCtx, dbManager.DataDb)
			_, err := q.Exec("INSERT INTO test (name) VALUES ('test')")
			if err != nil {
				return err
			}

			return nil
		})

		require.NoError(t, err)

		var count int
		err = dbManager.DataDb.QueryRow("SELECT COUNT(*) FROM test").Scan(&count)
		require.NoError(t, err)
		require.Equal(t, 1, count)
	})
}
