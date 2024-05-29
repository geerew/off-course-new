package daos

import (
	"testing"

	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/utils/appFs"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rzajac/zltest"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func setup(t *testing.T) *database.DatabaseManager {
	loggerHook := zltest.New(t)
	log.Logger = zerolog.New(loggerHook).Level(zerolog.DebugLevel)

	appFs := appFs.NewAppFs(afero.NewMemMapFs())

	dbManager, err := database.NewDBManager(&database.DatabaseConfig{
		IsDebug:  false,
		DataDir:  "./oc_data",
		AppFs:    appFs,
		InMemory: true,
	})

	require.Nil(t, err)
	require.NotNil(t, dbManager)

	return dbManager
}
