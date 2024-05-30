package daos

import (
	"sync"
	"testing"

	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/utils/appFs"
	"github.com/geerew/off-course/utils/logger"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func setup(t *testing.T) *database.DatabaseManager {
	// Logger
	var logs []*logger.Log
	var logsMux sync.Mutex
	loggy, err := logger.InitLogger(logger.TestWriteFn(&logs, &logsMux), 1)
	require.NoError(t, err, "Failed to initialize logger")

	// Filesystem
	appFs := appFs.NewAppFs(afero.NewMemMapFs(), loggy)

	// DB
	dbManager, err := database.NewSqliteDBManager(&database.DatabaseConfig{
		IsDebug:  false,
		DataDir:  "./oc_data",
		AppFs:    appFs,
		InMemory: true,
	})

	require.Nil(t, err)
	require.NotNil(t, dbManager)

	return dbManager
}
