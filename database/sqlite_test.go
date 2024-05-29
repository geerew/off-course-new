package database

import (
	"testing"

	"github.com/geerew/off-course/utils/appFs"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

func Test_Bootstrap(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		appFs := appFs.NewAppFs(afero.NewMemMapFs())

		db, err := NewSqliteDB(&DatabaseConfig{
			IsDebug:    false,
			DataDir:    "./oc_data",
			DSN:        "data.db",
			MigrateDir: "data",
			AppFs:      appFs,
			InMemory:   true,
		})

		require.Nil(t, err)
		require.NotNil(t, db)

	})

	t.Run("error creating data dir", func(t *testing.T) {
		appFs := appFs.NewAppFs(afero.NewReadOnlyFs(afero.NewMemMapFs()))

		db, err := NewSqliteDB(&DatabaseConfig{
			IsDebug:    false,
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
		appFs := appFs.NewAppFs(afero.NewMemMapFs())

		db, err := NewSqliteDB(&DatabaseConfig{
			IsDebug:    false,
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
