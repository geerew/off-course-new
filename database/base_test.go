package database

import (
	"testing"

	"github.com/geerew/off-course/utils/appFs"
	"github.com/geerew/off-course/utils/logger"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_NewSqliteDBManager(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		logger, _, err := logger.InitLogger(&logger.BatchOptions{
			BatchSize: 1,
			WriteFn:   logger.NilWriteFn(),
		})
		require.NoError(t, err)

		appFs := appFs.NewAppFs(afero.NewMemMapFs(), logger)

		dbManager, err := NewSqliteDBManager(&DatabaseConfig{
			DataDir:  "./oc_data",
			AppFs:    appFs,
			InMemory: true,
		})

		require.NoError(t, err)
		require.NotNil(t, dbManager)

	})

	t.Run("error creating data dir", func(t *testing.T) {
		logger, _, err := logger.InitLogger(&logger.BatchOptions{
			BatchSize: 1,
			WriteFn:   logger.NilWriteFn(),
		})
		require.NoError(t, err)

		appFs := appFs.NewAppFs(afero.NewReadOnlyFs(afero.NewMemMapFs()), logger)

		dbManager, err := NewSqliteDBManager(&DatabaseConfig{
			DataDir:  "./oc_data",
			AppFs:    appFs,
			InMemory: true,
		})

		require.NotNil(t, err)
		require.EqualError(t, err, "operation not permitted")
		require.Nil(t, dbManager)
	})
}
