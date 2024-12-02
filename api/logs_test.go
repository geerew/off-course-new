package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils/pagination"
	"github.com/geerew/off-course/utils/types"
	"github.com/stretchr/testify/require"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestLogs_GetLogs(t *testing.T) {
	t.Run("200 (empty)", func(t *testing.T) {
		router, _ := setup(t)

		status, body, err := requestHelper(t, router, httptest.NewRequest(http.MethodGet, "/api/logs/", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, _ := unmarshalHelper[logResponse](t, body)
		require.Zero(t, int(paginationResp.TotalItems))
		require.Zero(t, len(paginationResp.Items))
	})

	t.Run("200 (found)", func(t *testing.T) {
		router, ctx := setup(t)

		for i := range 5 {
			log := &models.Log{
				Data:    map[string]any{},
				Level:   0,
				Message: fmt.Sprintf("log %d", i+1),
			}

			require.Nil(t, router.logDao.WriteLog(ctx, log))
		}

		status, body, err := requestHelper(t, router, httptest.NewRequest(http.MethodGet, "/api/logs/", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, logResponses := unmarshalHelper[logResponse](t, body)
		require.Equal(t, 5, int(paginationResp.TotalItems))
		require.Len(t, logResponses, 5)
	})

	t.Run("200 (min level)", func(t *testing.T) {
		router, ctx := setup(t)

		require.NoError(t, router.logDao.WriteLog(ctx, &models.Log{Data: map[string]any{}, Level: -4, Message: "debug log"}))
		time.Sleep(1 * time.Millisecond)

		require.NoError(t, router.logDao.WriteLog(ctx, &models.Log{Data: map[string]any{}, Level: 0, Message: "info log"}))
		time.Sleep(1 * time.Millisecond)

		require.NoError(t, router.logDao.WriteLog(ctx, &models.Log{Data: map[string]any{}, Level: 4, Message: "warn log"}))
		time.Sleep(1 * time.Millisecond)

		require.NoError(t, router.logDao.WriteLog(ctx, &models.Log{Data: map[string]any{}, Level: 8, Message: "error log"}))

		// All
		status, body, err := requestHelper(t, router, httptest.NewRequest(http.MethodGet, "/api/logs/", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, logResponses := unmarshalHelper[logResponse](t, body)
		require.Equal(t, 4, int(paginationResp.TotalItems))
		require.Len(t, logResponses, 4)

		// Debug
		status, body, err = requestHelper(t, router, httptest.NewRequest(http.MethodGet, "/api/logs/?levels=-4", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, logResponses = unmarshalHelper[logResponse](t, body)
		require.Equal(t, 1, int(paginationResp.TotalItems))
		require.Len(t, logResponses, 1)
		require.Equal(t, "debug log", logResponses[0].Message)

		// Debug and info
		status, body, err = requestHelper(t, router, httptest.NewRequest(http.MethodGet, "/api/logs/?levels=-4,0", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, logResponses = unmarshalHelper[logResponse](t, body)
		require.Equal(t, 2, int(paginationResp.TotalItems))
		require.Len(t, logResponses, 2)
		require.Equal(t, "debug log", logResponses[1].Message)
		require.Equal(t, "info log", logResponses[0].Message)

		// Warn and error only (with spaces)
		status, body, err = requestHelper(t, router, httptest.NewRequest(http.MethodGet, "/api/logs/?levels=4,%20%208", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, logResponses = unmarshalHelper[logResponse](t, body)
		require.Equal(t, 2, int(paginationResp.TotalItems))
		require.Len(t, logResponses, 2)
		require.Equal(t, "error log", logResponses[0].Message)
		require.Equal(t, "warn log", logResponses[1].Message)
	})

	t.Run("200 (types)", func(t *testing.T) {
		router, ctx := setup(t)

		require.NoError(t, router.logDao.WriteLog(ctx, &models.Log{Data: map[string]any{"type": types.LogTypeRequest.String()}, Level: -4, Message: "log 1"}))
		time.Sleep(1 * time.Millisecond)

		require.NoError(t, router.logDao.WriteLog(ctx, &models.Log{Data: map[string]any{"type": types.LogTypeRequest.String()}, Level: 0, Message: "log 2"}))
		time.Sleep(1 * time.Millisecond)

		require.NoError(t, router.logDao.WriteLog(ctx, &models.Log{Data: map[string]any{"type": types.LogTypeDB.String()}, Level: 4, Message: "log 3"}))
		time.Sleep(1 * time.Millisecond)

		require.NoError(t, router.logDao.WriteLog(ctx, &models.Log{Data: map[string]any{"type": types.LogTypeCourseScan.String()}, Level: 8, Message: "log 4"}))

		// Request
		status, body, err := requestHelper(t, router, httptest.NewRequest(http.MethodGet, "/api/logs/?types="+types.LogTypeRequest.String(), nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, logResponses := unmarshalHelper[logResponse](t, body)
		require.Equal(t, 2, int(paginationResp.TotalItems))
		require.Len(t, logResponses, 2)
		require.Equal(t, "log 2", logResponses[0].Message)
		require.Equal(t, "log 1", logResponses[1].Message)

		// Database and course scan
		typesQuery := url.QueryEscape(types.LogTypeDB.String() + ",   " + types.LogTypeCourseScan.String())
		status, body, err = requestHelper(t, router, httptest.NewRequest(http.MethodGet, "/api/logs/?types="+typesQuery, nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, logResponses = unmarshalHelper[logResponse](t, body)
		require.Equal(t, 2, int(paginationResp.TotalItems))
		require.Len(t, logResponses, 2)
		require.Equal(t, "log 4", logResponses[0].Message)
		require.Equal(t, "log 3", logResponses[1].Message)

		// Database
		status, body, err = requestHelper(t, router, httptest.NewRequest(http.MethodGet, "/api/logs/?types="+types.LogTypeDB.String(), nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, logResponses = unmarshalHelper[logResponse](t, body)
		require.Equal(t, 1, int(paginationResp.TotalItems))
		require.Len(t, logResponses, 1)
		require.Equal(t, "log 3", logResponses[0].Message)
	})

	t.Run("200 (messages)", func(t *testing.T) {
		router, ctx := setup(t)

		require.NoError(t, router.logDao.WriteLog(ctx, &models.Log{Data: map[string]any{"type": types.LogTypeRequest.String()}, Level: -4, Message: "log 1"}))
		time.Sleep(1 * time.Millisecond)

		require.NoError(t, router.logDao.WriteLog(ctx, &models.Log{Data: map[string]any{"type": types.LogTypeRequest.String()}, Level: 0, Message: "log 2"}))
		time.Sleep(1 * time.Millisecond)

		require.NoError(t, router.logDao.WriteLog(ctx, &models.Log{Data: map[string]any{"type": types.LogTypeDB.String()}, Level: 4, Message: "log 3"}))
		time.Sleep(1 * time.Millisecond)

		require.NoError(t, router.logDao.WriteLog(ctx, &models.Log{Data: map[string]any{"type": types.LogTypeCourseScan.String()}, Level: 8, Message: "log 4"}))

		// log 1
		status, body, err := requestHelper(t, router, httptest.NewRequest(http.MethodGet, "/api/logs/?messages=log%201", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, logResponses := unmarshalHelper[logResponse](t, body)
		require.Equal(t, 1, int(paginationResp.TotalItems))
		require.Len(t, logResponses, 1)
		require.Equal(t, "log 1", logResponses[0].Message)

		// log 2 and log 4
		status, body, err = requestHelper(t, router, httptest.NewRequest(http.MethodGet, "/api/logs/?messages=log%202,log%204", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, logResponses = unmarshalHelper[logResponse](t, body)
		require.Equal(t, 2, int(paginationResp.TotalItems))
		require.Len(t, logResponses, 2)
		require.Equal(t, "log 4", logResponses[0].Message)
		require.Equal(t, "log 2", logResponses[1].Message)
	})

	t.Run("200 (pagination)", func(t *testing.T) {
		router, ctx := setup(t)

		for i := range 17 {
			require.Nil(t, router.logDao.WriteLog(ctx, &models.Log{Data: map[string]any{}, Level: 0, Message: fmt.Sprintf("log %d", i+1)}))
			time.Sleep(1 * time.Millisecond)
		}

		// Get the first page (10 logs)
		params := url.Values{
			pagination.PageQueryParam:    {"1"},
			pagination.PerPageQueryParam: {"10"},
		}

		status, body, err := requestHelper(t, router, httptest.NewRequest(http.MethodGet, "/api/logs/?"+params.Encode(), nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, logResponses := unmarshalHelper[logResponse](t, body)
		require.Equal(t, 17, int(paginationResp.TotalItems))
		require.Len(t, paginationResp.Items, 10)
		require.Equal(t, "log 17", logResponses[0].Message)
		require.Equal(t, "log 8", logResponses[9].Message)

		// Get the second page (7 logs)
		params = url.Values{
			pagination.PageQueryParam:    {"2"},
			pagination.PerPageQueryParam: {"10"},
		}
		status, body, err = requestHelper(t, router, httptest.NewRequest(http.MethodGet, "/api/logs/?"+params.Encode(), nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, logResponses = unmarshalHelper[logResponse](t, body)
		require.Equal(t, 17, int(paginationResp.TotalItems))
		require.Len(t, paginationResp.Items, 7)
		require.Equal(t, "log 7", logResponses[0].Message)
		require.Equal(t, "log 1", logResponses[6].Message)
	})

	t.Run("500 (internal error)", func(t *testing.T) {
		router, _ := setup(t)

		// Drop the courses table
		_, err := router.config.DbManager.LogsDb.Exec("DROP TABLE IF EXISTS " + models.LOG_TABLE)
		require.NoError(t, err)

		status, _, err := requestHelper(t, router, httptest.NewRequest(http.MethodGet, "/api/logs/", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusInternalServerError, status)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestLogs_GetLogTypes(t *testing.T) {
	router, _ := setup(t)

	status, body, err := requestHelper(t, router, httptest.NewRequest(http.MethodGet, "/api/logs/types", nil))
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, status)

	var typesResp []string
	err = json.Unmarshal(body, &typesResp)
	require.NoError(t, err)

	require.Equal(t, typesResp, types.AllLogTypes())
}
