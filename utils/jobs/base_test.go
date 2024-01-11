package jobs

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

// SetupCourseScanner sets up a course scanner for testing
func setupCourseScanner(t *testing.T) (*CourseScanner, *zltest.Tester, func(t *testing.T)) {
	loggerHook := zltest.New(t)
	log.Logger = zerolog.New(loggerHook).Level(zerolog.DebugLevel)

	appFs := appFs.NewAppFs(afero.NewMemMapFs())

	db := database.NewSqliteDB(&database.SqliteDbConfig{
		IsDebug: false,
		DataDir: "./co_data",
		AppFs:   appFs,
	})

	// Force DB to be in-memory
	os.Setenv("OC_InMemDb", "true")

	require.Nil(t, db.Bootstrap())

	courseScanner := NewCourseScanner(&CourseScannerConfig{
		Db:    db,
		AppFs: appFs,
	})

	// teardown
	return courseScanner, loggerHook, func(t *testing.T) {
		os.Unsetenv("OC_InMemDb")
	}
}
