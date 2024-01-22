package api

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/utils/appFs"
	"github.com/geerew/off-course/utils/jobs"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rzajac/zltest"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_RouterLogger(t *testing.T) {

	appFs, db, courseScanner, lh, teardown := setup(t)
	defer teardown(t)

	router := New(&RouterConfig{
		Db:            db,
		AppFs:         appFs,
		CourseScanner: courseScanner,
		Port:          ":1234",
		IsProduction:  false,
	})

	_, err := router.router.Test(httptest.NewRequest(http.MethodGet, "/", nil))
	assert.NoError(t, err)

	lh.LastEntry().ExpMsg("Request completed")
	lh.LastEntry().ExpStr("path", "/")
	lh.LastEntry().ExpStr("method", "GET")
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// HELPERS
// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func setup(t *testing.T) (*appFs.AppFs, database.Database, *jobs.CourseScanner, *zltest.Tester, func(t *testing.T)) {
	// Set test logger
	loggerHook := zltest.New(t)
	log.Logger = zerolog.New(loggerHook)

	appFs := appFs.NewAppFs(afero.NewMemMapFs())

	db := database.NewSqliteDB(&database.SqliteDbConfig{
		IsDebug: false,
		DataDir: "./oc_data",
		AppFs:   appFs,
	})

	// Force DB to be in-memory
	os.Setenv("OC_InMemDb", "true")

	require.Nil(t, db.Bootstrap())

	courseScanner := jobs.NewCourseScanner(&jobs.CourseScannerConfig{
		Db:    db,
		AppFs: appFs,
	})

	// teardown
	return appFs, db, courseScanner, loggerHook, func(t *testing.T) {
		// store.Close()
		os.Unsetenv("OC_InMemDb")
	}
}
