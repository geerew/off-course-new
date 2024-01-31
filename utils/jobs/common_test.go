package jobs

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

// SetupCourseScanner sets up a course scanner for testing
func setupCourseScanner(t *testing.T) (*CourseScanner, database.Database, *zltest.Tester) {
	loggerHook := zltest.New(t)
	log.Logger = zerolog.New(loggerHook).Level(zerolog.DebugLevel)

	appFs := appFs.NewAppFs(afero.NewMemMapFs())

	db := database.NewSqliteDB(&database.SqliteDbConfig{
		IsDebug:  false,
		DataDir:  "./oc_data",
		AppFs:    appFs,
		InMemory: true,
	})

	require.NoError(t, db.Bootstrap(), "Failed to bootstrap database")

	courseScanner := NewCourseScanner(&CourseScannerConfig{
		Db:    db,
		AppFs: appFs,
	})

	return courseScanner, db, loggerHook
}
