package api

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/utils/appFs"
	"github.com/geerew/off-course/utils/jobs"
	"github.com/geerew/off-course/utils/logger"
	"github.com/geerew/off-course/utils/pagination"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
func setup(t *testing.T) *Router {
	appFs := appFs.NewAppFs(afero.NewMemMapFs())

	dbManager, err := database.NewDBManager(&database.DatabaseConfig{
		IsDebug:  false,
		DataDir:  "./oc_data",
		AppFs:    appFs,
		InMemory: true,
	})

	require.Nil(t, err)
	require.NotNil(t, dbManager)

	courseScanner := jobs.NewCourseScanner(&jobs.CourseScannerConfig{
		Db:    dbManager.DataDb,
		AppFs: appFs,
	})

	loggy, err := logger.InitLogger(logger.BasicWriteFn())
	require.NoError(t, err, "Failed to initialize logger")

	config := &RouterConfig{
		DbManager:     dbManager,
		AppFs:         appFs,
		CourseScanner: courseScanner,
		Logger:        loggy,
	}

	router := New(config)

	// teardown
	return router
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func requestHelper(router *Router, req *http.Request) (int, []byte, error) {
	resp, err := router.router.Test(req)
	if err != nil {
		return -1, nil, err
	}

	body, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, body, err
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func unmarshalHelper[T any](t *testing.T, body []byte) (pagination.PaginationResult, []T) {
	var respData pagination.PaginationResult
	err := json.Unmarshal(body, &respData)
	require.Nil(t, err)

	var resp []T
	for _, item := range respData.Items {
		var r T
		require.Nil(t, json.Unmarshal(item, &r))
		resp = append(resp, r)
	}

	return respData, resp
}
