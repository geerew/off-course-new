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
		loggy, err := logger.InitLogger(logger.NilWriteFn(), 1)
		require.Nil(t, err)

		appFs := appFs.NewAppFs(afero.NewMemMapFs(), loggy)

		dbManager, err := NewSqliteDBManager(&DatabaseConfig{
			IsDebug:  false,
			DataDir:  "./oc_data",
			AppFs:    appFs,
			InMemory: true,
		})

		require.Nil(t, err)
		require.NotNil(t, dbManager)

	})

	t.Run("error creating data dir", func(t *testing.T) {
		loggy, err := logger.InitLogger(logger.NilWriteFn(), 1)
		require.Nil(t, err)

		appFs := appFs.NewAppFs(afero.NewReadOnlyFs(afero.NewMemMapFs()), loggy)

		dbManager, err := NewSqliteDBManager(&DatabaseConfig{
			IsDebug:  false,
			DataDir:  "./oc_data",
			AppFs:    appFs,
			InMemory: true,
		})

		require.NotNil(t, err)
		require.EqualError(t, err, "operation not permitted")
		require.Nil(t, dbManager)
	})
}
