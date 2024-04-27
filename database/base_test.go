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

		db := NewSqliteDB(&SqliteDbConfig{
			IsDebug:  false,
			DataDir:  "./oc_data",
			AppFs:    appFs,
			InMemory: true,
		})

		require.Nil(t, db.Bootstrap())

	})

	t.Run("error creating data dir", func(t *testing.T) {
		appFs := appFs.NewAppFs(afero.NewReadOnlyFs(afero.NewMemMapFs()))

		db := NewSqliteDB(&SqliteDbConfig{
			IsDebug:  false,
			DataDir:  "./oc_data",
			AppFs:    appFs,
			InMemory: true,
		})

		err := db.Bootstrap()

		require.NotNil(t, err)
		require.EqualError(t, err, "operation not permitted")
	})

	t.Run("error opening db", func(t *testing.T) {
		appFs := appFs.NewAppFs(afero.NewMemMapFs())

		db := NewSqliteDB(&SqliteDbConfig{
			IsDebug:  false,
			DataDir:  string([]byte{0x7f}), // using an invalid path
			AppFs:    appFs,
			InMemory: true,
		})

		require.Nil(t, db.Bootstrap())
	})
}
