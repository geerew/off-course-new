package models

import (
	"os"
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

func setup(t *testing.T) (*appFs.AppFs, database.Database, func(t *testing.T)) {
	loggerHook := zltest.New(t)
	log.Logger = zerolog.New(loggerHook).Level(zerolog.DebugLevel)

	appFs := appFs.NewAppFs(afero.NewMemMapFs())

	db := database.NewSqliteDB(&database.SqliteDbConfig{
		IsDebug: false,
		DataDir: "./oc_data",
		AppFs:   appFs,
	})

	// Force DB to be in-memory
	os.Setenv("OC_InMemDb", "true")

	require.Nil(t, db.Bootstrap())

	// teardown
	return appFs, db, func(t *testing.T) {
		os.Unsetenv("OC_InMemDb")
	}
}
