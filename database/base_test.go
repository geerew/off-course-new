package database

import (
	"os"
	"testing"

	"github.com/geerew/off-course/utils/appFs"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_Bootstrap(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		appFs := appFs.NewAppFs(afero.NewMemMapFs())

		db := NewSqliteDB(&SqliteDbConfig{
			IsDebug: false,
			DataDir: "./oc_data",
			AppFs:   appFs,
		})

		// Force DB to be in-memory
		os.Setenv("OC_InMemDb", "true")

		require.Nil(t, db.Bootstrap())

	})

	t.Run("error creating data dir", func(t *testing.T) {
		appFs := appFs.NewAppFs(afero.NewReadOnlyFs(afero.NewMemMapFs()))

		db := NewSqliteDB(&SqliteDbConfig{
			IsDebug: false,
			DataDir: "./oc_data",
			AppFs:   appFs,
		})

		// Force DB to be in-memory
		os.Setenv("OC_InMemDb", "true")

		err := db.Bootstrap()

		require.NotNil(t, err)
		assert.EqualError(t, err, "operation not permitted")
	})

	t.Run("error opening db", func(t *testing.T) {
		appFs := appFs.NewAppFs(afero.NewMemMapFs())

		db := NewSqliteDB(&SqliteDbConfig{
			IsDebug: false,
			DataDir: string([]byte{0x7f}), // using an invalid path
			AppFs:   appFs,
		})

		// Force DB to be in-memory
		os.Setenv("OC_InMemDb", "true")

		require.Nil(t, db.Bootstrap())
	})
}
