package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/geerew/off-course/daos"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestScans_GetScan(t *testing.T) {
	t.Run("200 (found)", func(t *testing.T) {
		router := setup(t)

		testData := daos.NewTestBuilder(t).Db(router.config.DbManager.DataDb).Courses(5).Scan().Build()

		req := httptest.NewRequest(http.MethodGet, "/api/scans/"+testData[2].ID, nil)
		status, body, err := requestHelper(t, router, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		var respData scanResponse
		err = json.Unmarshal(body, &respData)
		require.Nil(t, err)
		require.Equal(t, testData[2].Scan.ID, respData.ID)
		require.Equal(t, testData[2].Scan.CourseID, respData.CourseID)
		require.Equal(t, testData[2].Scan.Status, respData.Status)
	})

	t.Run("404 (not found)", func(t *testing.T) {
		router := setup(t)

		req := httptest.NewRequest(http.MethodGet, "/api/scans/test", nil)
		status, _, err := requestHelper(t, router, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusNotFound, status)
	})

	t.Run("500 (internal error)", func(t *testing.T) {
		router := setup(t)

		_, err := router.config.DbManager.DataDb.Exec("DROP TABLE IF EXISTS " + daos.NewScanDao(router.config.DbManager.DataDb).Table())
		require.Nil(t, err)

		req := httptest.NewRequest(http.MethodGet, "/api/scans/test", nil)
		status, _, err := requestHelper(t, router, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusInternalServerError, status)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestScans_CreateScan(t *testing.T) {
	t.Run("201 (created)", func(t *testing.T) {
		router := setup(t)

		testData := daos.NewTestBuilder(t).Db(router.config.DbManager.DataDb).Courses(1).Build()

		req := httptest.NewRequest(http.MethodPost, "/api/scans/", strings.NewReader(fmt.Sprintf(`{"courseID": "%s"}`, testData[0].ID)))
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

		status, body, err := requestHelper(t, router, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusCreated, status)

		var respData scanResponse
		err = json.Unmarshal(body, &respData)
		require.Nil(t, err)
		require.Equal(t, testData[0].ID, respData.CourseID)
	})

	t.Run("400 (bind error)", func(t *testing.T) {
		router := setup(t)

		req := httptest.NewRequest(http.MethodPost, "/api/scans/", strings.NewReader(`{`))
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

		status, body, err := requestHelper(t, router, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, status)
		require.Contains(t, string(body), "Error parsing data")
	})

	t.Run("400 (invalid data)", func(t *testing.T) {
		router := setup(t)

		req := httptest.NewRequest(http.MethodPost, "/api/scans/", strings.NewReader(`{"courseID": ""}`))
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

		status, body, err := requestHelper(t, router, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, status)
		require.Contains(t, string(body), "A course ID is required")
	})

	t.Run("400 (invalid course id)", func(t *testing.T) {
		router := setup(t)

		req := httptest.NewRequest(http.MethodPost, "/api/scans/", strings.NewReader(`{"courseID": "test"}`))
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

		status, body, err := requestHelper(t, router, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, status)
		require.Contains(t, string(body), "Invalid course ID")
	})

	t.Run("500 (internal error)", func(t *testing.T) {
		router := setup(t)

		_, err := router.config.DbManager.DataDb.Exec("DROP TABLE IF EXISTS " + daos.NewScanDao(router.config.DbManager.DataDb).Table())
		require.Nil(t, err)

		req := httptest.NewRequest(http.MethodPost, "/api/scans/", strings.NewReader(`{"courseID": "test"}`))
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

		status, body, err := requestHelper(t, router, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusInternalServerError, status)
		require.Contains(t, string(body), "Error creating scan job")
	})
}
