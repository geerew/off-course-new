package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/geerew/off-course/daos"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/utils/appFs"
	"github.com/geerew/off-course/utils/jobs"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestScans_GetScan(t *testing.T) {
	t.Run("200 (found)", func(t *testing.T) {
		appFs, db, cs, _ := setup(t)

		workingData := daos.NewTestData(t, db, 5, true, 0, 0)

		req := httptest.NewRequest(http.MethodGet, "/api/scans/"+workingData[2].ID, nil)
		status, body, err := scansRequestHelper(t, appFs, db, cs, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		var respData scanResponse
		err = json.Unmarshal(body, &respData)
		require.Nil(t, err)
		assert.Equal(t, workingData[2].Scan.ID, respData.ID)
		assert.Equal(t, workingData[2].Scan.CourseID, respData.CourseID)
		assert.Equal(t, workingData[2].Scan.Status, respData.Status)
	})

	t.Run("404 (not found)", func(t *testing.T) {
		appFs, db, cs, _ := setup(t)

		req := httptest.NewRequest(http.MethodGet, "/api/scans/test", nil)
		status, _, err := scansRequestHelper(t, appFs, db, cs, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusNotFound, status)
	})

	t.Run("500 (internal error)", func(t *testing.T) {
		appFs, db, cs, _ := setup(t)

		f := fiber.New()
		bindScansApi(f.Group("/api"), appFs, db, cs)

		_, err := db.Exec("DROP TABLE IF EXISTS " + daos.TableScans())
		require.Nil(t, err)

		req := httptest.NewRequest(http.MethodGet, "/api/scans/test", nil)
		status, _, err := scansRequestHelper(t, appFs, db, cs, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusInternalServerError, status)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestScans_CreateScan(t *testing.T) {
	t.Run("201 (created)", func(t *testing.T) {
		appFs, db, cs, _ := setup(t)

		workingData := daos.NewTestData(t, db, 1, false, 0, 0)

		req := httptest.NewRequest(http.MethodPost, "/api/scans/", strings.NewReader(fmt.Sprintf(`{"courseID": "%s"}`, workingData[0].ID)))
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

		status, body, err := scansRequestHelper(t, appFs, db, cs, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusCreated, status)

		var respData scanResponse
		err = json.Unmarshal(body, &respData)
		require.Nil(t, err)
		assert.Equal(t, workingData[0].ID, respData.CourseID)
	})

	t.Run("400 (bind error)", func(t *testing.T) {
		appFs, db, cs, _ := setup(t)

		req := httptest.NewRequest(http.MethodPost, "/api/scans/", strings.NewReader(`{`))
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

		status, body, err := scansRequestHelper(t, appFs, db, cs, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, status)
		assert.Contains(t, string(body), "error parsing data")
	})

	t.Run("400 (invalid data)", func(t *testing.T) {
		appFs, db, cs, _ := setup(t)

		req := httptest.NewRequest(http.MethodPost, "/api/scans/", strings.NewReader(`{"courseID": ""}`))
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

		status, body, err := scansRequestHelper(t, appFs, db, cs, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, status)
		assert.Contains(t, string(body), "a course ID is required")
	})

	t.Run("400 (invalid course id)", func(t *testing.T) {
		appFs, db, cs, _ := setup(t)

		req := httptest.NewRequest(http.MethodPost, "/api/scans/", strings.NewReader(`{"courseID": "test"}`))
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

		status, body, err := scansRequestHelper(t, appFs, db, cs, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, status)
		assert.Contains(t, string(body), "invalid course ID")
	})

	t.Run("500 (internal error)", func(t *testing.T) {
		appFs, db, cs, _ := setup(t)

		_, err := db.Exec("DROP TABLE IF EXISTS " + daos.TableScans())
		require.Nil(t, err)

		req := httptest.NewRequest(http.MethodPost, "/api/scans/", strings.NewReader(`{"courseID": "test"}`))
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

		status, body, err := scansRequestHelper(t, appFs, db, cs, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusInternalServerError, status)
		assert.Contains(t, string(body), "error creating scan job")
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// HELPERS
// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func scansRequestHelper(t *testing.T, appFs *appFs.AppFs, db database.Database, cs *jobs.CourseScanner, req *http.Request) (int, []byte, error) {
	f := fiber.New()
	bindScansApi(f.Group("/api"), appFs, db, cs)

	resp, err := f.Test(req)
	if err != nil {
		return -1, nil, err
	}

	b, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, b, err
}
